// Command argcheck reports whether a numbered argument has syntactically
// transparent validity under the deterministic entailment gate.
//
// The local backend checks numbered_argument_syntax, then the entailment gate.
// --prod asks a remote validate endpoint (MCP) instead; --both runs each
// and flags divergence.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bits-notes/argcheck/validation"
)

const (
	exitValid     = 0
	exitInvalid   = 1
	exitError     = 2
	exitDivergent = 3
)

// check mirrors a check in the remote validate endpoint response.
type check struct {
	Passed  bool           `json:"passed"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// verdict mirrors the remote validate endpoint response shape plus provenance.
type verdict struct {
	Valid      bool                         `json:"valid"`
	Argument   *validation.NumberedArgument `json:"argument,omitempty"`
	Syntax     check                        `json:"syntax"`
	Entailment *check                       `json:"entailment,omitempty"`
	Catalog    string                       `json:"entailment_catalog"`
	// Inert lists conclusions whose full citation set fails but a proper
	// subset passes a known form. Local backend only; advisory.
	Inert   []validation.InertCitationReport `json:"inert_citations,omitempty"`
	Backend string                           `json:"backend"`
	Version string                           `json:"version"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("argcheck", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOut := flags.Bool("json", false, "print machine-readable JSON")
	prod := flags.Bool("prod", false, "check against the remote MCP endpoint configured via ARGCHECK_REMOTE_URL and ARGCHECK_REMOTE_TOOL")
	both := flags.Bool("both", false, "check locally and against the remote endpoint; exit 3 if they disagree")
	version := flags.Bool("version", false, "print the catalog and build version and exit")
	forms := flags.Bool("forms", false, "list the accepted inference forms and exit")
	quiet := flags.Bool("quiet", false, "print only the verdict word (VALID or INVALID; DIVERGENT with --both on disagreement)")
	flags.Usage = func() {
		fmt.Fprintln(stderr, `usage: argcheck [--json|--quiet] [--prod|--both] [file|-]

Reads a numbered argument (file, or stdin) in the gate's format:
  1. Premise statement.
  2. Another premise.
  3. Therefore, conclusion. [1,2]

Clauses must match word for word (after case and whitespace normalization).
This is deliberate and will not change: if two wordings mean the same thing,
state that as a premise.

Exit 0 valid, 1 invalid, 2 error, 3 local/remote divergence (--both).`)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return exitError
	}
	if *version {
		fmt.Fprintf(stdout, "entailment gate catalog %s (%s)\n", validation.EntailmentGateCatalogVersion, localVersion())
		return exitValid
	}
	if *forms {
		fmt.Fprintf(stdout, "entailment gate catalog %s (%s)\n", validation.EntailmentGateCatalogVersion, localVersion())
		for _, name := range validation.EntailmentGateForms() {
			fmt.Fprintln(stdout, "  "+name)
		}
		return exitValid
	}
	if flags.NArg() > 1 {
		flags.Usage()
		return exitError
	}

	var remoteURL, remoteTool string
	if *prod || *both {
		remoteURL = strings.TrimSpace(os.Getenv("ARGCHECK_REMOTE_URL"))
		if remoteURL == "" {
			fmt.Fprintln(stderr, "argcheck: --prod and --both require ARGCHECK_REMOTE_URL to be set")
			return exitError
		}
		remoteTool = strings.TrimSpace(os.Getenv("ARGCHECK_REMOTE_TOOL"))
		if remoteTool == "" {
			fmt.Fprintln(stderr, "argcheck: --prod and --both require ARGCHECK_REMOTE_TOOL to be set")
			return exitError
		}
	}

	input, err := readInput(flags.Arg(0), stdin)
	if err != nil {
		fmt.Fprintln(stderr, "argcheck:", err)
		return exitError
	}
	argument := strings.TrimSpace(input)
	if argument == "" {
		fmt.Fprintln(stderr, "argcheck: argument is empty")
		return exitError
	}

	var verdicts []verdict
	if !*prod || *both {
		verdicts = append(verdicts, checkLocal(argument))
	}
	if *prod || *both {
		remote, err := checkProd(remoteURL, remoteTool, argument)
		if err != nil {
			fmt.Fprintln(stderr, "argcheck: remote check failed:", err)
			return exitError
		}
		verdicts = append(verdicts, remote)
	}

	divergent := len(verdicts) == 2 && !sameVerdict(verdicts[0], verdicts[1])
	if *jsonOut {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		var payload any = verdicts[0]
		if len(verdicts) == 2 {
			payload = map[string]any{"local": verdicts[0], "prod": verdicts[1], "divergent": divergent}
		}
		if err := encoder.Encode(payload); err != nil {
			fmt.Fprintln(stderr, "argcheck:", err)
			return exitError
		}
	} else if *quiet {
		switch {
		case divergent:
			fmt.Fprintln(stdout, "DIVERGENT")
		case verdicts[len(verdicts)-1].Valid:
			fmt.Fprintln(stdout, "VALID")
		default:
			fmt.Fprintln(stdout, "INVALID")
		}
	} else {
		// The verdict line always comes first, so `head -1` is the verdict.
		for index, v := range verdicts {
			if index > 0 {
				fmt.Fprintln(stdout)
			}
			printVerdict(stdout, v)
		}
		if divergent {
			fmt.Fprintln(stdout, "\nDIVERGENT: local and remote disagree (likely a different catalog or commit).")
		}
		if strings.Contains(argument, "| ") {
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, argument)
		}
	}

	switch {
	case divergent:
		return exitDivergent
	case verdicts[len(verdicts)-1].Valid:
		return exitValid
	default:
		return exitInvalid
	}
}

