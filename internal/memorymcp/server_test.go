package memorymcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/acoz-labs/my-friday/internal/memorybank"
	"github.com/acoz-labs/my-friday/internal/portable"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPDiscoveryWritesRecallAndValidation(t *testing.T) {
	store, err := portable.CreateMemoryBank(filepath.Join(t.TempDir(), "bank"), "Example", "device-test", "Test machine")
	if err != nil {
		t.Fatal(err)
	}
	service, err := memorybank.Open(store.Root, portable.Authorship{DeviceID: "device-test", Actor: "Example user", Harness: "codex"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := New(service).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	client, err := mcp.NewClient(&mcp.Implementation{Name: "synthetic", Version: "1"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	listed, err := client.ListTools(ctx, nil)
	if err != nil || len(listed.Tools) != 7 {
		t.Fatalf("discovery: %+v %v", listed, err)
	}
	for _, tool := range listed.Tools {
		if tool.InputSchema == nil || tool.Annotations == nil {
			t.Fatalf("missing contract: %+v", tool)
		}
	}
	call := func(name string, args any) *mcp.CallToolResult {
		t.Helper()
		result, err := client.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	if result := call("memory_remember", map[string]any{"kind": "fact"}); !result.IsError {
		t.Fatal("invalid input accepted")
	}
	input := map[string]any{"kind": "decision", "summary": "Project name", "body": "Silver Heron", "basis": "user-direction", "reason": "Selected by user"}
	if result := call("memory_remember", input); result.IsError {
		t.Fatalf("remember: %+v", result)
	}
	recalled := call("memory_recall", map[string]any{"query": "project"})
	if recalled.IsError {
		t.Fatalf("recall: %+v", recalled)
	}
	encoded, err := json.Marshal(recalled.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var packet memorybank.RecallPacket
	if err := json.Unmarshal(encoded, &packet); err != nil || len(packet.Current) != 1 || packet.Current[0].Body != "Silver Heron" {
		t.Fatalf("recall packet: %s %v", encoded, err)
	}
	if result := call("memory_recall", map[string]any{"query": "project", "unknown": true}); !result.IsError {
		t.Fatal("unknown argument silently ignored")
	}
	entries, err := service.Journal("", 10)
	if err != nil || len(entries) != 0 {
		t.Fatalf("MCP invented journal entries: %+v %v", entries, err)
	}
	if result := call("memory_journal_append", map[string]any{"kind": "session", "summary": "Selected Silver Heron."}); result.IsError {
		t.Fatalf("journal append: %+v", result)
	}
	if result := call("memory_history", map[string]any{"record_id": packet.Current[0].RecordID}); result.IsError {
		t.Fatalf("history: %+v", result)
	}
}
