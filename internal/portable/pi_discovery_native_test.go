package portable

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Opt-in native resource loading only: no model, auth, or lifecycle execution.
func TestNativePiResourceInheritance(t *testing.T) {
	pkg := os.Getenv("FRIDAY_TEST_PI_PACKAGE")
	if pkg == "" {
		t.Skip("set FRIDAY_TEST_PI_PACKAGE to an installed Pi package directory")
	}
	if !filepath.IsAbs(pkg) {
		t.Fatal("FRIDAY_TEST_PI_PACKAGE must be absolute")
	}
	s := fixtureStore(t)
	i, err := Bind(s, filepath.Join(t.TempDir(), "instance"), "pi-pilot", "/synthetic/my-friday", "device-laptop")
	if err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	state := filepath.Join(i.Root, "pi")
	write := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(project, ".agents/skills/project/SKILL.md"), "---\nname: friday-project-canary\ndescription: Synthetic project fixture.\n---\nDo nothing.\n")
	target := filepath.Join(t.TempDir(), "user-skill")
	write(filepath.Join(target, "SKILL.md"), "---\nname: friday-user-canary\ndescription: Synthetic user fixture.\n---\nDo nothing.\n")
	if err := os.MkdirAll(filepath.Join(state, "skills"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(state, "skills/user")); err != nil {
		t.Fatal(err)
	}
	const settings = `{"defaultModel":"synthetic-model","theme":"dark","futureField":{"canary":true}}`
	write(filepath.Join(state, "settings.json"), settings)
	plan, err := i.Plan(s, "pi", project, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	// Import from the supplied installed package so the test exercises its real
	// discovery implementation without installing npm dependencies into this repo.
	program := `
const {pathToFileURL} = await import('node:url');
const {join} = await import('node:path');
const {readFile} = await import('node:fs/promises');
const pkg = process.argv[1];
const {DefaultResourceLoader} = await import(pathToFileURL(join(pkg,'dist/core/resource-loader.js')));
const {SettingsManager} = await import(pathToFileURL(join(pkg,'dist/core/settings-manager.js')));
const state = process.env.PI_CODING_AGENT_DIR;
const settings = SettingsManager.create(process.cwd(), state);
settings.setProjectTrusted(true);
const loader = new DefaultResourceLoader({cwd:process.cwd(),agentDir:state,settingsManager:settings});
await loader.reload();
const names = new Set(loader.getSkills().skills.map(s=>s.name));
if (!names.has('friday-user-canary') || !names.has('friday-project-canary')) throw Error('Expected inherited skills missing');
const ext = loader.getExtensions();
if (ext.errors.length || !ext.extensions.some(e=>e.handlers.has('before_agent_start') && e.handlers.has('agent_settled'))) throw Error('My Friday lifecycle extension missing');
const data = JSON.parse(await readFile(join(state,'settings.json'),'utf8'));
if (data.defaultModel !== 'synthetic-model' || !data.futureField.canary || data.skills !== undefined) throw Error('Native settings changed');
console.log('PASS native Pi skill inheritance and lifecycle extension discovery');
`
	cmd := exec.CommandContext(ctx, "node", "--input-type=module", "-e", program, pkg)
	cmd.Dir, cmd.Env = plan.Directory, plan.Environment
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native Pi discovery failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "PASS native Pi") {
		t.Fatal("missing native result")
	}
	got, err := os.ReadFile(filepath.Join(state, "settings.json"))
	if err != nil || string(got) != settings {
		t.Fatalf("native settings changed: %v", err)
	}
}
