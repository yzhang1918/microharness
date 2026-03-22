---
template_version: 0.2.0
created_at: "2026-03-22T00:00:00+08:00"
source_type: issue
source_refs:
    - '#22'
---

# Clarify automatic review closeout and status reminders

## Goal

Clarify the controller's review-discipline contract so a future agent can tell
when step-closeout review must happen, when finalize review must start
automatically, and when routine review progression should proceed without
asking the human to micromanage.

Add a deterministic `harness status` reminder layer for missed step-closeout
review. The reminder should surface only after a completed step is missing a
qualifying clean step-closeout review or an explicit
`NO_STEP_REVIEW_NEEDED: <reason>` marker, while keeping ordinary "this slice
may now be reviewable" guidance in `next_actions` instead of heuristic
warnings.

## Scope

### In Scope

- Tighten `AGENTS.md`, `harness-execute`, and the execute references so the
  controller owns routine review progression, runs `harness status` at explicit
  checkpoints, and does not stop to ask the human for permission before
  ordinary step-closeout or finalize review.
- Update the normative specs for step-closeout review and `harness status` so
  the repository explicitly distinguishes:
  - routine review guidance in `next_actions`
  - workflow-discipline exceptions in `warnings`
  - the explicit `NO_STEP_REVIEW_NEEDED: <reason>` suppression marker in
    step-local `Review Notes`
- Teach `harness status` to warn when an already completed earlier step lacks a
  qualifying clean `step_closeout` review, while keeping the current node
  stable even if the warning is first noticed during a later step or finalize
  closeout.
- Add focused tests for the new status warnings and suppression behavior.

### Out of Scope

- Adding a hard execution gate that prevents agents from editing tracked plan
  markdown directly or marking a step done before review.
- Introducing a new command-owned step-closeout command or another new CLI
  surface just for review discipline.
- Heuristically warning during `execution/step-<n>/implement` that the current
  slice might be ready for review before a step is actually marked done.
- Reworking finalize node transitions so later discovery of a missing earlier
  step review rewinds the node back to `execution/step-<i>/review`.

## Acceptance Criteria

- [x] `AGENTS.md`, `harness-execute`, and the execute references make it clear
      that the controller must run `harness status` at routine execution
      checkpoints, automatically start step-closeout or finalize review when
      the workflow calls for it, and only pause for blockers, scope changes, or
      explicit merge approval.
- [x] The specs clearly define that a completed step is review-complete when it
      has either a clean `step_closeout` review (`delta` by default, `full`
      allowed when the slice needs a broader pass) or a
      `NO_STEP_REVIEW_NEEDED: <reason>` marker in `Review Notes`.
- [x] `harness status` keeps ordinary review prompts in `next_actions`, but
      emits `warnings` once an already completed earlier step is missing
      qualifying step-closeout review evidence; the warning remains informative
      during later-step and finalize nodes without forcing a node rollback.
- [x] Focused Go tests cover missing earlier-step review warnings, finalize-time
      warning behavior, suppression via `NO_STEP_REVIEW_NEEDED`, and a clean
      reviewed step that does not warn.

## Deferred Items

- Revisit whether missed step-closeout review should eventually become a harder
  archive or execution gate instead of a reminder-only contract.
- Consider adding a dedicated retrospective-review workflow or command if
  reminder-only status guidance proves too soft in practice.

## Work Breakdown

### Step 1: Tighten controller review-discipline guidance

- Done: [x]

#### Objective

Update the durable docs and skills so a cold reader can see when the controller
must run status, start step-closeout review, start finalize review, and avoid
routine stop-and-ask pauses.

#### Details

Fold the discovery decisions into the tracked docs instead of leaving them in
chat. The guidance should explicitly say the controller should run
`harness status` at start/resume, before marking a step done, after each review
aggregate, and before relying on finalize progression. It should also
distinguish routine review progression from real escalation conditions.

#### Expected Files

- `AGENTS.md`
- `.agents/skills/harness-execute/SKILL.md`
- `.agents/skills/harness-execute/references/step-inner-loop.md`
- `.agents/skills/harness-execute/references/review-orchestration.md`

#### Validation

- A cold reader can tell when step-closeout review must happen versus when
  finalize review must start automatically.
- The docs clearly call for routine `harness status` checkpoints instead of
  assuming the controller will remember them from chat.

#### Execution Notes

