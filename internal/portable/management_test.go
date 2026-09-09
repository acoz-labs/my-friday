package portable

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestManagedHarnessChangeIsCleanAndAttributed(t *testing.T) {
	s := fixtureStore(t)
	ctx := context.Background()
	if err := s.InitGit(ctx); err != nil {
		t.Fatal(err)
	}
	before := s.Agent.ID
	if err := s.SetDefaultHarness(ctx, "pi"); err != nil {
		t.Fatal(err)
	}
	reloaded, err := Open(s.Root)
	if err != nil || reloaded.Agent.ID != before || reloaded.Agent.DefaultHarness != "pi" {
		t.Fatalf("%+v %v", reloaded, err)
	}
	if gitTest(t, s.Root, "status", "--porcelain") != "" {
		t.Fatal("uncheckpointed change")
	}
	os.WriteFile(filepath.Join(s.Root, "instructions/draft.md"), []byte("unfinished"), 0600)
	if err := reloaded.SetDefaultHarness(ctx, "codex"); err == nil {
		t.Fatal("checkpointed unrelated dirt")
	}
	if err := reloaded.SetDefaultHarness(ctx, "unknown"); err == nil {
		t.Fatal("accepted invalid harness")
	}
}

func TestRebindPreservesNativeAndRejectsCustomLauncher(t *testing.T) {
	s := fixtureStore(t)
	binary, _ := os.Executable()
	binary, _ = filepath.EvalSymlinks(binary)
	i, err := Bind(s, filepath.Join(t.TempDir(), "instance"), "pilot", "/old/toolkit", "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	launcher := filepath.Join(t.TempDir(), "pilot")
	if err := i.InstallLauncher(launcher); err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(i.Root, "codex/config.toml")
	os.WriteFile(native, []byte("native-owned"), 0600)
	backup, err := i.UseToolkit(s, binary, launcher)
	if err != nil {
		t.Fatal(err)
	}
	next, _, err := LoadInstance(i.Root)
	if err != nil || next.Binary != binary {
		t.Fatalf("%+v %v", next, err)
	}
	if data, _ := os.ReadFile(native); string(data) != "native-owned" {
		t.Fatal("native changed")
	}
	if _, err := os.Stat(filepath.Join(backup, "receipt.json")); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(launcher, []byte("custom launcher"), 0700)
	if _, err := next.UseToolkit(s, binary, launcher); err == nil {
		t.Fatal("custom launcher overwritten")
	}
}

func TestToolkitRollbackPreservesSourceAndDetectsLaterEdits(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(fmt.Sprint(changed), func(t *testing.T) {
			s := fixtureStore(t)
			if err := s.InitGit(context.Background()); err != nil {
				t.Fatal(err)
			}
			head := gitTest(t, s.Root, "rev-parse", "HEAD")
			binary, _ := os.Executable()
			i, err := Bind(s, filepath.Join(t.TempDir(), "instance"), "pilot", "/previous/toolkit", "device-laptop")
			if err != nil {
				t.Fatal(err)
			}
			launcher := filepath.Join(t.TempDir(), "pilot")
			if err := i.InstallLauncher(launcher); err != nil {
				t.Fatal(err)
			}
			oldBinding, _ := os.ReadFile(filepath.Join(i.Root, "binding.json"))
			oldLauncher, _ := os.ReadFile(launcher)
			backup, err := i.UseToolkit(s, binary, launcher)
			if err != nil {
				t.Fatal(err)
			}
			next, _, err := LoadInstance(i.Root)
			if err != nil {
				t.Fatal(err)
			}
			if changed {
				os.WriteFile(launcher, []byte("later user edit"), 0700)
			}
			err = next.RollbackToolkit(s, backup)
			if changed {
				if err == nil {
					t.Fatal("overwrote later change")
				}
				data, _ := os.ReadFile(launcher)
				if string(data) != "later user edit" {
					t.Fatal("lost edit")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			binding, _ := os.ReadFile(filepath.Join(i.Root, "binding.json"))
			script, _ := os.ReadFile(launcher)
			if !bytes.Equal(binding, oldBinding) || !bytes.Equal(script, oldLauncher) {
				t.Fatal("rollback mismatch")
			}
			if gitTest(t, s.Root, "rev-parse", "HEAD") != head || gitTest(t, s.Root, "status", "--porcelain") != "" {
				t.Fatal("source changed")
			}
		})
	}
}

func TestDiscoverIncludesBrokenButDoesNotFollowInstanceSymlink(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "broken"), 0700)
	os.Symlink(t.TempDir(), filepath.Join(dir, "linked"))
	entries, err := DiscoverInstances(dir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("%+v %v", entries, err)
	}
	for _, entry := range entries {
		if entry.Problem == "" {
			t.Fatal("broken entry hidden")
		}
	}
}
