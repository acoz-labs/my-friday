package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

var ErrCriticalPolicyLoss = errors.New("critical policy loss")

type UsageMetric struct {
	Value         *int64 `json:"value"`
	Provenance    string `json:"provenance"`
	Complete      bool   `json:"complete"`
	MissingReason string `json:"missing_reason"`
}

type Telemetry struct {
	RootInputCumulative   UsageMetric `json:"root_input_cumulative"`
	WorkerInputCumulative UsageMetric `json:"worker_input_cumulative"`
	TotalInput            UsageMetric `json:"total_input"`
	TotalOutput           UsageMetric `json:"total_output"`
	CachedInput           UsageMetric `json:"cached_input"`
	PeakRootRequestInput  UsageMetric `json:"peak_root_request_input"`
	ActualWindowOccupancy UsageMetric `json:"actual_window_occupancy"`
	ToolCalls             int         `json:"tool_calls"`
	WorkerStarts          int         `json:"worker_starts"`
	WorkerReturns         int         `json:"worker_returns"`
	PolicyLoss            bool        `json:"policy_loss"`
}

type telemetryEvent struct {
	EventID                string `json:"event_id"`
	Type                   string `json:"type"`
	AgentID                string `json:"agent_id,omitempty"`
	Role                   string `json:"role,omitempty"`
	Aggregate              bool   `json:"aggregate,omitempty"`
	InputTokens            *int64 `json:"input_tokens,omitempty"`
	OutputTokens           *int64 `json:"output_tokens,omitempty"`
	CachedInputTokens      *int64 `json:"cached_input_tokens,omitempty"`
	RootRequestInputTokens *int64 `json:"root_request_input_tokens,omitempty"`
	WindowOccupancyTokens  *int64 `json:"window_occupancy_tokens,omitempty"`
	WorkerDepth            *int   `json:"worker_depth,omitempty"`
}

