package status_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yzhang1918/superharness/internal/plan"
	"github.com/yzhang1918/superharness/internal/status"
)

func TestStatusMinimalActivePlan(t *testing.T) {
	root := t.TempDir()
	planPath := writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		return content
	})

	result := status.Service{Workdir: root}.Read()
	if !result.OK {
		t.Fatalf("expected OK status result, got %#v", result)
	}
	if result.State.PlanStatus != "active" || result.State.Lifecycle != "awaiting_plan_approval" {
		t.Fatalf("unexpected state: %#v", result.State)
	}
	if result.State.Step != "Step 1: Replace with first step title" {
		t.Fatalf("unexpected step: %#v", result.State)
	}
	if result.State.StepState != "" {
		t.Fatalf("expected no step_state outside executing, got %#v", result.State)
	}
	if result.Artifacts.PlanPath != planPath {
		t.Fatalf("unexpected plan path: %#v", result.Artifacts)
	}
}

func TestStatusReviewInProgress(t *testing.T) {
	root := t.TempDir()
	planPath := writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: executing")
		content = replaceOnce(t, content, "- Status: pending", "- Status: in_progress")
		return content
	})
	writeCurrentPlan(t, root, "docs/plans/active/2026-03-18-status-plan.md")
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"active_review_round": map[string]any{
			"round_id":   "round-1",
			"kind":       "delta",
			"aggregated": false,
		},
	})

	result := status.Service{Workdir: root}.Read()
	if !result.OK || result.State.StepState != "reviewing" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.Artifacts.PlanPath != planPath || result.Artifacts.ReviewRoundID != "round-1" {
		t.Fatalf("unexpected artifacts: %#v", result.Artifacts)
	}
}

func TestStatusWaitingCI(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: executing")
		content = replaceOnce(t, content, "- Status: pending", "- Status: in_progress")
		return content
	})
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"latest_ci": map[string]any{
			"snapshot_id": "ci-1",
			"status":      "pending",
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.StepState != "waiting_ci" {
		t.Fatalf("unexpected step state: %#v", result.State)
	}
}

func TestStatusFixRequiredAfterAggregatedReviewFailure(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: executing")
		content = replaceOnce(t, content, "- Status: pending", "- Status: in_progress")
		return content
	})
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"active_review_round": map[string]any{
			"round_id":   "review-005-delta",
			"kind":       "delta",
			"aggregated": true,
			"decision":   "changes_requested",
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.StepState != "fix_required" {
		t.Fatalf("unexpected step state: %#v", result.State)
	}
	if !strings.Contains(result.Summary, "requested changes") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
	if len(result.NextAction) < 2 {
		t.Fatalf("expected multiple next actions, got %#v", result.NextAction)
	}
	if !strings.Contains(result.NextAction[0].Description, "review-005-delta") {
		t.Fatalf("expected round-specific guidance, got %#v", result.NextAction)
	}
	if result.NextAction[1].Command == nil || *result.NextAction[1].Command != "harness review start --spec <path>" {
		t.Fatalf("unexpected second next action: %#v", result.NextAction)
	}
}

func TestStatusUsesAggregateArtifactForLegacyReviewDecision(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: executing")
		content = replaceOnce(t, content, "- Status: pending", "- Status: in_progress")
		return content
	})
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"active_review_round": map[string]any{
			"round_id":   "review-004-full",
			"kind":       "full",
			"aggregated": true,
		},
	})
	writeAggregate(t, root, "2026-03-18-status-plan", "review-004-full", map[string]any{
		"decision": "changes_requested",
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.StepState != "fix_required" {
		t.Fatalf("unexpected step state: %#v", result.State)
	}
	if !strings.Contains(result.Summary, "review-004-full") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
}

func TestStatusFixRequiredBeatsCloseoutGuidance(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: executing")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "PENDING_UNTIL_ARCHIVE", "Ready.")
		return content
	})
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"active_review_round": map[string]any{
			"round_id":   "review-007-delta",
			"kind":       "delta",
			"aggregated": true,
			"decision":   "changes_requested",
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.StepState != "fix_required" {
		t.Fatalf("unexpected step state: %#v", result.State)
	}
	if len(result.NextAction) < 2 {
		t.Fatalf("expected fix-required next actions, got %#v", result.NextAction)
	}
	if strings.Contains(result.NextAction[0].Description, "Validation, Review, Archive, and Outcome summaries") {
		t.Fatalf("expected fix guidance to win over closeout guidance, got %#v", result.NextAction)
	}
	if !strings.Contains(result.NextAction[0].Description, "review-007-delta") {
		t.Fatalf("expected round-specific fix guidance, got %#v", result.NextAction)
	}
}

