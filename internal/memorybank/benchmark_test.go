package memorybank

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acoz-labs/my-friday/internal/portable"
)

// Build a synthetic on-disk corpus directly so the benchmark measures retrieval,
// not thousands of individually validated writes. Validate it before timing.
func benchmarkBank(b *testing.B, count int) *Service {
	b.Helper()
	s, err := portable.CreateMemoryBank(filepath.Join(b.TempDir(), "bank"), "Benchmark", "device-benchmark", "Synthetic benchmark host")
	if err != nil {
		b.Fatal(err)
	}
	author := portable.Authorship{DeviceID: "device-benchmark", Actor: "Benchmark", Harness: "benchmark"}
	for n := 0; n < count; n++ {
		r := portable.Revision{
			Version: 1, ID: fmt.Sprintf("revision-%06d", n), RecordID: fmt.Sprintf("record-%06d", n),
			Kind: "fact", Scope: portable.Scope{Kind: "assistant", ID: s.Agent.ID},
			Summary:     fmt.Sprintf("Project observation %d", n),
			Body:        strings.Repeat("A synthetic project observation about tests, documentation, and previously completed work. ", 6),
			Sensitivity: "private", Volatility: "drift-prone", RecordedAt: "2026-01-01T00:00:00Z", EffectiveFrom: "2026-01-01T00:00:00Z",
			Authorship: author, Evidence: portable.Evidence{Basis: "observation", Confidence: "medium", SourceRefs: []string{}},
			Supersedes: []string{}, ChangeReason: "Synthetic benchmark fixture",
		}
		data, err := json.Marshal(r)
		if err != nil {
			b.Fatal(err)
		}
		dir := filepath.Join(s.Root, "memory/records", r.RecordID)
		if err := os.Mkdir(dir, 0700); err != nil {
			b.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, r.ID+".json"), data, 0600); err != nil {
			b.Fatal(err)
		}
	}
	if err := s.Validate(); err != nil {
		b.Fatal(err)
	}
	service, err := Open(s.Root, author)
	if err != nil {
		b.Fatal(err)
	}
	return service
}

func BenchmarkRecallCorpus(b *testing.B) {
	for _, count := range []int{100, 1000, 5000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			s := benchmarkBank(b, count)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				p, err := s.Recall("project tests completed", nil, 5, 8192)
				if err != nil || len(p.Current) != 5 || p.MatchingCount != count || !p.Truncated {
					b.Fatalf("recall failed: %+v %v", p, err)
				}
			}
		})
	}
}
