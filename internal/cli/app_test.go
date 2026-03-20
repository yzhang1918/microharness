package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yzhang1918/superharness/internal/cli"
	"github.com/yzhang1918/superharness/internal/plan"
	"github.com/yzhang1918/superharness/internal/runstate"
	"github.com/yzhang1918/superharness/internal/status"
)

func TestPlanTemplateWritesOutputFile(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	app := cli.New(stdout, stderr)
	app.Now = func() time.Time {
		return time.Date(2026, 3, 17, 15, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	}

	outputPath := filepath.Join(t.TempDir(), "docs/plans/active/2026-03-17-test-plan.md")
	exitCode := app.Run([]string{
		"plan", "template",
		"--title", "CLI Generated Plan",
		"--output", outputPath,
		"--source-type", "issue",
		"--source-ref", "#42",
	})
	if exitCode != 0 {
		t.Fatalf("unexpected exit code %d, stderr=%s", exitCode, stderr.String())
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !bytes.Contains(data, []byte("# CLI Generated Plan")) {
		t.Fatalf("generated file missing title:\n%s", data)
	}
}

func TestPlanLintCommandReturnsJSON(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	app := cli.New(stdout, stderr)
	app.Now = func() time.Time {
		return time.Date(2026, 3, 17, 15, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	}

	outputPath := filepath.Join(t.TempDir(), "docs/plans/active/2026-03-17-test-plan.md")
	if exitCode := app.Run([]string{
		"plan", "template",
		"--title", "CLI Generated Plan",
		"--output", outputPath,
	}); exitCode != 0 {
		t.Fatalf("template command failed with %d: %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	exitCode := app.Run([]string{"plan", "lint", outputPath})
	if exitCode != 0 {
		t.Fatalf("lint command failed with %d: %s", exitCode, stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON lint output: %v\n%s", err, stdout.String())
	}
	if ok, _ := payload["ok"].(bool); !ok {
		t.Fatalf("expected ok=true, got %v", payload["ok"])
	}
}

func TestPlanTemplateHelpExitsZero(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	app := cli.New(stdout, stderr)

	exitCode := app.Run([]string{"plan", "template", "--help"})
	if exitCode != 0 {
		t.Fatalf("expected help exit code 0, got %d", exitCode)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("Usage: harness plan template")) {
		t.Fatalf("expected help text, got %s", stderr.String())
	}
}

func TestPlanLintHelpExitsZero(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	app := cli.New(stdout, stderr)

	exitCode := app.Run([]string{"plan", "lint", "--help"})
	if exitCode != 0 {
		t.Fatalf("expected help exit code 0, got %d", exitCode)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("Usage: harness plan lint")) {
		t.Fatalf("expected help text, got %s", stderr.String())
	}
}

func TestStatusCommandReturnsJSON(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	app := cli.New(stdout, stderr)
	root := t.TempDir()
	app.Getwd = func() (string, error) { return root, nil }
	app.Now = func() time.Time {
		return time.Date(2026, 3, 18, 15, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	}

	outputPath := filepath.Join(root, "docs/plans/active/2026-03-18-test-plan.md")
	if exitCode := app.Run([]string{
		"plan", "template",
		"--title", "CLI Generated Plan",
		"--output", outputPath,
	}); exitCode != 0 {
		t.Fatalf("template command failed with %d: %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()

	exitCode := app.Run([]string{"status"})
	if exitCode != 0 {
		t.Fatalf("status command failed with %d: %s", exitCode, stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON status output: %v\n%s", err, stdout.String())
	}
	if payload["command"] != "status" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestLandRecordCommandReturnsJSON(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	app := cli.New(stdout, stderr)
	root := t.TempDir()
	app.Getwd = func() (string, error) { return root, nil }
	app.Now = func() time.Time {
		return time.Date(2026, 3, 18, 6, 0, 0, 0, time.UTC)
	}

	writeArchivedPlanForCLI(t, root, "docs/plans/archived/2026-03-18-landed-plan.md")
	if _, err := runstate.SaveCurrentPlan(root, "docs/plans/archived/2026-03-18-landed-plan.md"); err != nil {
		t.Fatalf("save current plan: %v", err)
	}

	exitCode := app.Run([]string{"land", "record"})
	if exitCode != 0 {
		t.Fatalf("land record command failed with %d: %s", exitCode, stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON land record output: %v\n%s", err, stdout.String())
	}
	if payload["command"] != "land record" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestLandRecordCommandRejectsActivePlanWithoutWritingLandedMarker(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	app := cli.New(stdout, stderr)
	root := t.TempDir()
	app.Getwd = func() (string, error) { return root, nil }
	app.Now = func() time.Time {
		return time.Date(2026, 3, 18, 6, 0, 0, 0, time.UTC)
	}

	outputPath := filepath.Join(root, "docs/plans/active/2026-03-18-active-plan.md")
	if exitCode := app.Run([]string{
		"plan", "template",
		"--title", "CLI Active Plan",
		"--output", outputPath,
	}); exitCode != 0 {
		t.Fatalf("template command failed with %d: %s", exitCode, stderr.String())
	}
	if _, err := runstate.SaveCurrentPlan(root, "docs/plans/active/2026-03-18-active-plan.md"); err != nil {
		t.Fatalf("save current plan: %v", err)
	}

	stdout.Reset()
	stderr.Reset()

	exitCode := app.Run([]string{"land", "record"})
	if exitCode != 1 {
		t.Fatalf("expected land record failure exit code, got %d", exitCode)
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON land record output: %v\n%s", err, stdout.String())
	}
	if payload["command"] != "land record" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if ok, _ := payload["ok"].(bool); ok {
		t.Fatalf("expected ok=false, got %#v", payload)
	}

	current, err := runstate.LoadCurrentPlan(root)
	if err != nil {
		t.Fatalf("load current plan: %v", err)
	}
	if current == nil || current.PlanPath != "docs/plans/active/2026-03-18-active-plan.md" {
		t.Fatalf("expected active current plan to remain, got %#v", current)
	}
	if current.LastLandedPlanPath != "" || current.LastLandedAt != "" {
		t.Fatalf("expected no landed marker on failed command, got %#v", current)
	}

	statusResult := status.Service{Workdir: root}.Read()
	if !statusResult.OK {
		t.Fatalf("expected active-plan status after failed land record, got %#v", statusResult)
	}
	if statusResult.State.WorktreeState == "idle_after_land" {
		t.Fatalf("failed land record should not switch status to idle_after_land: %#v", statusResult)
	}
}

func writeArchivedPlanForCLI(t *testing.T, root, relPath string) string {
	t.Helper()
	rendered, err := plan.RenderTemplate(plan.TemplateOptions{
		Title:      "CLI Landed Plan",
		Timestamp:  time.Date(2026, 3, 18, 2, 0, 0, 0, time.UTC),
		SourceType: "direct_request",
	})
	if err != nil {
		t.Fatalf("RenderTemplate: %v", err)
	}
	content := rendered
	content = bytes.NewBufferString(content).String()
	content = replaceCLI(content, "status: active", "status: archived")
	content = replaceCLI(content, "lifecycle: awaiting_plan_approval", "lifecycle: awaiting_merge_approval")
	content = replaceCLIAll(content, "- Status: pending", "- Status: completed")
	content = replaceCLIAll(content, "- [ ]", "- [x]")
	content = replaceCLIAll(content, "PENDING_STEP_EXECUTION", "Done.")
	content = replaceCLIAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
	content = replaceCLI(content, "## Validation Summary\n\nPENDING_UNTIL_ARCHIVE", "## Validation Summary\n\nValidated the slice.")
	content = replaceCLI(content, "## Review Summary\n\nPENDING_UNTIL_ARCHIVE", "## Review Summary\n\nFull review passed.")
	content = replaceCLI(content, "## Archive Summary\n\nPENDING_UNTIL_ARCHIVE", "## Archive Summary\n\n- Archived At: 2026-03-18T02:00:00Z\n- Revision: 1\n- PR: NONE\n- Ready: Ready for merge approval.\n- Merge Handoff: Commit and push before merge approval.")
	content = replaceCLI(content, "### Delivered\n\nPENDING_UNTIL_ARCHIVE", "### Delivered\n\nDelivered the slice.")
	content = replaceCLI(content, "### Not Delivered\n\nPENDING_UNTIL_ARCHIVE", "### Not Delivered\n\nNONE.")
	path := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write archived plan: %v", err)
	}
	return path
}

func replaceCLI(content, old, new string) string {
	tuned := bytes.Replace([]byte(content), []byte(old), []byte(new), 1)
	return string(tuned)
}

func replaceCLIAll(content, old, new string) string {
	tuned := bytes.ReplaceAll([]byte(content), []byte(old), []byte(new))
	return string(tuned)
}
