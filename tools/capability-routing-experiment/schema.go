package main

const SchemaVersion = 1

var RoutingModes = []string{"native-catalogue", "lookup-direct", "lookup-worker"}

var TaskCategories = []string{
	"explicit-selection",
	"informal-paraphrase",
	"ambiguous-alternatives",
	"no-match",
	"dependency-loading",
	"stale-index",
	"unsupported-required-semantics",
	"conflicting-instructions",
	"permission-denial",
	"short-direct-work",
	"complex-worker-work",
	"material-summary-omission",
}

type Bundle struct {
	Capabilities CapabilityCorpus
	Tasks        TaskCorpus
	Labels       LabelCorpus
	Manifest     Manifest
}

type CapabilityCorpus struct {
	Version      int          `json:"version"`
	Revision     string       `json:"revision"`
	Capabilities []Capability `json:"capabilities"`
}

type Capability struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Summary      string       `json:"summary"`
	Body         string       `json:"body"`
	Revision     string       `json:"revision"`
	Dependencies []Dependency `json:"dependencies"`
}

type Dependency struct {
	ID       string `json:"id"`
	Revision string `json:"revision"`
}

type TaskCorpus struct {
	Version  int    `json:"version"`
	Revision string `json:"revision"`
	Tasks    []Task `json:"tasks"`
}

type Task struct {
	ID                string    `json:"id"`
	Split             string    `json:"split"`
	Category          string    `json:"category"`
	Prompt            string    `json:"prompt"`
	IndexRevision     string    `json:"index_revision"`
	ReadPaths         []string  `json:"read_paths"`
	WritePaths        []string  `json:"write_paths"`
	Fixtures          []Fixture `json:"fixtures"`
	RequiredSemantics []string  `json:"required_semantics"`
	RequiresIsolation bool      `json:"requires_isolation"`
}

type Fixture struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type LabelCorpus struct {
	Version  int     `json:"version"`
	Revision string  `json:"revision"`
	Labels   []Label `json:"labels"`
}

type Label struct {
	TaskID                string          `json:"task_id"`
	AllowedCapabilitySets [][]string      `json:"allowed_capability_sets"`
	Expectation           string          `json:"expectation"`
	RequiredFacts         []string        `json:"required_facts"`
	RequiredEffects       []string        `json:"required_effects"`
	ForbiddenEffects      []string        `json:"forbidden_effects"`
	RequiredSummary       SummaryEvidence `json:"required_summary"`
}

type Manifest struct {
	Version        int            `json:"version"`
	SourceCommit   string         `json:"source_commit"`
	CorpusRevision string         `json:"corpus_revision"`
	Hashes         ManifestHashes `json:"hashes"`
	Budgets        Budgets        `json:"budgets"`
	Harnesses      []HarnessSpec  `json:"harnesses"`
	Modes          []string       `json:"modes"`
	Repetitions    int            `json:"repetitions"`
	Cells          []ManifestCell `json:"cells"`
}

type ManifestHashes struct {
	Capabilities string `json:"capabilities"`
	Tasks        string `json:"tasks"`
	Labels       string `json:"labels"`
}

type Budgets struct {
	TrialWallSeconds     int `json:"trial_wall_seconds"`
	TrialAggregateTokens int `json:"trial_aggregate_tokens"`
	TrialToolCalls       int `json:"trial_tool_calls"`
	TrialWorkers         int `json:"trial_workers"`
	TrialWorkerDepth     int `json:"trial_worker_depth"`
	AgentOutputTokens    int `json:"agent_output_tokens"`
	BatchTrials          int `json:"batch_trials"`
	BatchWallSeconds     int `json:"batch_wall_seconds"`
	BatchAggregateTokens int `json:"batch_aggregate_tokens"`
}

type HarnessSpec struct {
	ID                string `json:"id"`
	ExecutableVersion string `json:"executable_version"`
	Model             string `json:"model"`
	Config            string `json:"config"`
}

type ManifestCell struct {
	TrialID    string `json:"trial_id"`
	Sequence   int    `json:"sequence"`
	HarnessID  string `json:"harness_id"`
	TaskID     string `json:"task_id"`
	Mode       string `json:"mode"`
	Repetition int    `json:"repetition"`
	Cache      string `json:"cache"`
}

type AttemptSet struct {
	Version        int              `json:"version"`
	ManifestSHA256 string           `json:"manifest_sha256"`
	Runner         RunnerProvenance `json:"runner_provenance"`
	Attempts       []Attempt        `json:"attempts"`
}

type RunnerProvenance struct {
	Revision  string `json:"revision"`
	Modified  bool   `json:"modified"`
	Available bool   `json:"available"`
	Reason    string `json:"reason"`
}

type Attempt struct {
	TrialID                 string            `json:"trial_id"`
	AttemptID               string            `json:"attempt_id"`
	RetryOf                 string            `json:"retry_of"`
	Primary                 bool              `json:"primary"`
	State                   string            `json:"state"`
	Reason                  string            `json:"reason"`
	SelectedCapabilities    []string          `json:"selected_capabilities"`
	Disposition             string            `json:"disposition"`
	ResultFacts             []string          `json:"result_facts"`
	AttemptedEffects        []string          `json:"attempted_effects"`
	ActualEffects           []string          `json:"actual_effects"`
	FixtureSnapshotCaptured bool              `json:"fixture_snapshot_captured"`
	FixtureSnapshot         []FixtureSnapshot `json:"fixture_snapshot"`
	ExecutionIdentity       *HarnessSpec      `json:"execution_identity"`
	PolicyLoss              bool              `json:"policy_loss"`
	WallMillis              *int64            `json:"wall_millis"`
	Summary                 *SummaryEvidence  `json:"summary"`
	Telemetry               *Telemetry        `json:"telemetry"`
}

type SummaryEvidence struct {
	Changes      []string `json:"changes"`
	Failures     []string `json:"failures"`
	Verification []string `json:"verification"`
	Limitations  []string `json:"limitations"`
}

type FixtureSnapshot struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type FixtureEffect struct {
	Effect       string `json:"effect"`
	Path         string `json:"path"`
	BeforeSHA256 string `json:"before_sha256"`
	AfterSHA256  string `json:"after_sha256"`
}