func TestStatusUnknownAggregatedReviewDecisionBlocksCloseoutGuidance(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: executing")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "PENDING_UNTIL_ARCHIVE", "Ready.")
		return content
	})
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"active_review_round": map[string]any{
			"round_id":   "review-008-delta",
			"kind":       "delta",
			"aggregated": true,
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.StepState != "fix_required" {
		t.Fatalf("unexpected step state: %#v", result.State)
	}
	if !strings.Contains(result.Summary, "could not be recovered") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
	if len(result.NextAction) < 2 {
		t.Fatalf("expected conservative next actions, got %#v", result.NextAction)
	}
	if !strings.Contains(result.NextAction[0].Description, "Recover or rerun review-008-delta") {
		t.Fatalf("unexpected first next action: %#v", result.NextAction)
	}
	if len(result.Warnings) == 0 || !strings.Contains(result.Warnings[0], "could not be recovered") {
		t.Fatalf("expected recovery warning, got %#v", result.Warnings)
	}
}

func TestStatusResolvingConflicts(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: executing")
		content = replaceOnce(t, content, "- Status: pending", "- Status: in_progress")
		return content
	})
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"sync": map[string]any{
			"freshness": "stale",
			"conflicts": true,
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.StepState != "resolving_conflicts" {
		t.Fatalf("unexpected step state: %#v", result.State)
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("expected remote freshness warning")
	}
}

func TestStatusReadyForArchive(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: executing")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "## Validation Summary\n\nPENDING_UNTIL_ARCHIVE", "## Validation Summary\n\nValidated the implementation.")
		content = stringsReplaceAll(content, "## Review Summary\n\nPENDING_UNTIL_ARCHIVE", "## Review Summary\n\nNo blocking review findings remain.")
		content = stringsReplaceAll(content, "## Archive Summary\n\nPENDING_UNTIL_ARCHIVE", "## Archive Summary\n\n- PR: NONE\n- Ready: The candidate satisfies the acceptance criteria and is ready for merge approval.\n- Merge Handoff: Commit and push the archive move before treating this candidate as awaiting merge approval.")
		content = stringsReplaceAll(content, "### Delivered\n\nPENDING_UNTIL_ARCHIVE", "### Delivered\n\nDelivered the planned slice.")
		content = stringsReplaceAll(content, "### Not Delivered\n\nPENDING_UNTIL_ARCHIVE", "### Not Delivered\n\nNONE.")
		return content
	})
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"active_review_round": map[string]any{
			"round_id":   "review-001-full",
			"kind":       "full",
			"aggregated": true,
			"decision":   "pass",
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.StepState != "ready_for_archive" {
		t.Fatalf("unexpected step state: %#v", result.State)
	}
	if len(result.NextAction) == 0 || result.NextAction[0].Command != nil {
		t.Fatalf("expected archive-ready guidance, got %#v", result.NextAction)
	}
	if len(result.NextAction) < 2 || result.NextAction[1].Command == nil || *result.NextAction[1].Command != "harness archive" {
		t.Fatalf("expected archive command guidance, got %#v", result.NextAction)
	}
	if !strings.Contains(result.Summary, "ready to archive") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
}

func TestStatusCloseoutBeforeArchive(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: executing")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		return content
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.Step != "" {
		t.Fatalf("expected no current step, got %#v", result.State)
	}
	if len(result.Blockers) == 0 {
		t.Fatalf("expected archive blockers, got %#v", result)
	}
	if len(result.NextAction) == 0 || !strings.Contains(result.NextAction[0].Description, "Fix the archive blockers") {
		t.Fatalf("unexpected next actions: %#v", result.NextAction)
	}
	if !strings.Contains(result.Summary, "archive blocker") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
}

