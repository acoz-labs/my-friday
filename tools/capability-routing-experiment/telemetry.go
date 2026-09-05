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
	RootInputCumulative   UsageMetric       `json:"root_input_cumulative"`
	WorkerInputCumulative UsageMetric       `json:"worker_input_cumulative"`
	TotalInput            UsageMetric       `json:"total_input"`
	TotalOutput           UsageMetric       `json:"total_output"`
	CachedInput           UsageMetric       `json:"cached_input"`
	PeakRootRequestInput  UsageMetric       `json:"peak_root_request_input"`
	ActualWindowOccupancy UsageMetric       `json:"actual_window_occupancy"`
	ToolCalls             int               `json:"tool_calls"`
	WorkerStarts          int               `json:"worker_starts"`
	WorkerReturns         int               `json:"worker_returns"`
	WorkerLifecycles      []WorkerLifecycle `json:"worker_lifecycles"`
	PolicyLoss            bool              `json:"policy_loss"`
}

type WorkerLifecycle struct {
	AgentID       string   `json:"agent_id"`
	StartEventID  string   `json:"start_event_id"`
	UsageEventIDs []string `json:"usage_event_ids"`
	ReturnEventID string   `json:"return_event_id"`
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
	workerID := ""
	workerReturned := false
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
					if workerID == "" || workerReturned || event.AgentID != workerID {
						return Telemetry{}, errors.New("worker usage lacks a matching active worker identity")
					}
					workerUsage++
					telemetry.WorkerLifecycles[0].UsageEventIDs = append(telemetry.WorkerLifecycles[0].UsageEventIDs, event.EventID)
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
			if event.AgentID == "" || event.WorkerDepth == nil || *event.WorkerDepth != 1 || telemetry.WorkerStarts > 1 {
				return Telemetry{}, errors.New("observed worker count/depth ceiling exceeded")
			}
			workerID = event.AgentID
			telemetry.WorkerLifecycles = append(telemetry.WorkerLifecycles, WorkerLifecycle{AgentID: event.AgentID, StartEventID: event.EventID, UsageEventIDs: []string{}})
		case "worker_return":
			telemetry.WorkerReturns++
			if event.AgentID == "" || event.AgentID != workerID || workerReturned || telemetry.WorkerReturns > telemetry.WorkerStarts {
				return Telemetry{}, errors.New("worker return lacks matching start")
			}
			workerReturned = true
			telemetry.WorkerLifecycles[0].ReturnEventID = event.EventID
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
	if err := ValidateTelemetry(&telemetry); err != nil {
		return telemetry, err
	}
	return telemetry, nil
}

func ValidateTelemetry(telemetry *Telemetry) error {
	if telemetry == nil {
		return errors.New("telemetry is nil")
	}
	metrics := map[string]UsageMetric{
		"root input": telemetry.RootInputCumulative, "worker input": telemetry.WorkerInputCumulative,
		"total input": telemetry.TotalInput, "total output": telemetry.TotalOutput,
		"cached input": telemetry.CachedInput, "peak root input": telemetry.PeakRootRequestInput,
		"window occupancy": telemetry.ActualWindowOccupancy,
	}
	for name, metric := range metrics {
		if metric.Provenance == "" {
			return fmt.Errorf("%s metric lacks provenance", name)
		}
		if metric.Complete {
			if metric.Value == nil || *metric.Value < 0 || metric.MissingReason != "" {
				return fmt.Errorf("complete %s metric has invalid value or missing reason", name)
			}
		} else if metric.Value != nil || metric.MissingReason == "" {
			return fmt.Errorf("incomplete %s metric must be null with a reason", name)
		}
	}
	if telemetry.ToolCalls < 0 || telemetry.ToolCalls > 8 || telemetry.WorkerStarts < 0 || telemetry.WorkerStarts > 1 || telemetry.WorkerReturns != telemetry.WorkerStarts {
		return errors.New("telemetry count invariants failed")
	}
	if len(telemetry.WorkerLifecycles) != telemetry.WorkerStarts {
		return errors.New("worker lifecycle evidence does not match worker counts")
	}
	if telemetry.WorkerStarts == 1 {
		worker := telemetry.WorkerLifecycles[0]
		if worker.AgentID == "" || worker.StartEventID == "" || worker.ReturnEventID == "" || worker.StartEventID == worker.ReturnEventID {
			return errors.New("worker lifecycle identity or events are incomplete")
		}
		seenEvents := map[string]bool{worker.StartEventID: true, worker.ReturnEventID: true}
		for _, eventID := range worker.UsageEventIDs {
			if eventID == "" || seenEvents[eventID] {
				return errors.New("worker lifecycle contains missing or duplicate event identity")
			}
			seenEvents[eventID] = true
		}
		if telemetry.WorkerInputCumulative.Complete && len(worker.UsageEventIDs) == 0 {
			return errors.New("complete worker usage lacks identity-bound events")
		}
	}
	if telemetry.TotalInput.Complete {
		if !telemetry.RootInputCumulative.Complete || !telemetry.WorkerInputCumulative.Complete || *telemetry.TotalInput.Value != *telemetry.RootInputCumulative.Value+*telemetry.WorkerInputCumulative.Value {
			return errors.New("total input does not match complete root and worker accounting")
		}
		if telemetry.CachedInput.Complete && *telemetry.CachedInput.Value > *telemetry.TotalInput.Value {
			return errors.New("cached input exceeds total input")
		}
	}
	if telemetry.WorkerStarts == 0 && telemetry.WorkerInputCumulative.Complete && *telemetry.WorkerInputCumulative.Value != 0 {
		return errors.New("worker input reported without a worker lifecycle")
	}
	if telemetry.WorkerStarts == 1 && telemetry.TotalInput.Complete && !telemetry.WorkerInputCumulative.Complete {
		return errors.New("complete total lacks launched worker accounting")
	}
	if telemetry.PeakRootRequestInput.Complete && (!telemetry.RootInputCumulative.Complete || *telemetry.PeakRootRequestInput.Value > *telemetry.RootInputCumulative.Value) {
		return errors.New("peak root input is inconsistent with cumulative root input")
	}
	if telemetry.TotalInput.Complete && telemetry.TotalOutput.Complete && *telemetry.TotalInput.Value+*telemetry.TotalOutput.Value > 30000 {
		return errors.New("aggregate token stop threshold exceeded")
	}
	return nil
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
		PeakRootRequestInput: metric("peak root per-request input"), ActualWindowOccupancy: metric("actual context-window occupancy"), WorkerLifecycles: []WorkerLifecycle{},
	}
}

func provenance(aggregateCount int) string {
	if aggregateCount == 1 {
		return "one unique harness aggregate event; components not added"
	}
	return "sum of unique per-agent usage events"
}
