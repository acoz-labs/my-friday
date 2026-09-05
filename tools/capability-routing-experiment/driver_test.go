package main

import "testing"

func TestUnavailableProbeNeverBecomesLiveEligible(t *testing.T) {
	probe := DriverProbe{State: "unavailable", Controls: DriverControl{OSFixtureOnlyReadBoundary: true, ModelToolNetworkDenied: true, NativeSkillVisibility: true, NativeBodyReadConstrained: true, NativeWorkerEvents: true, WorkerPrelaunchLimit: true, BuiltinPredispatchLimit: true, BrokerPredispatchLimit: true}}
	for _, mode := range RoutingModes {
		if probe.LiveEligible(mode) {
			t.Fatalf("unavailable probe eligible for %s", mode)
		}
	}
}

func TestModeSpecificDriverControls(t *testing.T) {
	probe := DriverProbe{State: "supported", Controls: DriverControl{OSFixtureOnlyReadBoundary: true, ModelToolNetworkDenied: true, BuiltinPredispatchLimit: true, BrokerPredispatchLimit: true}}
	if !probe.LiveEligible("lookup-direct") {
		t.Fatal("broker-only direct mode should be eligible")
	}
	if probe.LiveEligible("lookup-worker") || probe.LiveEligible("native-catalogue") {
		t.Fatal("native modes eligible without native controls")
	}
}