func TestStatusArchivedPlanNeedsPublishHandoff(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "status: active", "status: archived")
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: awaiting_merge_approval")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "PENDING_UNTIL_ARCHIVE", "Ready.")
		return content
	})
	writeCurrentPlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md")

	result := status.Service{Workdir: root}.Read()
	if !result.OK || result.State.PlanStatus != "archived" || result.State.Lifecycle != "awaiting_merge_approval" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.State.StepState != "" {
		t.Fatalf("expected no step_state for archived plan, got %#v", result.State)
	}
	if result.State.HandoffState != "pending_publish" {
		t.Fatalf("expected pending_publish handoff state, got %#v", result.State)
	}
	if !strings.Contains(result.Summary, "archived locally") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
	if len(result.NextAction) == 0 || !strings.Contains(result.NextAction[0].Description, "open or update the PR") {
		t.Fatalf("unexpected next actions: %#v", result.NextAction)
	}
}

func TestStatusArchivedPlanWaitingForPostArchiveCI(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "status: active", "status: archived")
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: awaiting_merge_approval")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "PENDING_UNTIL_ARCHIVE", "Ready.")
		return content
	})
	writeCurrentPlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md")
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"latest_publish": map[string]any{
			"attempt_id": "publish-001",
			"pr_url":     "https://github.com/yzhang1918/superharness/pull/13",
		},
		"latest_ci": map[string]any{
			"snapshot_id": "ci-001",
			"status":      "pending",
		},
		"sync": map[string]any{
			"freshness": "fresh",
			"conflicts": false,
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.HandoffState != "waiting_post_archive_ci" {
		t.Fatalf("expected waiting_post_archive_ci handoff state, got %#v", result.State)
	}
	if !strings.Contains(result.Summary, "post-archive CI") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
	if result.Artifacts == nil || result.Artifacts.PRURL != "https://github.com/yzhang1918/superharness/pull/13" {
		t.Fatalf("unexpected artifacts: %#v", result.Artifacts)
	}
}

func TestStatusArchivedPlanNeedsCIEvidenceBeforeMergeApproval(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "status: active", "status: archived")
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: awaiting_merge_approval")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "PENDING_UNTIL_ARCHIVE", "Ready.")
		return content
	})
	writeCurrentPlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md")
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"latest_publish": map[string]any{
			"attempt_id": "publish-001",
			"pr_url":     "https://github.com/yzhang1918/superharness/pull/13",
		},
		"sync": map[string]any{
			"freshness": "fresh",
			"conflicts": false,
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.HandoffState != "waiting_post_archive_ci" {
		t.Fatalf("expected waiting_post_archive_ci without CI evidence, got %#v", result.State)
	}
	if !strings.Contains(result.Summary, "post-archive CI") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
}

func TestStatusArchivedPlanReadyForMergeApproval(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "status: active", "status: archived")
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: awaiting_merge_approval")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "PENDING_UNTIL_ARCHIVE", "Ready.")
		return content
	})
	writeCurrentPlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md")
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"latest_publish": map[string]any{
			"attempt_id": "publish-001",
			"pr_url":     "https://github.com/yzhang1918/superharness/pull/13",
		},
		"latest_ci": map[string]any{
			"snapshot_id": "ci-001",
			"status":      "success",
		},
		"sync": map[string]any{
			"freshness": "fresh",
			"conflicts": false,
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.HandoffState != "ready_for_merge_approval" {
		t.Fatalf("expected ready_for_merge_approval handoff state, got %#v", result.State)
	}
	if !strings.Contains(result.Summary, "ready to wait for merge approval") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
	if len(result.NextAction) == 0 || !strings.Contains(result.NextAction[0].Description, "Wait for merge approval") {
		t.Fatalf("unexpected next actions: %#v", result.NextAction)
	}
}

