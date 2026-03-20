package lifecycle_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yzhang1918/superharness/internal/lifecycle"
	"github.com/yzhang1918/superharness/internal/plan"
	"github.com/yzhang1918/superharness/internal/runstate"
	"github.com/yzhang1918/superharness/internal/status"
)

func TestArchiveMovesPlanAndUpdatesPointers(t *testing.T) {
	root := t.TempDir()
	activeRelPath := "docs/plans/active/2026-03-18-archive-smoke.md"
	activePath := writeActiveArchiveCandidate(t, root, activeRelPath)
	if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", &runstate.State{
		PlanPath: activeRelPath,
		PlanStem: "2026-03-18-archive-smoke",
		ActiveReviewRound: &runstate.ReviewRound{
			RoundID:    "review-001-full",
			Kind:       "full",
			Aggregated: true,
			Decision:   "pass",
		},
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}

	svc := lifecycle.Service{
		Workdir: root,
		Now: func() time.Time {
			return time.Date(2026, 3, 18, 2, 0, 0, 0, time.UTC)
		},
	}
	result := svc.Archive()
	if !result.OK {
		t.Fatalf("expected archive success, got %#v", result)
	}

	archivedPath := filepath.Join(root, "docs/plans/archived/2026-03-18-archive-smoke.md")
	if _, err := os.Stat(archivedPath); err != nil {
		t.Fatalf("archived path missing: %v", err)
	}
	if _, err := os.Stat(activePath); !os.IsNotExist(err) {
		t.Fatalf("expected active path to be removed, got %v", err)
	}
	if lint := plan.LintFile(archivedPath); !lint.OK {
		t.Fatalf("archived plan should lint, got %#v", lint)
	}
	current, err := runstate.LoadCurrentPlan(root)
	if err != nil {
		t.Fatalf("load current-plan: %v", err)
	}
	if current == nil || current.PlanPath != "docs/plans/archived/2026-03-18-archive-smoke.md" {
		t.Fatalf("unexpected current plan: %#v", current)
	}
	state, _, err := runstate.LoadState(root, "2026-03-18-archive-smoke")
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if state == nil || state.PlanPath != "docs/plans/archived/2026-03-18-archive-smoke.md" {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestArchiveRejectsMissingArchiveSummaryFields(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "docs/plans/active/2026-03-18-archive-smoke.md")
	content := buildActiveArchiveCandidate(t)
	content = strings.Replace(content, "- PR: NONE\n", "", 1)
	writeFile(t, path, content)
	if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", &runstate.State{
		PlanPath: "docs/plans/active/2026-03-18-archive-smoke.md",
		PlanStem: "2026-03-18-archive-smoke",
		ActiveReviewRound: &runstate.ReviewRound{
			RoundID:    "review-001-full",
			Kind:       "full",
			Aggregated: true,
			Decision:   "pass",
		},
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}

	svc := lifecycle.Service{Workdir: root}
	result := svc.Archive()
	if result.OK {
		t.Fatalf("expected archive failure, got %#v", result)
	}
	assertErrorPath(t, result.Errors, "section.Archive Summary")
}

