package portable

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Opt-in, model-free contract check against an explicitly selected native Codex.
// Only disposable state and synthetic project/instance-user roots are written.
func TestNativeCodexSkillInheritance(t *testing.T) {
	binary := os.Getenv("FRIDAY_TEST_CODEX")
	if binary == "" {
		t.Skip("set FRIDAY_TEST_CODEX to an absolute Codex executable")
	}
	if !filepath.IsAbs(binary) {
		t.Fatal("FRIDAY_TEST_CODEX must be absolute")
	}
	root := t.TempDir()
	project := filepath.Join(root, "project")
	writeSkill := func(dir, name string) {
		t.Helper()
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\ndescription: Synthetic discovery-only test fixture.\n---\nDo nothing.\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeSkill(filepath.Join(project, ".agents/skills/project-canary"), "friday-project-canary")
	writeSkill(filepath.Join(root, "linked-target"), "friday-user-canary")
	for _, userAvailable := range []bool{false, true} {
		t.Run(map[bool]string{false: "without-user-skill", true: "with-user-skill"}[userAvailable], func(t *testing.T) {
			s := fixtureStore(t)
			instance, err := Bind(s, filepath.Join(t.TempDir(), "instance"), "native-pilot", "/synthetic/my-friday", "device-laptop")
			if err != nil {
				t.Fatal(err)
			}
			state := filepath.Join(instance.Root, "codex")
			userSkills := filepath.Join(state, "skills")
			if err := os.Mkdir(userSkills, 0700); err != nil {
				t.Fatal(err)
			}
			if userAvailable {
				if err := os.Symlink(filepath.Join(root, "linked-target"), filepath.Join(userSkills, "alias")); err != nil {
					t.Fatal(err)
				}
			}
			plan, err := instance.Plan(s, "codex", project, []string{"app-server"})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, plan.Arguments...)
			cmd.Dir = plan.Directory
			for _, env := range plan.Environment {
				if !strings.HasPrefix(env, "CODEX_") && !strings.HasPrefix(env, "OPENAI_") {
					cmd.Env = append(cmd.Env, env)
				}
			}
			cmd.Env = append(cmd.Env, "CODEX_HOME="+state)
			in, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			out, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() { in.Close(); cancel(); cmd.Wait() }()
			enc := json.NewEncoder(in)
			scanner := bufio.NewScanner(out)
			scanner.Buffer(make([]byte, 65536), 4*1024*1024)
			request := func(id int, method string, params any) json.RawMessage {
				t.Helper()
				if err := enc.Encode(map[string]any{"id": id, "method": method, "params": params}); err != nil {
					t.Fatal(err)
				}
				for scanner.Scan() {
					var reply struct {
						ID     int             `json:"id"`
						Result json.RawMessage `json:"result"`
						Error  json.RawMessage `json:"error"`
					}
					if json.Unmarshal(scanner.Bytes(), &reply) == nil && reply.ID == id {
						if len(reply.Error) != 0 {
							t.Fatalf("%s returned a protocol error", method)
						}
						return reply.Result
					}
				}
				t.Fatalf("%s ended without a response: %v", method, scanner.Err())
				return nil
			}
			request(1, "initialize", map[string]any{"clientInfo": map[string]string{"name": "my_friday_test", "version": "1"}, "capabilities": map[string]bool{"experimentalApi": true}})
			if err := enc.Encode(map[string]string{"method": "initialized"}); err != nil {
				t.Fatal(err)
			}
			result := request(2, "skills/list", map[string]any{"cwds": []string{project}, "forceReload": true})
			var catalog struct {
				Data []struct {
					Skills []struct {
						Name    string `json:"name"`
						Enabled bool   `json:"enabled"`
					} `json:"skills"`
				} `json:"data"`
			}
			if err := json.Unmarshal(result, &catalog); err != nil {
				t.Fatal(err)
			}
			found := map[string]bool{}
			for _, entry := range catalog.Data {
				for _, skill := range entry.Skills {
					found[skill.Name] = skill.Enabled
				}
			}
			if !found["friday-project-canary"] {
				t.Fatal("project skill missing or disabled")
			}
			if enabled, exists := found["friday-user-canary"]; exists != userAvailable || enabled != userAvailable {
				t.Fatalf("user skill enabled=%v exists=%v, available=%v", enabled, exists, userAvailable)
			}
		})
	}
}