func TestStatusArchivedPlanRequiresSyncEvidenceBeforeMergeApproval(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "status: active", "status: archived")
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: awaiting_merge_approval")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "PENDING_UNTIL_ARCHIVE", "Ready.")
		return content
	})
	writeCurrentPlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md")
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"latest_publish": map[string]any{
			"attempt_id": "publish-001",
			"pr_url":     "https://github.com/yzhang1918/superharness/pull/13",
		},
		"latest_ci": map[string]any{
			"snapshot_id": "ci-001",
			"status":      "success",
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.HandoffState != "followup_required" {
		t.Fatalf("expected followup_required without sync evidence, got %#v", result.State)
	}
	if !strings.Contains(result.Summary, "needs follow-up") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
}

func TestStatusArchivedPlanRequiresFollowupWhenSyncIsNotClean(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "status: active", "status: archived")
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: awaiting_merge_approval")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "PENDING_UNTIL_ARCHIVE", "Ready.")
		return content
	})
	writeCurrentPlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md")
	writeState(t, root, "2026-03-18-status-plan", map[string]any{
		"latest_publish": map[string]any{
			"attempt_id": "publish-001",
			"pr_url":     "https://github.com/yzhang1918/superharness/pull/13",
		},
		"latest_ci": map[string]any{
			"snapshot_id": "ci-001",
			"status":      "success",
		},
		"sync": map[string]any{
			"freshness": "stale",
			"conflicts": false,
		},
	})

	result := status.Service{Workdir: root}.Read()
	if result.State.HandoffState != "followup_required" {
		t.Fatalf("expected followup_required handoff state, got %#v", result.State)
	}
	if !strings.Contains(result.Summary, "needs follow-up") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
	if len(result.NextAction) == 0 || !strings.Contains(result.NextAction[0].Description, "Refresh remote state") {
		t.Fatalf("expected remote follow-up guidance, got %#v", result.NextAction)
	}
}

func TestStatusSurfacesArchiveBlockersBeforeArchive(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/active/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: executing")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "## Deferred Items\n\n- None.\n", "## Deferred Items\n\n- Deferred cleanup still needs a durable handoff.\n")
		content = stringsReplaceAll(content, "## Validation Summary\n\nPENDING_UNTIL_ARCHIVE", "## Validation Summary\n\nValidated the implementation.")
		content = stringsReplaceAll(content, "## Review Summary\n\nPENDING_UNTIL_ARCHIVE", "## Review Summary\n\nNo blocking review findings remain.")
		content = stringsReplaceAll(content, "## Archive Summary\n\nPENDING_UNTIL_ARCHIVE", "## Archive Summary\n\n- Ready: Closeout is nearly done.\n- Merge Handoff: Commit and push the archive move before merge approval.")
		content = stringsReplaceAll(content, "### Delivered\n\nPENDING_UNTIL_ARCHIVE", "### Delivered\n\nDelivered the planned slice.")
		content = stringsReplaceAll(content, "### Not Delivered\n\nPENDING_UNTIL_ARCHIVE", "### Not Delivered\n\nDeferred cleanup remains.")
		return content
	})

	result := status.Service{Workdir: root}.Read()
	if !result.OK {
		t.Fatalf("expected OK status result, got %#v", result)
	}
	if result.State.StepState == "ready_for_archive" {
		t.Fatalf("expected blockers to prevent ready_for_archive, got %#v", result.State)
	}
	if len(result.Blockers) < 2 {
		t.Fatalf("expected multiple archive blockers, got %#v", result.Blockers)
	}
	if !hasBlockerPath(result.Blockers, "section.Archive Summary") {
		t.Fatalf("expected Archive Summary blocker, got %#v", result.Blockers)
	}
	if !hasBlockerPath(result.Blockers, "section.Outcome Summary.Follow-Up Issues") {
		t.Fatalf("expected follow-up blocker, got %#v", result.Blockers)
	}
	if !strings.Contains(result.Summary, "archive blocker") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
	if len(result.NextAction) == 0 || !strings.Contains(result.NextAction[0].Description, "Fix the archive blockers") {
		t.Fatalf("unexpected next actions: %#v", result.NextAction)
	}
}