func TestArchivePreflightFailureLeavesPlanAndPointersUntouched(t *testing.T) {
	root := t.TempDir()
	activeRelPath := "docs/plans/active/2026-03-18-archive-smoke.md"
	activePath := filepath.Join(root, activeRelPath)
	content := buildActiveArchiveCandidate(t)
	content = strings.Replace(content, "## Deferred Items\n\n- None.\n", "## Deferred Items\n\n- Deferred follow-up still needs to be written down.\n", 1)
	content = strings.Replace(content, "- PR: NONE\n", "", 1)
	writeFile(t, activePath, content)
	if _, err := runstate.SaveCurrentPlan(root, activeRelPath); err != nil {
		t.Fatalf("save current plan: %v", err)
	}
	if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", &runstate.State{
		PlanPath: activeRelPath,
		PlanStem: "2026-03-18-archive-smoke",
		ActiveReviewRound: &runstate.ReviewRound{
			RoundID:    "review-001-full",
			Kind:       "full",
			Aggregated: true,
			Decision:   "pass",
		},
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}

	result := lifecycle.Service{Workdir: root}.Archive()
	if result.OK {
		t.Fatalf("expected archive failure, got %#v", result)
	}
	assertErrorPath(t, result.Errors, "section.Archive Summary")
	assertErrorPath(t, result.Errors, "section.Outcome Summary.Follow-Up Issues")

	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("expected active plan to remain after failed archive, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs/plans/archived/2026-03-18-archive-smoke.md")); !os.IsNotExist(err) {
		t.Fatalf("expected no archived plan to be written, got %v", err)
	}
	current, err := runstate.LoadCurrentPlan(root)
	if err != nil {
		t.Fatalf("load current plan: %v", err)
	}
	if current == nil || current.PlanPath != activeRelPath {
		t.Fatalf("expected current plan pointer to remain on active plan, got %#v", current)
	}
	state, _, err := runstate.LoadState(root, "2026-03-18-archive-smoke")
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if state == nil || state.PlanPath != activeRelPath {
		t.Fatalf("expected state pointer to remain on active plan, got %#v", state)
	}
}

func TestArchiveRollsBackWhenCurrentPlanWriteFails(t *testing.T) {
	root := t.TempDir()
	activeRelPath := "docs/plans/active/2026-03-18-archive-smoke.md"
	activePath := writeActiveArchiveCandidate(t, root, activeRelPath)
	if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", &runstate.State{
		PlanPath: activeRelPath,
		PlanStem: "2026-03-18-archive-smoke",
		ActiveReviewRound: &runstate.ReviewRound{
			RoundID:    "review-001-full",
			Kind:       "full",
			Aggregated: true,
			Decision:   "pass",
		},
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}
	currentPlanAsDir := filepath.Join(root, ".local", "harness", "current-plan.json")
	if err := os.MkdirAll(currentPlanAsDir, 0o755); err != nil {
		t.Fatalf("mkdir current-plan dir: %v", err)
	}

	result := lifecycle.Service{Workdir: root}.Archive()
	if result.OK {
		t.Fatalf("expected archive failure, got %#v", result)
	}
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("expected active plan to remain after rollback, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "docs/plans/archived/2026-03-18-archive-smoke.md")); !os.IsNotExist(err) {
		t.Fatalf("expected archived target to be removed on rollback, got %v", err)
	}
	state, _, err := runstate.LoadState(root, "2026-03-18-archive-smoke")
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if state == nil || state.PlanPath != activeRelPath {
		t.Fatalf("expected state to roll back to active path, got %#v", state)
	}
}

func TestArchiveRejectsUnresolvedLocalState(t *testing.T) {
	testCases := []struct {
		name       string
		state      *runstate.State
		errorPath  string
		errorMatch string
	}{
		{
			name: "active review round",
			state: &runstate.State{
				ActiveReviewRound: &runstate.ReviewRound{RoundID: "review-001-full", Kind: "full", Aggregated: false},
			},
			errorPath:  "state.active_review_round",
			errorMatch: "aggregate or clear",
		},
		{
			name: "non-green ci",
			state: &runstate.State{
				LatestCI: &runstate.CIState{SnapshotID: "ci-1", Status: "pending"},
			},
			errorPath:  "state.latest_ci",
			errorMatch: "not archive-ready",
		},
		{
			name: "stale sync",
			state: &runstate.State{
				Sync: &runstate.SyncState{Freshness: "stale"},
			},
			errorPath:  "state.sync",
			errorMatch: "refresh remote state",
		},
		{
			name: "merge conflicts",
			state: &runstate.State{
				Sync: &runstate.SyncState{Freshness: "fresh", Conflicts: true},
			},
			errorPath:  "state.sync",
			errorMatch: "resolve merge conflicts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			activeRelPath := "docs/plans/active/2026-03-18-archive-smoke.md"
			writeActiveArchiveCandidate(t, root, activeRelPath)
			tc.state.PlanPath = activeRelPath
			tc.state.PlanStem = "2026-03-18-archive-smoke"
			if tc.state.ActiveReviewRound == nil && tc.errorPath != "state.active_review_round" {
				tc.state.ActiveReviewRound = &runstate.ReviewRound{
					RoundID:    "review-001-full",
					Kind:       "full",
					Aggregated: true,
					Decision:   "pass",
				}
			}
			if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", tc.state); err != nil {
				t.Fatalf("save state: %v", err)
			}

			result := lifecycle.Service{Workdir: root}.Archive()
			if result.OK {
				t.Fatalf("expected archive failure, got %#v", result)
			}
			assertErrorPath(t, result.Errors, tc.errorPath)
			assertErrorContains(t, result.Errors, tc.errorPath, tc.errorMatch)
		})
	}
}

