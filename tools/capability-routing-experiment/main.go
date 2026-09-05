package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fatalUsage(errors.New("subcommand required"))
	}
	var err error
	switch os.Args[1] {
	case "validate":
		err = validateCommand(os.Args[2:])
	case "prepare":
		err = prepareCommand(os.Args[2:])
	case "run":
		err = runCommand(os.Args[2:])
	case "score":
		err = scoreCommand(os.Args[2:])
	case "report":
		err = reportCommand(os.Args[2:])
	default:
		fatalUsage(fmt.Errorf("unknown subcommand %q", os.Args[1]))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func validateCommand(args []string) error {
	set := flag.NewFlagSet("validate", flag.ContinueOnError)
	data := set.String("data", "testdata", "corpus directory")
	if err := set.Parse(args); err != nil {
		return err
	}
	bundle, err := LoadBundle(*data)
	if err != nil {
		return err
	}
	if err = ValidateBundle(bundle); err != nil {
		return err
	}
	if bundle.Manifest.Version != 0 {
		return validateManifest(bundle)
	}
	return nil
}

func prepareCommand(args []string) error {
	set := flag.NewFlagSet("prepare", flag.ContinueOnError)
	data := set.String("data", "testdata", "corpus directory")
	out := set.String("out", "", "manifest destination")
	sourceCommit := set.String("source-commit", "", "exact frozen source commit")
	stageRoot := set.String("stage-root", "", "optional fresh model-visible trial root")
	trialID := set.String("trial", "", "manifest trial to stage with --stage-root")
	if err := set.Parse(args); err != nil {
		return err
	}
	if *out == "" {
		*out = filepath.Join(*data, "manifest.json")
	}
	bundle, err := LoadBundle(*data)
	if err != nil {
		return err
	}
	if *sourceCommit != TrustedSourceCommit {
		return fmt.Errorf("--source-commit must equal trusted preregistration %s", TrustedSourceCommit)
	}
	manifest, err := PrepareManifest(bundle, *sourceCommit, TrustedHarnesses())
	if err != nil {
		return err
	}
	if err = WriteOrVerifyJSON(*out, manifest); err != nil {
		return err
	}
	if *stageRoot == "" && *trialID == "" {
		return nil
	}
	if *stageRoot == "" || *trialID == "" {
		return errors.New("--stage-root and --trial must be supplied together")
	}
	bundle.Manifest = manifest
	for _, cell := range manifest.Cells {
		if cell.TrialID == *trialID {
			return StageTrial(*stageRoot, bundle, cell)
		}
	}
	return fmt.Errorf("unknown manifest trial %q", *trialID)
}

func runCommand(args []string) error {
	set := flag.NewFlagSet("run", flag.ContinueOnError)
	data := set.String("data", "testdata", "corpus directory")
	runRoot := set.String("run-root", "", "owned evidence directory")
	live := set.Bool("live", false, "explicitly opt into native preflight and eligible live cells")
	if err := set.Parse(args); err != nil {
		return err
	}
	if !*live {
		return errors.New("run refuses without explicit --live opt-in")
	}
	if *runRoot == "" {
		return errors.New("--run-root is required")
	}
	bundle, err := LoadBundle(*data)
	if err != nil {
		return err
	}
	if err = ValidateBundle(bundle); err != nil {
		return err
	}
	if err = validateManifest(bundle); err != nil {
		return err
	}
	lock, err := AcquireRunLock(*runRoot)
	if err != nil {
		return err
	}
	defer lock.Release()
	var probes []DriverProbe
	expectedVersions := map[string]string{}
	for _, spec := range bundle.Manifest.Harnesses {
		expectedVersions[spec.ID] = spec.ExecutableVersion
	}
	for _, harness := range []HarnessName{HarnessCodex, HarnessClaude} {
		path, lookupErr := exec.LookPath(string(harness))
		if lookupErr != nil {
			probes = append(probes, DriverProbe{Version: SchemaVersion, Harness: harness, Executable: string(harness), State: "unavailable", Unavailable: []string{"installed executable was not found"}})
			continue
		}
		probe, probeErr := ProbeHarness(context.Background(), harness, path)
		if probeErr != nil {
			probes = append(probes, DriverProbe{Version: SchemaVersion, Harness: harness, Executable: string(harness), State: "unavailable", Unavailable: []string{probeErr.Error()}})
			continue
		}
		if probe.CLIRevision != expectedVersions[string(harness)] {
			probe.State = "unavailable"
			probe.Unavailable = append(probe.Unavailable, "installed CLI revision differs from the preregistered harness revision")
		}
		probes = append(probes, probe)
	}
	probeByHarness := map[string]DriverProbe{}
	for _, probe := range probes {
		probeByHarness[string(probe.Harness)] = probe
	}
	attempts := AttemptSet{Version: SchemaVersion, ManifestSHA256: digestJSON(bundle.Manifest)}
	for _, cell := range bundle.Manifest.Cells {
		probe := probeByHarness[cell.HarnessID]
		reason := strings.Join(probe.Unavailable, "; ")
		if probe.LiveEligible(cell.Mode) {
			return errors.New("supported live driver execution is intentionally absent until native boundary implementation is independently reviewed")
		}
		attempts.Attempts = append(attempts.Attempts, Attempt{TrialID: cell.TrialID, AttemptID: cell.TrialID + "-a1", Primary: true, State: "unavailable", Reason: reason, SelectedCapabilities: []string{}, ResultFacts: []string{}, AttemptedEffects: []string{}, ActualEffects: []string{}, FixtureSnapshot: []FixtureSnapshot{}, Telemetry: unavailableTelemetry("native driver unavailable; no model trial executed")})
	}
	if err = WriteOrVerifyJSON(filepath.Join(*runRoot, "probes.json"), probes); err != nil {
		return err
	}
	return WriteOrVerifyJSON(filepath.Join(*runRoot, "attempts.json"), attempts)
}

func scoreCommand(args []string) error {
	set := flag.NewFlagSet("score", flag.ContinueOnError)
	data := set.String("data", "testdata", "corpus directory")
	runRoot := set.String("run-root", "", "owned evidence directory")
	if err := set.Parse(args); err != nil {
		return err
	}
	bundle, err := LoadBundle(*data)
	if err != nil {
		return err
	}
	var attempts AttemptSet
	if err = ReadJSON(filepath.Join(*runRoot, "attempts.json"), &attempts); err != nil {
		return err
	}
	comparison, err := ScoreAttempts(bundle, attempts)
	if err != nil {
		return err
	}
	return WriteOrVerifyJSON(filepath.Join(*runRoot, "scores.json"), comparison)
}

func reportCommand(args []string) error {
	set := flag.NewFlagSet("report", flag.ContinueOnError)
	data := set.String("data", "testdata", "corpus directory")
	runRoot := set.String("run-root", "", "owned evidence directory")
	if err := set.Parse(args); err != nil {
		return err
	}
	var comparison Comparison
	if err := ReadJSON(filepath.Join(*runRoot, "scores.json"), &comparison); err != nil {
		return err
	}
	var probes []DriverProbe
	if err := ReadJSON(filepath.Join(*runRoot, "probes.json"), &probes); err != nil {
		return err
	}
	bundle, err := LoadBundle(*data)
	if err != nil {
		return err
	}
	var attempts AttemptSet
	if err = ReadJSON(filepath.Join(*runRoot, "attempts.json"), &attempts); err != nil {
		return err
	}
	if err = ValidateReportInputs(bundle, attempts, comparison, probes); err != nil {
		return err
	}
	body, err := MarshalReport(comparison, probes)
	if err != nil {
		return err
	}
	if err = WriteOrVerifyBytes(filepath.Join(*runRoot, "report.json"), append(body, '\n')); err != nil {
		return err
	}
	return WriteOrVerifyBytes(filepath.Join(*runRoot, "report.md"), []byte(RenderMarkdown(comparison, probes)))
}

func validateManifest(bundle Bundle) error {
	manifest := bundle.Manifest
	if err := validateTrustedIdentity(manifest); err != nil {
		return err
	}
	expected, err := PrepareManifest(bundle, TrustedSourceCommit, TrustedHarnesses())
	if err != nil {
		return err
	}
	if digestJSON(expected) != digestJSON(manifest) {
		return errors.New("manifest differs from the deterministic harness-task-mode-repetition declaration")
	}
	return nil
}

func createImmutableBytes(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(body)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func fatalUsage(err error) {
	fmt.Fprintln(os.Stderr, "usage: capability-routing-experiment {validate|prepare|run|score|report} [options]")
	fmt.Fprintln(os.Stderr, err)
	os.Exit(64)
}
