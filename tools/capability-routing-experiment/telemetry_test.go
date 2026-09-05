package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestTelemetryDeduplicatesAggregateAndReportsDistinctMetrics(t *testing.T) {
	input := strings.Join([]string{
		`{"event_id":"w-start","type":"worker_start","agent_id":"worker","worker_depth":1}`,
		`{"event_id":"root-1","type":"usage","agent_id":"root","role":"root","input_tokens":100,"output_tokens":10,"cached_input_tokens":20,"root_request_input_tokens":100}`,
		`{"event_id":"worker-1","type":"usage","agent_id":"worker","role":"worker","input_tokens":70,"output_tokens":7,"cached_input_tokens":5}`,
		`{"event_id":"w-return","type":"worker_return","agent_id":"worker"}`,
		`{"event_id":"aggregate","type":"usage","agent_id":"all","role":"root","aggregate":true,"input_tokens":170,"output_tokens":17,"cached_input_tokens":25}`,
	}, "\n")
	telemetry, err := ParseTelemetry(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if *telemetry.TotalInput.Value != 170 || *telemetry.RootInputCumulative.Value != 100 || *telemetry.WorkerInputCumulative.Value != 70 {
		t.Fatalf("telemetry = %#v", telemetry)
	}
	if !telemetry.TotalInput.Complete || !telemetry.PeakRootRequestInput.Complete {
		t.Fatal("complete telemetry marked missing")
	}
	if telemetry.ActualWindowOccupancy.Complete || telemetry.ActualWindowOccupancy.Value != nil {
		t.Fatal("peak request input was mislabeled as window occupancy")
	}
}

func TestTelemetryRejectsUnknownDuplicateAndMissingWorkerUsage(t *testing.T) {
	for name, input := range map[string]string{
		"unknown":   `{"event_id":"x","type":"mystery"}`,
		"field":     `{"event_id":"x","type":"result","secret":"no"}`,
		"duplicate": "{\"event_id\":\"x\",\"type\":\"result\"}\n{\"event_id\":\"x\",\"type\":\"result\"}",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseTelemetry(strings.NewReader(input)); err == nil {
				t.Fatal("invalid telemetry accepted")
			}
		})
	}
	telemetry, err := ParseTelemetry(strings.NewReader("{\"event_id\":\"w\",\"type\":\"worker_start\",\"agent_id\":\"worker\",\"worker_depth\":1}\n{\"event_id\":\"r\",\"type\":\"usage\",\"agent_id\":\"root\",\"role\":\"root\",\"input_tokens\":1,\"output_tokens\":1}\n{\"event_id\":\"wr\",\"type\":\"worker_return\",\"agent_id\":\"worker\"}"))
	if err != nil {
		t.Fatal(err)
	}
	if telemetry.TotalInput.Complete || telemetry.WorkerInputCumulative.Complete {
		t.Fatal("missing child usage was accepted as complete")
	}
}

func TestTelemetryDoesNotInventMissingMeasurementsAsZero(t *testing.T) {
	telemetry, err := ParseTelemetry(strings.NewReader(`{"event_id":"done","type":"result"}`))
	if err != nil {
		t.Fatal(err)
	}
	for name, metric := range map[string]UsageMetric{"root": telemetry.RootInputCumulative, "worker": telemetry.WorkerInputCumulative, "input": telemetry.TotalInput, "output": telemetry.TotalOutput, "cache": telemetry.CachedInput, "peak": telemetry.PeakRootRequestInput} {
		if metric.Complete || metric.Value != nil {
			t.Fatalf("%s was invented as complete zero: %#v", name, metric)
		}
	}
}