Updated `AGENTS.md`, `harness-execute`, `step-inner-loop.md`, and
`review-orchestration.md` so the controller-owned review flow is explicit:
routine `harness status` checkpoints are named directly, ordinary
step-closeout/finalize review no longer asks the human for permission, and
`NO_STEP_REVIEW_NEEDED: <reason>` is documented as the step-local exception.
Validated the wording by rereading the affected files together to make sure the
checkpoint list, review-start rules, and non-goals stay aligned.

#### Review Notes

`review-001-delta` passed clean with `docs_consistency` and `agent_ux` slots.
The reviewers agreed the controller checkpoints, routine review ownership, and
human-escalation boundaries are explicit enough for a future controller to
follow without relying on discovery chat.

### Step 2: Define the status reminder contract

- Done: [x]

#### Objective

Write the normative spec updates for missing step-closeout review reminders,
including the explicit suppression marker and the split between `next_actions`
and `warnings`.

#### Details

Capture the accepted direction precisely:
- status should not guess whether the current in-progress slice is reviewable
- status should warn only after a completed earlier step is missing review
  discipline
- later-step or finalize warnings should keep the current node stable
- a clean `step_closeout` review may be `delta` or `full`
- `Review Notes` may suppress the warning with
  `NO_STEP_REVIEW_NEEDED: <reason>`

#### Expected Files

- `docs/specs/state-model.md`
- `docs/specs/cli-contract.md`
- `docs/specs/plan-schema.md`

#### Validation

- The specs define when a completed step counts as review-complete and where
  the suppression marker belongs.
- The status contract distinguishes ordinary guidance from true reminder
  warnings without relying on hidden discovery context.

#### Execution Notes

Updated the normative specs so the repository now defines review-complete step
closeout as either a clean `step_closeout` review or an explicit
`NO_STEP_REVIEW_NEEDED: <reason>` marker. The status contract now reserves
`warnings` for recoverable ambiguity and missed-closeout reminders, keeps
ordinary review prompts in `next_actions`, and clarifies that later-step or
finalize warnings should not force a node rollback. The review-start contract
also now says `step_closeout` targets should use the tracked step title for
deterministic status matching.

#### Review Notes

`review-002-delta` passed clean with `correctness` and `docs_consistency`
slots. The reviewers agreed the specs now use compatible terminology for
review-complete step closeout, `NO_STEP_REVIEW_NEEDED`, stable-node late
warnings, and the `next_actions` versus `warnings` split.

### Step 3: Implement and test reminder warnings

- Done: [x]

#### Objective

Teach `harness status` to emit deterministic warnings and next actions for
missing earlier step-closeout review, then cover the behavior with focused Go
tests.

#### Details

The implementation should inspect completed steps that precede the current
workflow position, determine whether each one has a qualifying clean
`step_closeout` review or an explicit `NO_STEP_REVIEW_NEEDED` marker, and then
surface the earliest unresolved miss in `next_actions` plus compact summary
warnings. When the warning is first discovered during finalize, status should
remain in finalize and use warnings rather than forcing the node back to an
earlier step.

#### Expected Files

- `internal/status/service.go`
- `internal/status/service_test.go`

#### Validation

- `go test ./internal/status -count=1`
- The new tests prove:
  - no warning for a clean completed step
  - warning while working on a later step after an earlier done step missed
    step-closeout review
  - warning while already in finalize closeout
  - no warning when `Review Notes` contains
    `NO_STEP_REVIEW_NEEDED: <reason>`

#### Execution Notes

Implemented reminder-only step-closeout warning logic in
`internal/status/service.go`. Status now scans historical `step_closeout`
review artifacts, accepts either clean review evidence or
`NO_STEP_REVIEW_NEEDED: <reason>`, warns only for completed earlier steps in
later-step or finalize review/archive nodes, and prepends repair-first
`next_actions` without changing the resolved node. Added focused coverage in
`internal/status/service_test.go` for clean `full` step closeout, later-step
warnings, finalize warnings, and marker-based suppression, then validated the
slice with `go test ./internal/status -count=1`.

Finalize review then exposed two real follow-up gaps in the same slice: the
first finalize test did not actually pin historical closeout lookup, and the
reminder logic dropped away after archive into `execution/finalize/publish` and
`execution/finalize/await_merge`. Tightened the finalize assertion to prove the
clean Step 2 artifact suppresses warnings, extended reminder coverage across
all `execution/finalize/*` nodes, added archived publish/await-merge coverage,
and updated the archived CLI fixture to use explicit
`NO_STEP_REVIEW_NEEDED: ...` closeout so unrelated reopen tests keep using
review-complete data.