func TestArchiveRequiresPassingReviewForRevisionOne(t *testing.T) {
	testCases := []struct {
		name       string
		state      *runstate.State
		errorMatch string
	}{
		{
			name:       "missing review",
			state:      &runstate.State{},
			errorMatch: "passing full review",
		},
		{
			name: "passing delta is not enough",
			state: &runstate.State{
				ActiveReviewRound: &runstate.ReviewRound{
					RoundID:    "review-001-delta",
					Kind:       "delta",
					Aggregated: true,
					Decision:   "pass",
				},
			},
			errorMatch: "passing full review",
		},
		{
			name: "failed full review still blocks",
			state: &runstate.State{
				ActiveReviewRound: &runstate.ReviewRound{
					RoundID:    "review-001-full",
					Kind:       "full",
					Aggregated: true,
					Decision:   "changes_requested",
				},
			},
			errorMatch: "not archive-ready",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			activeRelPath := "docs/plans/active/2026-03-18-archive-smoke.md"
			writeActiveArchiveCandidate(t, root, activeRelPath)
			tc.state.PlanPath = activeRelPath
			tc.state.PlanStem = "2026-03-18-archive-smoke"
			if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", tc.state); err != nil {
				t.Fatalf("save state: %v", err)
			}

			result := lifecycle.Service{Workdir: root}.Archive()
			if result.OK {
				t.Fatalf("expected archive failure, got %#v", result)
			}
			assertErrorPath(t, result.Errors, "state.active_review_round")
			assertErrorContains(t, result.Errors, "state.active_review_round", tc.errorMatch)
		})
	}
}

func TestArchiveAllowsPassingDeltaReviewForReopenedRevision(t *testing.T) {
	root := t.TempDir()
	activeRelPath := "docs/plans/active/2026-03-18-archive-smoke.md"
	activePath := writeActiveArchiveCandidate(t, root, activeRelPath)
	data, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("read active plan: %v", err)
	}
	updated := strings.Replace(string(data), "revision: 1", "revision: 2", 1)
	writeFile(t, activePath, updated)

	if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", &runstate.State{
		PlanPath: activeRelPath,
		PlanStem: "2026-03-18-archive-smoke",
		ActiveReviewRound: &runstate.ReviewRound{
			RoundID:    "review-002-delta",
			Kind:       "delta",
			Aggregated: true,
			Decision:   "pass",
		},
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}

	result := lifecycle.Service{
		Workdir: root,
		Now: func() time.Time {
			return time.Date(2026, 3, 18, 4, 0, 0, 0, time.UTC)
		},
	}.Archive()
	if !result.OK {
		t.Fatalf("expected archive success for reopened revision, got %#v", result)
	}
}