func TestTelemetryRequiresCacheAndPerRequestContextOnEveryUsageEvent(t *testing.T) {
	input := strings.Join([]string{
		`{"event_id":"r1","type":"usage","agent_id":"root","role":"root","input_tokens":10,"output_tokens":1,"cached_input_tokens":2,"root_request_input_tokens":10}`,
		`{"event_id":"r2","type":"usage","agent_id":"root","role":"root","input_tokens":5,"output_tokens":1}`,
	}, "\n")
	telemetry, err := ParseTelemetry(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if telemetry.CachedInput.Complete || telemetry.CachedInput.Value != nil {
		t.Fatal("partial cache reporting accepted")
	}
	if telemetry.PeakRootRequestInput.Complete || telemetry.PeakRootRequestInput.Value != nil {
		t.Fatal("partial root-request reporting accepted")
	}
}

func TestTelemetryStopsOnCriticalLossAndEnforcesObservedLimits(t *testing.T) {
	telemetry, err := ParseTelemetry(strings.NewReader("{\"event_id\":\"loss\",\"type\":\"policy_loss\"}\n{\"event_id\":\"after\",\"type\":\"result\"}"))
	if !errors.Is(err, ErrCriticalPolicyLoss) || !telemetry.PolicyLoss {
		t.Fatalf("telemetry=%#v err=%v", telemetry, err)
	}
	var calls []string
	for index := 0; index < 9; index++ {
		calls = append(calls, fmt.Sprintf(`{"event_id":"t%d","type":"tool_call"}`, index))
	}
	if _, err = ParseTelemetry(strings.NewReader(strings.Join(calls, "\n"))); err == nil {
		t.Fatal("ninth tool call accepted")
	}
	if _, err = ParseTelemetry(strings.NewReader(`{"event_id":"w","type":"worker_start","worker_depth":2}`)); err == nil {
		t.Fatal("worker depth two accepted")
	}
}

func TestTelemetryRejectsUnmatchedWorkerIdentityAndUsage(t *testing.T) {
	tests := map[string]string{
		"usage-without-start":   `{"event_id":"usage","type":"usage","agent_id":"worker","role":"worker","input_tokens":7,"output_tokens":1}`,
		"wrong-usage-identity":  "{\"event_id\":\"start\",\"type\":\"worker_start\",\"agent_id\":\"worker-a\",\"worker_depth\":1}\n{\"event_id\":\"usage\",\"type\":\"usage\",\"agent_id\":\"worker-b\",\"role\":\"worker\",\"input_tokens\":7,\"output_tokens\":1}",
		"wrong-return-identity": "{\"event_id\":\"start\",\"type\":\"worker_start\",\"agent_id\":\"worker-a\",\"worker_depth\":1}\n{\"event_id\":\"return\",\"type\":\"worker_return\",\"agent_id\":\"worker-b\"}",
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseTelemetry(strings.NewReader(input)); err == nil {
				t.Fatal("unmatched worker lifecycle accepted")
			}
		})
	}
}

func TestImportedTelemetryValidationRejectsMalformedMetricsAndAccounting(t *testing.T) {
	zero, one, negative := int64(0), int64(1), int64(-1)
	valid := unavailableTelemetry("unavailable")
	tests := map[string]*Telemetry{
		"complete-null": func() *Telemetry {
			value := *valid
			value.TotalInput = UsageMetric{Complete: true, Provenance: "source"}
			return &value
		}(),
		"negative": func() *Telemetry {
			value := *valid
			value.TotalInput = UsageMetric{Complete: true, Value: &negative, Provenance: "source"}
			return &value
		}(),
		"missing-provenance": func() *Telemetry {
			value := *valid
			value.TotalInput = UsageMetric{Complete: true, Value: &one}
			return &value
		}(),
		"incomplete-value": func() *Telemetry {
			value := *valid
			value.TotalInput = UsageMetric{Value: &zero, Provenance: "source", MissingReason: "missing"}
			return &value
		}(),
		"inconsistent-total": completeTelemetry(100, 10, 99, 5, 50, 50),
		"unmatched-worker-evidence": func() *Telemetry {
			value := completeTelemetry(100, 10, 110, 5, 50, 50)
			value.WorkerStarts, value.WorkerReturns = 1, 1
			return value
		}(),
	}
	for name, telemetry := range tests {
		t.Run(name, func(t *testing.T) {
			if err := ValidateTelemetry(telemetry); err == nil {
				t.Fatal("invalid imported telemetry accepted")
			}
		})
	}
}