func TestStatusReportsIdleAfterLandMarker(t *testing.T) {
	root := t.TempDir()
	writePlan(t, root, "docs/plans/archived/2026-03-18-status-plan.md", func(content string) string {
		content = replaceOnce(t, content, "status: active", "status: archived")
		content = replaceOnce(t, content, "lifecycle: awaiting_plan_approval", "lifecycle: awaiting_merge_approval")
		content = stringsReplaceAll(content, "- Status: pending", "- Status: completed")
		content = stringsReplaceAll(content, "- [ ]", "- [x]")
		content = stringsReplaceAll(content, "PENDING_STEP_EXECUTION", "Done.")
		content = stringsReplaceAll(content, "PENDING_STEP_REVIEW", "Reviewed.")
		content = stringsReplaceAll(content, "PENDING_UNTIL_ARCHIVE", "Ready.")
		return content
	})
	writeCurrentPlanPayload(t, root, map[string]any{
		"last_landed_plan_path": "docs/plans/archived/2026-03-18-status-plan.md",
		"last_landed_at":        "2026-03-19T12:00:00Z",
	})

	result := status.Service{Workdir: root}.Read()
	if !result.OK {
		t.Fatalf("expected OK idle-after-land result, got %#v", result)
	}
	if result.State.WorktreeState != "idle_after_land" {
		t.Fatalf("unexpected worktree state: %#v", result.State)
	}
	if result.State.PlanStatus != "" || result.State.Lifecycle != "" {
		t.Fatalf("expected no active current plan state, got %#v", result.State)
	}
	if result.Artifacts == nil || result.Artifacts.LastLandedPlanPath != "docs/plans/archived/2026-03-18-status-plan.md" {
		t.Fatalf("unexpected artifacts: %#v", result.Artifacts)
	}
	if !strings.Contains(result.Summary, "most recent landed candidate") {
		t.Fatalf("unexpected summary: %q", result.Summary)
	}
}

func writePlan(t *testing.T, root, relPath string, mutate func(string) string) string {
	t.Helper()
	rendered, err := plan.RenderTemplate(plan.TemplateOptions{
		Title:      "Status Plan",
		Timestamp:  time.Date(2026, 3, 18, 10, 0, 0, 0, time.FixedZone("CST", 8*60*60)),
		SourceType: "direct_request",
	})
	if err != nil {
		t.Fatalf("RenderTemplate: %v", err)
	}
	content := mutate(rendered)
	path := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	return path
}

func writeCurrentPlan(t *testing.T, root, relPath string) {
	t.Helper()
	writeCurrentPlanPayload(t, root, map[string]any{"plan_path": relPath})
}

func writeCurrentPlanPayload(t *testing.T, root string, payloadMap map[string]any) {
	t.Helper()
	dir := filepath.Join(root, ".local", "harness")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir current-plan dir: %v", err)
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		t.Fatalf("marshal current-plan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "current-plan.json"), payload, 0o644); err != nil {
		t.Fatalf("write current-plan: %v", err)
	}
}

func writeState(t *testing.T, root, planStem string, payload map[string]any) {
	t.Helper()
	dir := filepath.Join(root, ".local", "harness", "plans", planStem)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "state.json"), data, 0o644); err != nil {
		t.Fatalf("write state: %v", err)
	}
}

func writeAggregate(t *testing.T, root, planStem, roundID string, payload map[string]any) {
	t.Helper()
	dir := filepath.Join(root, ".local", "harness", "plans", planStem, "reviews", roundID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir aggregate dir: %v", err)
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal aggregate: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "aggregate.json"), data, 0o644); err != nil {
		t.Fatalf("write aggregate: %v", err)
	}
}

func replaceOnce(t *testing.T, content, old, new string) string {
	t.Helper()
	updated := stringsReplaceOnce(content, old, new)
	if updated == content {
		t.Fatalf("expected replacement %q -> %q", old, new)
	}
	return updated
}

func stringsReplaceOnce(content, old, new string) string {
	return strings.Replace(content, old, new, 1)
}

func stringsReplaceAll(content, old, new string) string {
	return strings.ReplaceAll(content, old, new)
}

func hasBlockerPath(blockers []status.StatusError, path string) bool {
	for _, blocker := range blockers {
		if blocker.Path == path {
			return true
		}
	}
	return false
}