func TestArchiveUsesAggregateArtifactForLegacyReviewDecision(t *testing.T) {
	root := t.TempDir()
	activeRelPath := "docs/plans/active/2026-03-18-archive-smoke.md"
	writeActiveArchiveCandidate(t, root, activeRelPath)
	if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", &runstate.State{
		PlanPath: activeRelPath,
		PlanStem: "2026-03-18-archive-smoke",
		ActiveReviewRound: &runstate.ReviewRound{
			RoundID:    "review-001-full",
			Kind:       "full",
			Aggregated: true,
		},
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}
	writeAggregateArtifact(t, root, "2026-03-18-archive-smoke", "review-001-full", map[string]any{
		"decision": "pass",
	})

	result := lifecycle.Service{
		Workdir: root,
		Now: func() time.Time {
			return time.Date(2026, 3, 18, 4, 30, 0, 0, time.UTC)
		},
	}.Archive()
	if !result.OK {
		t.Fatalf("expected archive success for legacy review decision, got %#v", result)
	}
}

func TestReopenMovesArchivedPlanBackToActiveAndResetsSummaries(t *testing.T) {
	root := t.TempDir()
	writeActiveArchiveCandidate(t, root, "docs/plans/active/2026-03-18-archive-smoke.md")
	if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", &runstate.State{
		PlanPath: "docs/plans/active/2026-03-18-archive-smoke.md",
		PlanStem: "2026-03-18-archive-smoke",
		ActiveReviewRound: &runstate.ReviewRound{
			RoundID:    "review-001-full",
			Kind:       "full",
			Aggregated: true,
			Decision:   "pass",
		},
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}

	svc := lifecycle.Service{
		Workdir: root,
		Now: func() time.Time {
			return time.Date(2026, 3, 18, 2, 0, 0, 0, time.UTC)
		},
	}
	archive := svc.Archive()
	if !archive.OK {
		t.Fatalf("archive failed: %#v", archive)
	}

	svc.Now = func() time.Time {
		return time.Date(2026, 3, 18, 3, 0, 0, 0, time.UTC)
	}
	reopen := svc.Reopen()
	if !reopen.OK {
		t.Fatalf("reopen failed: %#v", reopen)
	}

	activePath := filepath.Join(root, "docs/plans/active/2026-03-18-archive-smoke.md")
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("reopened active path missing: %v", err)
	}
	if lint := plan.LintFile(activePath); !lint.OK {
		t.Fatalf("reopened active plan should lint, got %#v", lint)
	}
	data, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("read reopened plan: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "revision: 2") {
		t.Fatalf("expected revision bump, got:\n%s", text)
	}
	if !strings.Contains(text, "## Archive Summary\n\nPENDING_UNTIL_ARCHIVE") {
		t.Fatalf("expected Archive Summary reset, got:\n%s", text)
	}
	if !strings.Contains(text, "### Follow-Up Issues\n\nNONE") {
		t.Fatalf("expected follow-up reset, got:\n%s", text)
	}
}

func TestReopenClearsStaleCIAndSyncSignals(t *testing.T) {
	root := t.TempDir()
	activeRelPath := "docs/plans/active/2026-03-18-archive-smoke.md"
	writeActiveArchiveCandidate(t, root, activeRelPath)
	if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", &runstate.State{
		PlanPath: activeRelPath,
		PlanStem: "2026-03-18-archive-smoke",
		ActiveReviewRound: &runstate.ReviewRound{
			RoundID:    "review-001-full",
			Kind:       "full",
			Aggregated: true,
			Decision:   "pass",
		},
		LatestCI: &runstate.CIState{SnapshotID: "ci-1", Status: "success"},
		Sync:     &runstate.SyncState{Freshness: "fresh", Conflicts: false},
	}); err != nil {
		t.Fatalf("save initial state: %v", err)
	}

	svc := lifecycle.Service{
		Workdir: root,
		Now: func() time.Time {
			return time.Date(2026, 3, 18, 2, 0, 0, 0, time.UTC)
		},
	}
	archive := svc.Archive()
	if !archive.OK {
		t.Fatalf("archive failed: %#v", archive)
	}

	if _, err := runstate.SaveState(root, "2026-03-18-archive-smoke", &runstate.State{
		PlanPath:          "docs/plans/archived/2026-03-18-archive-smoke.md",
		PlanStem:          "2026-03-18-archive-smoke",
		ActiveReviewRound: &runstate.ReviewRound{RoundID: "review-001-full", Kind: "full", Aggregated: true, Decision: "pass"},
		LatestCI:          &runstate.CIState{SnapshotID: "ci-2", Status: "failed"},
		Sync:              &runstate.SyncState{Freshness: "stale", Conflicts: true},
	}); err != nil {
		t.Fatalf("save archived state: %v", err)
	}

	svc.Now = func() time.Time {
		return time.Date(2026, 3, 18, 3, 0, 0, 0, time.UTC)
	}
	reopen := svc.Reopen()
	if !reopen.OK {
		t.Fatalf("reopen failed: %#v", reopen)
	}

	state, _, err := runstate.LoadState(root, "2026-03-18-archive-smoke")
	if err != nil {
		t.Fatalf("load reopened state: %v", err)
	}
	if state == nil {
		t.Fatalf("expected reopened state")
	}
	if state.ActiveReviewRound != nil || state.LatestCI != nil || state.Sync != nil {
		t.Fatalf("expected reopened state to clear stale review/ci/sync signals, got %#v", state)
	}
}