func readInput(path string, stdin io.Reader) (string, error) {
	if path == "" || path == "-" {
		data, err := io.ReadAll(stdin)
		return string(data), err
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

// checkLocal runs syntax validation before checking transparent entailment.
func checkLocal(argument string) verdict {
	ctx := context.Background()
	syntax := validation.Validate(ctx, validation.BuiltinSpec(validation.BuiltinNumberedArgumentSyntax), argument)
	result := verdict{
		Syntax:  check{Passed: syntax.Passed, Message: syntax.Message, Details: syntax.Details},
		Catalog: validation.EntailmentGateCatalogVersion,
		Backend: "local",
		Version: localVersion(),
	}
	if !syntax.Passed {
		return result
	}
	if parsed, r := validation.ParseScopedNumberedArgumentAllowingNoConclusions(argument); r.Passed && parsed.HasSubproofs() {
		result.Argument = &parsed
	}
	_, entailment := validation.CheckEntailmentGate(argument)
	result.Entailment = &check{Passed: entailment.Passed, Message: entailment.Message, Details: entailment.Details}
	result.Valid = entailment.Passed
	if !result.Valid {
		result.Inert = inertCitations(argument)
	}
	return result
}

func inertCitations(argument string) []validation.InertCitationReport {
	parsed, parse := validation.ParseScopedNumberedArgumentAllowingNoConclusions(argument)
	if !parse.Passed {
		return nil
	}
	expanded, _, err := validation.ExpandNumberedArgumentBridges(parsed)
	if err != nil {
		return nil
	}
	return validation.NumberedArgumentInertCitations(expanded)
}

// commit is set by install-argcheck.sh via -ldflags "-X main.commit=...".
var commit string

func localVersion() string {
	if commit != "" {
		return "argcheck@" + commit
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "argcheck@unknown"
	}
	revision, modified := "", false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	if revision == "" {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return "argcheck@" + info.Main.Version
		}
		return "argcheck@unknown (build with go build, not go run)"
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	if modified {
		revision += "-dirty"
	}
	return "argcheck@" + revision
}

func checkProd(remoteURL, remoteTool, argument string) (verdict, error) {
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      remoteTool,
			"arguments": map[string]string{"argument": argument},
		},
	})
	if err != nil {
		return verdict{}, err
	}
	request, err := http.NewRequest(http.MethodPost, remoteURL, bytes.NewReader(body))
	if err != nil {
		return verdict{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return verdict{}, fmt.Errorf("%w (is the remote validate endpoint reachable?)", err)
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return verdict{}, err
	}
	if response.StatusCode != http.StatusOK {
		return verdict{}, fmt.Errorf("remote endpoint returned %s: %s", response.Status, strings.TrimSpace(string(raw)))
	}

	var envelope struct {
		Result *struct {
			IsError           bool            `json:"isError"`
			StructuredContent json.RawMessage `json:"structuredContent"`
			Content           []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return verdict{}, fmt.Errorf("unreadable remote response: %w", err)
	}
	if envelope.Error != nil {
		return verdict{}, errors.New(envelope.Error.Message)
	}
	if envelope.Result == nil {
		return verdict{}, errors.New("remote response has no result")
	}
	if envelope.Result.IsError || len(envelope.Result.StructuredContent) == 0 {
		texts := make([]string, 0, len(envelope.Result.Content))
		for _, item := range envelope.Result.Content {
			texts = append(texts, item.Text)
		}
		return verdict{}, fmt.Errorf("remote validate endpoint error: %s (token may lack the required scope)", strings.Join(texts, " "))
	}
	var result verdict
	if err := json.Unmarshal(envelope.Result.StructuredContent, &result); err != nil {
		return verdict{}, fmt.Errorf("unreadable validation result: %w", err)
	}
	result.Backend = "prod"
	result.Version = "remote validate endpoint (live)"
	return result, nil
}

func sameVerdict(a, b verdict) bool {
	if a.Valid != b.Valid || a.Syntax.Passed != b.Syntax.Passed || a.Catalog != b.Catalog {
		return false
	}
	return a.Syntax.Message == b.Syntax.Message && entailmentMessage(a) == entailmentMessage(b)
}

func entailmentMessage(v verdict) string {
	if v.Entailment == nil {
		return ""
	}
	return v.Entailment.Message
}

// stepReport is decoded from Details["steps"], which arrives either as typed
// structs (local) or as JSON maps (prod).
type stepReport struct {
	Step        int      `json:"step"`
	Depth       int      `json:"depth,omitempty"`
	Ranges      [][2]int `json:"ranges,omitempty"`
	Citations   []int    `json:"citations"`
	Passed      bool     `json:"passed"`
	Form        string   `json:"form"`
	NearestForm string   `json:"nearest_form"`
	Mismatch    string   `json:"mismatch"`
}

func stepsOf(c *check) []stepReport {
	if c == nil || c.Details["steps"] == nil {
		return nil
	}
	raw, err := json.Marshal(c.Details["steps"])
	if err != nil {
		return nil
	}
	var steps []stepReport
	if json.Unmarshal(raw, &steps) != nil {
		return nil
	}
	return steps
}

func printVerdict(w io.Writer, v verdict) {
	status := "INVALID"
	if v.Valid {
		status = "VALID"
	}
	fmt.Fprintf(w, "%s  (gate catalog %s; %s %s)\n", status, v.Catalog, v.Backend, v.Version)
	if !v.Syntax.Passed {
		fmt.Fprintf(w, "syntax: FAIL: %s\n", v.Syntax.Message)
		return
	}
	fmt.Fprintln(w, "syntax: ok")
	steps := stepsOf(v.Entailment)
	if len(steps) == 0 && v.Entailment != nil {
		fmt.Fprintf(w, "entailment: %s\n", v.Entailment.Message)
		return
	}
	inert := make(map[int]validation.InertCitationReport, len(v.Inert))
	for _, report := range v.Inert {
		inert[report.Step] = report
	}
	for _, step := range steps {
		cites := joinInts(step.Citations)
		for _, r := range step.Ranges {
			if cites != "" {
				cites += ","
			}
			cites += fmt.Sprintf("%d-%d", r[0], r[1])
		}
		if step.Passed {
			fmt.Fprintf(w, "step %d [%s]: ok: %s\n", step.Step, cites, step.Form)
			continue
		}
		fmt.Fprintf(w, "step %d [%s]: FAIL: nearest form %s: %s\n", step.Step, cites, step.NearestForm, step.Mismatch)
		if report, ok := inert[step.Step]; ok {
			fmt.Fprintf(w, "  hint: %s passes on [%s]; citation(s) [%s] do no work\n",
				report.Form, joinInts(report.UsedCitations), joinInts(report.InertCitations))
		}
	}
}

func joinInts(values []int) string {
	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = fmt.Sprint(value)
	}
	return strings.Join(parts, ",")
}