#### Review Notes

`review-003-delta` passed clean with `correctness` and `tests` slots. The
reviewers confirmed the reminder-only status behavior stays deterministic:
clean historical `full` step-closeout review satisfies the contract, later-step
and finalize warnings stay informative without rewinding the node, and
`NO_STEP_REVIEW_NEEDED: <reason>` suppresses the warning as specified.
Subsequent full finalize rounds `review-004-full` and `review-005-full`
surfaced one real gap each; both findings were fixed and `review-006-full`
passed clean across `correctness`, `tests`, and `docs_consistency`.

## Validation Strategy

- Run `harness plan lint` before execution starts and after any material scope
  update to this tracked plan.
- During execution, keep doc-only changes readable with direct file review and
  validate status behavior with `go test ./internal/status -count=1`.
- Before archive, run at least the focused status package tests and any broader
  Go test coverage needed by the touched files.

## Risks

- Risk: The reminder logic could misclassify older review history and create
  noisy warnings for already clean steps.
  - Mitigation: Reuse the existing structural review metadata path, accept
    either clean `delta` or clean `full` `step_closeout` review, and cover the
    no-warning path in tests.
- Risk: The docs could still leave too much room for controller interpretation
  around when to run status or start review.
  - Mitigation: Add explicit controller checkpoints and spell out that routine
    review progression is controller-owned once the plan is approved.

## Validation Summary

- `go test ./internal/status -count=1`
- `go test ./internal/cli -count=1`
- `go test ./...`
- Re-ran the focused and full suites after each finalize-review finding was
  fixed so the repaired candidate reached `review-006-full` with a green test
  baseline.

## Review Summary

- `review-001-delta` passed clean for the controller/skill wording changes.
- `review-002-delta` passed clean for the state-model, CLI-contract, and
  plan-schema updates.
- `review-003-delta` passed clean for the initial status reminder
  implementation.
- `review-004-full` found one real blocking gap: the finalize warning test did
  not actually prove historical step-closeout evidence stayed satisfied.
- `review-005-full` found one real blocking gap: reminders disappeared after
  archive into `execution/finalize/publish` and `execution/finalize/await_merge`.
- Both finalize findings were fixed, validated, and cleared by
  `review-006-full`, which passed clean with `correctness`, `tests`, and
  `docs_consistency`.

## Archive Summary

- Archived At: 2026-03-22T18:07:04+08:00
- Revision: 1
- PR: not created yet; publish evidence will record the PR URL after archive.
- Ready: controller docs/skills, the normative specs, and `harness status`
  now agree on automatic routine review progression, review-complete step
  closeout, and reminder-only handling for missed earlier reviews across later,
  finalize, publish, and await-merge states.
- Merge Handoff: run `harness archive`, commit and push the archive move plus
  tracked code/doc changes, open or update the PR, then record publish/CI/sync
  evidence before asking for merge approval.

## Outcome Summary

### Delivered

- Tightened `AGENTS.md` and the `harness-execute` skill pack so the controller
  owns routine review progression, runs `harness status` at named checkpoints,
  and does not pause to ask the human before ordinary step-closeout or finalize
  review.
- Updated `state-model.md`, `cli-contract.md`, and `plan-schema.md` so
  review-complete step closeout is explicit: a clean `step_closeout` review
  (`delta` by default, `full` allowed when needed) or
  `NO_STEP_REVIEW_NEEDED: <reason>` in `Review Notes`.
- Implemented reminder-only `harness status` warnings for earlier completed
  steps missing review-complete closeout, while keeping ordinary review prompts
  in `next_actions` and keeping the resolved node stable.
- Extended those reminders through the full finalize workflow, including
  `execution/finalize/publish` and `execution/finalize/await_merge`, so review
  debt does not disappear right before merge readiness.
- Added and repaired focused Go coverage for later-step warnings, finalize
  warnings, archived publish/await-merge reminders, clean historical full
  reviews, and `NO_STEP_REVIEW_NEEDED` suppression, then kept the broader Go
  suite green.

### Not Delivered

- A hard execution/archive gate that rejects unresolved step-closeout review
  debt instead of warning.
- A dedicated retrospective step-closeout workflow or command beyond the
  reminder-only contract landed here.

### Follow-Up Issues

- `#24` Decide whether missed step-closeout review should stay reminder-only or
  grow a stronger deterministic gate and/or retrospective closeout workflow.