func TestRecordLandedWritesIdleMarkerForStatus(t *testing.T) {
	root := t.TempDir()
	writeArchivedLandedPlan(t, root, "docs/plans/archived/2026-03-18-landed-plan.md")
	if _, err := runstate.SaveCurrentPlan(root, "docs/plans/archived/2026-03-18-landed-plan.md"); err != nil {
		t.Fatalf("save current plan: %v", err)
	}

	result := lifecycle.Service{
		Workdir: root,
		Now: func() time.Time {
			return time.Date(2026, 3, 18, 6, 0, 0, 0, time.UTC)
		},
	}.RecordLanded()
	if !result.OK {
		t.Fatalf("expected landed record success, got %#v", result)
	}

	current, err := runstate.LoadCurrentPlan(root)
	if err != nil {
		t.Fatalf("load current plan: %v", err)
	}
	if current == nil || current.PlanPath != "" || current.LastLandedPlanPath != "docs/plans/archived/2026-03-18-landed-plan.md" {
		t.Fatalf("unexpected current plan marker: %#v", current)
	}

	statusResult := status.Service{Workdir: root}.Read()
	if !statusResult.OK {
		t.Fatalf("expected idle-after-land status, got %#v", statusResult)
	}
	if statusResult.State.WorktreeState != "idle_after_land" {
		t.Fatalf("unexpected worktree state: %#v", statusResult.State)
	}
}

func writeActiveArchiveCandidate(t *testing.T, root, relPath string) string {
	t.Helper()
	path := filepath.Join(root, relPath)
	writeFile(t, path, buildActiveArchiveCandidate(t))
	return path
}

func writeArchivedLandedPlan(t *testing.T, root, relPath string) string {
	t.Helper()
	path := filepath.Join(root, relPath)
	rendered, err := plan.RenderTemplate(plan.TemplateOptions{
		Title:      "Landed Plan",
		Timestamp:  time.Date(2026, 3, 18, 2, 0, 0, 0, time.UTC),
		SourceType: "direct_request",
	})
	if err != nil {
		t.Fatalf("render template: %v", err)
	}
	rendered = strings.Replace(rendered, "status: active", "status: archived", 1)
	rendered = strings.Replace(rendered, "lifecycle: awaiting_plan_approval", "lifecycle: awaiting_merge_approval", 1)
	rendered = strings.ReplaceAll(rendered, "- Status: pending", "- Status: completed")
	rendered = strings.ReplaceAll(rendered, "- [ ]", "- [x]")
	rendered = strings.ReplaceAll(rendered, "PENDING_STEP_EXECUTION", "Done.")
	rendered = strings.ReplaceAll(rendered, "PENDING_STEP_REVIEW", "Reviewed.")
	rendered = strings.Replace(rendered, "## Validation Summary\n\nPENDING_UNTIL_ARCHIVE", "## Validation Summary\n\nValidated the slice.", 1)
	rendered = strings.Replace(rendered, "## Review Summary\n\nPENDING_UNTIL_ARCHIVE", "## Review Summary\n\nFull review passed.", 1)
	rendered = strings.Replace(rendered, "## Archive Summary\n\nPENDING_UNTIL_ARCHIVE", "## Archive Summary\n\n- Archived At: 2026-03-18T02:00:00Z\n- Revision: 1\n- PR: NONE\n- Ready: Ready for merge approval.\n- Merge Handoff: Commit and push before merge approval.", 1)
	rendered = strings.Replace(rendered, "### Delivered\n\nPENDING_UNTIL_ARCHIVE", "### Delivered\n\nDelivered the slice.", 1)
	rendered = strings.Replace(rendered, "### Not Delivered\n\nPENDING_UNTIL_ARCHIVE", "### Not Delivered\n\nNONE.", 1)
	writeFile(t, path, rendered)
	return path
}