func ParseTelemetry(reader io.Reader) (Telemetry, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	seen := map[string]bool{}
	var aggregate []telemetryEvent
	rootInput, workerInput, output, cached := int64(0), int64(0), int64(0), int64(0)
	var peak *int64
	var occupancy *int64
	telemetry := Telemetry{}
	rootUsage, workerUsage := 0, 0
	allCacheReported, allRootRequestReported, allWindowReported := true, true, true
	for scanner.Scan() {
		var event telemetryEvent
		if err := decodeStrictJSON("telemetry event", bytes.Clone(scanner.Bytes()), &event); err != nil {
			return Telemetry{}, fmt.Errorf("malformed telemetry: %w", err)
		}
		if event.EventID == "" || seen[event.EventID] {
			return Telemetry{}, fmt.Errorf("missing or duplicate telemetry event id %q", event.EventID)
		}
		seen[event.EventID] = true
		switch event.Type {
		case "usage":
			if event.AgentID == "" || (event.Role != "root" && event.Role != "worker") || event.InputTokens == nil || event.OutputTokens == nil {
				return Telemetry{}, errors.New("usage event lacks agent role or token values")
			}
			if *event.InputTokens < 0 || *event.OutputTokens < 0 || (event.CachedInputTokens != nil && *event.CachedInputTokens < 0) || (event.RootRequestInputTokens != nil && *event.RootRequestInputTokens < 0) || (event.WindowOccupancyTokens != nil && *event.WindowOccupancyTokens < 0) {
				return Telemetry{}, errors.New("usage event contains a negative metric")
			}
			if event.Aggregate {
				aggregate = append(aggregate, event)
			} else {
				if event.Role == "root" {
					rootUsage++
					rootInput += *event.InputTokens
					if event.RootRequestInputTokens == nil {
						allRootRequestReported = false
					} else if peak == nil || *event.RootRequestInputTokens > *peak {
						value := *event.RootRequestInputTokens
						peak = &value
					}
					if event.WindowOccupancyTokens == nil {
						allWindowReported = false
					} else if occupancy == nil || *event.WindowOccupancyTokens > *occupancy {
						value := *event.WindowOccupancyTokens
						occupancy = &value
					}
				} else {
					workerUsage++
					workerInput += *event.InputTokens
				}
				output += *event.OutputTokens
				if event.CachedInputTokens == nil {
					allCacheReported = false
				} else {
					cached += *event.CachedInputTokens
				}
			}
		case "tool_call":
			telemetry.ToolCalls++
			if telemetry.ToolCalls > 8 {
				return Telemetry{}, errors.New("observed tool-call ceiling exceeded")
			}
		case "worker_start":
			telemetry.WorkerStarts++
			if event.WorkerDepth == nil || *event.WorkerDepth != 1 || telemetry.WorkerStarts > 1 {
				return Telemetry{}, errors.New("observed worker count/depth ceiling exceeded")
			}
		case "worker_return":
			telemetry.WorkerReturns++
			if telemetry.WorkerReturns > telemetry.WorkerStarts {
				return Telemetry{}, errors.New("worker return lacks matching start")
			}
		case "policy_loss":
			telemetry.PolicyLoss = true
			return telemetry, ErrCriticalPolicyLoss
		case "result":
		default:
			return Telemetry{}, fmt.Errorf("unknown telemetry event type %q", event.Type)
		}
	}
	if err := scanner.Err(); err != nil {
		return Telemetry{}, err
	}
	if telemetry.WorkerReturns != telemetry.WorkerStarts {
		return Telemetry{}, errors.New("worker did not return")
	}
	if rootUsage > 0 {
		telemetry.RootInputCumulative = measuredMetric(rootInput, "sum of unique root usage events", true, "")
	} else {
		telemetry.RootInputCumulative = missingMetric("sum of unique root usage events", "no root usage event")
	}
	if telemetry.WorkerStarts == 0 && rootUsage > 0 {
		telemetry.WorkerInputCumulative = measuredMetric(0, "no worker launched", true, "")
	} else if workerUsage > 0 {
		telemetry.WorkerInputCumulative = measuredMetric(workerInput, "sum of unique worker usage events", true, "")
	} else {
		telemetry.WorkerInputCumulative = missingMetric("sum of unique worker usage events", "worker launched without child usage")
	}
	totalInput := rootInput + workerInput
	if len(aggregate) == 1 {
		totalInput = *aggregate[0].InputTokens
		output = *aggregate[0].OutputTokens
		if aggregate[0].CachedInputTokens == nil {
			allCacheReported = false
		} else {
			cached = *aggregate[0].CachedInputTokens
		}
	} else if len(aggregate) > 1 {
		return Telemetry{}, errors.New("multiple aggregate usage events cannot be deduplicated")
	}
	completeTotals := (rootUsage > 0 || len(aggregate) == 1) && (telemetry.WorkerStarts == 0 || workerUsage > 0)
	if completeTotals {
		telemetry.TotalInput = measuredMetric(totalInput, provenance(len(aggregate)), true, "")
		telemetry.TotalOutput = measuredMetric(output, provenance(len(aggregate)), true, "")
	} else {
		telemetry.TotalInput = missingMetric(provenance(len(aggregate)), "missing root or worker usage accounting")
		telemetry.TotalOutput = missingMetric(provenance(len(aggregate)), "missing root or worker usage accounting")
	}
	if completeTotals && allCacheReported {
		telemetry.CachedInput = measuredMetric(cached, provenance(len(aggregate)), true, "")
	} else {
		telemetry.CachedInput = missingMetric(provenance(len(aggregate)), "cached input was not reported for every counted usage event")
	}
	if rootUsage > 0 && allRootRequestReported {
		telemetry.PeakRootRequestInput = pointerMetric(peak, "maximum of every explicit root per-request usage event", "")
	} else {
		telemetry.PeakRootRequestInput = missingMetric("maximum of every explicit root per-request usage event", "root per-request input was not reported for every root usage event")
	}
	if rootUsage > 0 && allWindowReported {
		telemetry.ActualWindowOccupancy = pointerMetric(occupancy, "explicit window occupancy on every root usage event", "")
	} else {
		telemetry.ActualWindowOccupancy = missingMetric("explicit window occupancy on every root usage event", "actual window occupancy was not reported for every root usage event")
	}
	if telemetry.TotalInput.Complete && telemetry.TotalOutput.Complete && *telemetry.TotalInput.Value+*telemetry.TotalOutput.Value > 30000 {
		return telemetry, errors.New("observed aggregate token stop threshold exceeded")
	}
	return telemetry, nil
}

func measuredMetric(value int64, source string, complete bool, reason string) UsageMetric {
	copy := value
	if complete {
		reason = ""
	}
	return UsageMetric{Value: &copy, Provenance: source, Complete: complete, MissingReason: reason}
}

func pointerMetric(value *int64, source, reason string) UsageMetric {
	if value == nil {
		return UsageMetric{Provenance: source, MissingReason: reason}
	}
	copy := *value
	return UsageMetric{Value: &copy, Provenance: source, Complete: true}
}

func missingMetric(source, reason string) UsageMetric {
	return UsageMetric{Provenance: source, Complete: false, MissingReason: reason}
}

func unavailableTelemetry(reason string) *Telemetry {
	metric := func(name string) UsageMetric { return missingMetric("native harness event stream: "+name, reason) }
	return &Telemetry{
		RootInputCumulative: metric("root cumulative input"), WorkerInputCumulative: metric("worker cumulative input"),
		TotalInput: metric("aggregate input"), TotalOutput: metric("aggregate output"), CachedInput: metric("cached input"),
		PeakRootRequestInput: metric("peak root per-request input"), ActualWindowOccupancy: metric("actual context-window occupancy"),
	}
}

func provenance(aggregateCount int) string {
	if aggregateCount == 1 {
		return "one unique harness aggregate event; components not added"
	}
	return "sum of unique per-agent usage events"
}