func buildActiveArchiveCandidate(t *testing.T) string {
	t.Helper()
	rendered, err := plan.RenderTemplate(plan.TemplateOptions{
		Title:      "Archive Smoke",
		Timestamp:  time.Date(2026, 3, 18, 2, 0, 0, 0, time.UTC),
		SourceType: "direct_request",
	})
	if err != nil {
		t.Fatalf("render template: %v", err)
	}
	rendered = strings.Replace(rendered, "lifecycle: awaiting_plan_approval", "lifecycle: executing", 1)
	rendered = strings.ReplaceAll(rendered, "- Status: pending", "- Status: completed")
	rendered = strings.ReplaceAll(rendered, "- [ ]", "- [x]")
	rendered = strings.ReplaceAll(rendered, "PENDING_STEP_EXECUTION", "Completed execution notes.")
	rendered = strings.ReplaceAll(rendered, "PENDING_STEP_REVIEW", "Completed review notes.")
	rendered = strings.Replace(rendered, "## Validation Summary\n\nPENDING_UNTIL_ARCHIVE", "## Validation Summary\n\nValidated the implementation and command surfaces.", 1)
	rendered = strings.Replace(rendered, "## Review Summary\n\nPENDING_UNTIL_ARCHIVE", "## Review Summary\n\nNo unresolved blocking review findings remain.", 1)
	rendered = strings.Replace(rendered, "## Archive Summary\n\nPENDING_UNTIL_ARCHIVE", "## Archive Summary\n\n- PR: NONE\n- Ready: The candidate satisfies the acceptance criteria and is ready for merge approval.\n- Merge Handoff: Commit and push the archive move before treating this candidate as awaiting merge approval.", 1)
	rendered = strings.Replace(rendered, "### Delivered\n\nPENDING_UNTIL_ARCHIVE", "### Delivered\n\nDelivered the planned CLI slice.", 1)
	rendered = strings.Replace(rendered, "### Not Delivered\n\nPENDING_UNTIL_ARCHIVE", "### Not Delivered\n\nNONE.", 1)
	return rendered
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func writeAggregateArtifact(t *testing.T, root, planStem, roundID string, payload map[string]any) {
	t.Helper()
	dir := filepath.Join(root, ".local", "harness", "plans", planStem, "reviews", roundID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir aggregate dir: %v", err)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal aggregate payload: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "aggregate.json"), data, 0o644); err != nil {
		t.Fatalf("write aggregate: %v", err)
	}
}

func assertErrorPath(t *testing.T, issues []lifecycle.CommandError, path string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Path == path {
			return
		}
	}
	t.Fatalf("expected error for %s, got %#v", path, issues)
}

func assertErrorContains(t *testing.T, issues []lifecycle.CommandError, path, fragment string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Path == path && strings.Contains(issue.Message, fragment) {
			return
		}
	}
	t.Fatalf("expected error for %s containing %q, got %#v", path, fragment, issues)
}
