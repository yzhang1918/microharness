---
template_version: 0.2.0
created_at: "2026-04-02T23:12:05+08:00"
source_type: issue
source_refs:
    - '#95'
---

# Hook review round data into the harness UI

## Goal

Replace the `Review` page placeholder with a real read-only workbench for the
active plan's review rounds. The page should help a human understand current
review state, compare rounds, inspect each reviewer slot's assigned task and
submitted result, and judge what needs attention next without falling back to
raw artifact spelunking.

This slice should follow the same product boundary as the live `Status` page:
the Go backend builds a read-only view model from existing harness-owned local
artifacts, and the frontend renders that model. It must not change CLI command
contracts, mutate review artifacts, or introduce new write-side indexing just
to support the UI.

## Scope

### In Scope

- Add a read-only review resource for `harness ui` that only reads review
  rounds under the current active plan.
- Build a `Review` round browser with:
  - a round list in the navigation pane
  - an overview-first detail pane for the selected round
  - reviewer-focused content that combines each slot's assigned instructions
    with its submitted summary and findings
- Treat `manifest`, `ledger`, `aggregate`, and raw submission artifacts as
  supporting evidence rather than the page's primary organizing principle.
- Handle incomplete, in-progress, missing, or malformed review artifacts
  conservatively so the UI stays usable without pretending the data is clean.
- Add focused Go/unit coverage plus browser automation using the
  [$playwright](/Users/yaozhang/.codex/skills/playwright/SKILL.md) skill.
- Include manual interactive Playwright verification with screenshots so the
  final UI slice is checked for readability, density, and overall aesthetics,
  not just data correctness.

### Out of Scope

- Any change to `harness review start`, `harness review submit`,
  `harness review aggregate`, or their CLI JSON contracts.
- Any new write-side artifact, background indexing, or timeline/event mutation
  added solely for the `Review` page.
- Reading review rounds outside the current active plan.
- UI-triggered review actions, command execution, or local-state mutation.
- Turning the page into a raw artifact browser where manifest/ledger/aggregate
  tabs dominate the default experience.

## Acceptance Criteria

- [x] `Review` no longer renders as a WIP placeholder and instead loads review
      rounds for the current active plan only.
- [x] The page presents a round browser shape where the second column lists
      available rounds with high-signal metadata such as round id, kind,
      title/target, timestamp, and current decision or waiting state.
- [x] The selected round's detail pane defaults to an overview-first view.
- [x] The overview surface makes review kind, title/target, revision, and
      timing easy to read.
- [x] The overview surface shows aggregate decision when present, or a
      conservative in-progress / incomplete status when not.
- [x] The overview surface shows reviewer submission progress.
- [x] The overview surface highlights high-signal blocking and non-blocking
      findings when aggregate data exists.
- [x] The detail pane supports reviewer-focused inspection where each reviewer
      slot combines the task/instructions that reviewer received with the
      submission summary, findings, and locations that reviewer returned.
- [x] Reviewer slots with no submission yet render a clear empty or pending
      state.
- [x] Supporting artifact views remain available for manifest, ledger,
      aggregate, and raw submissions, but they are secondary to the overview
      and reviewer content.
- [x] In-progress rounds with only partial artifacts still appear with a clear
      waiting status.
- [x] Missing or malformed artifacts produce warnings or degraded sections
      rather than crashing the page or showing false-clean status.
- [x] The page remains read-only and never rewrites local state to "repair"
      damaged data.
- [x] The implementation does not change review CLI contracts or write-side
      logic beyond read-only UI wiring.
- [x] Focused Go coverage and Playwright automation validate review data
      loading, round selection, degraded-state rendering, and reviewer detail
      presentation.
- [x] Before closeout, the implementation is also exercised interactively via
      Playwright with captured screenshots and a quick aesthetic pass on
      spacing, hierarchy, and legibility.

## Deferred Items

- Reading review history across archived or non-active plans.
- Deep file-anchor navigation from review finding locations into `Diff` or
  `Files`.
- Editing or triggering review actions from the UI.
- Rich side-by-side raw artifact diffing beyond the supporting evidence tabs.

## Work Breakdown

### Step 1: Define the read-only review resource and degraded-state rules

- Done: [x]

#### Objective

Lock the backend read-model boundary for active-plan review rounds and document
how incomplete or damaged artifacts should degrade in the UI.

#### Details

Follow the `status` pattern instead of the `timeline` write-side pattern. The
backend should detect the current active plan, enumerate only that plan's
review rounds from existing local artifacts, and assemble a UI-facing read
model without changing any review command contract or mutation behavior.

This step should make the degraded-state rules explicit: rounds may be waiting
for submissions, waiting for aggregation, missing aggregate data, missing
submission files, or partially malformed because a human or external tool
damaged local state. The resource should surface those conditions as warnings
and conservative status labels rather than failing the whole page whenever one
artifact is imperfect.

#### Expected Files

- `internal/ui/server.go`
- new read-only review resource file(s) under `internal/`
- `internal/ui/server_test.go`

#### Validation

- The review resource loads rounds only from the current active plan.
- A cold reader can tell from the plan and resource shape that no review CLI
  contracts or write-side logic need to change.
- Tests cover at least one clean round, one in-progress round, and one damaged
  or partially missing round.

#### Execution Notes

Added a new read-only `internal/reviewui` service plus `/api/review` wiring in
the UI server. The read model only inspects review rounds under the current
active plan and stays on the `status` pattern: no command-contract changes and
no new write-side runtime artifacts. The service now degrades conservatively
for missing review directories, malformed JSON artifacts, missing submissions,
and aggregate gaps while still returning the rest of the round browser data.

Focused backend coverage now locks the core states requested during discovery:
clean rounds, in-progress rounds, degraded rounds, archived-plan empty state,
and `/api/review` endpoint integration.

#### Review Notes

NO_STEP_REVIEW_NEEDED: Backend read-model work was intentionally landed as part
of one integrated review-UI slice with the frontend round browser and browser
validation, so a step-local review would be artificially narrower than the
real user-visible change.

### Step 2: Replace the Review placeholder with the round browser UI

- Done: [x]

#### Objective

Ship the `Review` page as an overview-first round browser that centers review
content rather than raw artifact management.

#### Details

The second column should list rounds for the active plan, with enough metadata
to quickly distinguish step/finalize rounds, newer versus older rounds, and
clean versus waiting/problem states. The selected round's detail pane should
lead with a compact overview of review state, then let the user inspect each
reviewer slot in a combined pane that pairs the assigned instructions with the
submitted summary/findings.

Raw artifact access should still exist, but as supporting evidence. The
frontend should avoid making `manifest`, `ledger`, and `aggregate` tabs feel
like the main event. The visual result should stay aligned with the existing
workbench shell: dense, calm, technical, and readable next to the already-live
`Status` and `Timeline` pages.

#### Expected Files

- `web/src/main.tsx`
- `web/src/styles.css`
- `internal/ui/static/*`

#### Validation

- The page defaults to the most relevant available round and renders a stable
  empty state when the active plan has no review rounds.
- Reviewer panes clearly show task plus result in one place and still make
  pending reviewers understandable.
- Supporting artifact views are present but visually secondary.

#### Execution Notes

Replaced the `Review` placeholder with a dedicated review workspace instead of
forcing it through the generic sidebar layout. The page now uses the accepted
product shape: round browser in the second column, overview-first detail pane
in the third column, combined reviewer task/result panes, and secondary raw
artifact tabs for manifest/ledger/aggregate/submissions. A small follow-up
visual pass trimmed the artifact inspector height so raw JSON stays secondary
to the review content itself.

#### Review Notes

NO_STEP_REVIEW_NEEDED: The round-browser UI depends directly on the Step 1
read model and Step 3 browser validation, so the meaningful review boundary is
the integrated slice rather than this frontend-only checkpoint.

### Step 3: Lock behavior and polish with automated and interactive browser validation

- Done: [x]

#### Objective

Prove the review UI with focused automation and a final interactive visual pass
instead of relying on static reasoning alone.

#### Details

Add browser coverage for round-list loading, round switching, reviewer detail
inspection, and degraded states such as missing aggregate or pending
submissions. Use the
[$playwright](/Users/yaozhang/.codex/skills/playwright/SKILL.md) skill for
browser automation, and keep the plan explicit that the controller should also
run an interactive local UI session, click through the page, capture
screenshots, and make any necessary aesthetic refinements before declaring the
slice complete.

This step is not just about "does it render"; it is also about whether the UI
looks intentional. The final pass should check layout balance, hierarchy,
density, and whether the review content feels more important than the support
artifacts.

#### Expected Files

- Playwright validation artifacts under `.local/`
- `scripts/ui-playwright-smoke`
- browser-focused test files under `web/` or existing UI test locations

#### Validation

- Automated Playwright coverage exercises clean and degraded review states.
- The controller runs an interactive Playwright session against the local UI.
- Screenshots exist for final inspection, and any obvious visual issues found
  during that pass are fixed before closeout.

#### Execution Notes

Added `scripts/ui-playwright-review-smoke` for review-specific browser
coverage and updated the existing `scripts/ui-playwright-smoke` expectations
now that `/review` is no longer a WIP placeholder. Validation includes:

- `pnpm --dir web check`
- `pnpm --dir web build`
- `scripts/ui-playwright-review-smoke`
- `scripts/ui-playwright-smoke`
- interactive Playwright inspection of the live review page with headed
  browsing plus captured screenshots under `output/playwright/manual-review-visual/`
- `go test ./...`

#### Review Notes

NO_STEP_REVIEW_NEEDED: Browser automation and visual polish only make sense
after the integrated review workspace exists, so this closeout is folded into
the later finalize review of the whole slice.

## Validation Strategy

- Lint the tracked plan with `harness plan lint`.
- Add focused Go tests for the review read model and UI server endpoint.
- Add or extend Playwright automation for review round browsing, reviewer
  detail inspection, and degraded artifact states.
- Run an interactive Playwright session against the implemented UI, capture
  screenshots, and use that pass to tune aesthetics as needed.

## Risks

- Risk: Reading raw local review artifacts directly could make the page brittle
  when artifacts are incomplete or damaged.
  - Mitigation: Define explicit conservative degradation rules in the read
    model and test them directly.
- Risk: A raw-artifact-first UI could technically work while still failing the
  product goal of helping humans steer review work.
  - Mitigation: Keep the default experience overview-first and reviewer-first,
    with supporting artifacts clearly secondary.
- Risk: Browser automation could verify data rendering but miss visual
  awkwardness or hierarchy problems.
  - Mitigation: Require an interactive Playwright pass with screenshots before
    closeout, not just automated checks.

## Validation Summary

- `harness plan lint docs/plans/active/2026-04-02-hook-review-round-data-into-harness-ui.md`
  passed after the closeout summaries and deferred-scope handoff were updated.
- `pnpm --dir web check` and `go test ./...` passed after the final
  supporting-artifact summary polish and review-smoke helper hardening.
- `scripts/ui-playwright-review-smoke` passed with populated review rounds,
  empty active-plan state, damaged manifest/ledger/aggregate artifacts, and a
  malformed submission artifact that now renders its validation summary in the
  supporting-evidence pane.
- `scripts/ui-playwright-smoke` passed with the live `/review` page in the
  main browser path and the archived-plan review copy still covered.
- Interactive headed Playwright inspection against the live worktree review UI
  confirmed the reviewer result pane remains more prominent than the assigned
  task pane and that supporting artifacts stay visually secondary after adding
  artifact summaries. Captured screenshots live under
  `output/playwright/manual-review-visual-r6/`.

## Review Summary

- `review-001-full` through `review-010-full` progressively tightened the
  slice around semantic artifact validation, degraded-round rendering,
  provenance display, reviewer-pane hierarchy, README validation guidance, and
  malformed-submission browser coverage.
- `review-010-full` was the final changes-requested round; it identified the
  last two blockers: the populated review smoke needed to be part of the
  documented browser validation path, and the review smoke fixture still
  lacked malformed submission coverage.
- Those findings were fixed by documenting
  `scripts/ui-playwright-review-smoke`, extending the review smoke fixture to
  exercise malformed submissions, hardening the smoke selector fallback under
  `set -euo pipefail`, and surfacing supporting-artifact summaries directly in
  the UI.
- `review-011-full` then passed cleanly across `correctness`, `tests`, and
  `agent-ux` with zero blocking and zero non-blocking findings.

## Archive Summary

- Archived At: 2026-04-03T02:36:48+08:00
- Revision: 1
- PR: NONE. The branch has not been pushed or opened as a PR yet.
- Ready: Acceptance criteria are satisfied, the `Review` page is now backed by
  a read-only active-plan review resource, degraded review rounds render
  conservatively, Playwright automation covers both general UI and
  review-specific paths, and the latest finalize review passed cleanly in
  `review-011-full`.
- Merge Handoff: Run `harness archive`, commit the tracked plan move plus the
  implementation and closeout summaries, push branch
  `codex/review-ui-round-browser`, open or refresh the PR, and record
  publish/CI/sync evidence until `harness status` reaches
  `execution/finalize/await_merge`.

## Outcome Summary

### Delivered

- Added a read-only `/api/review` resource and contract wiring that only reads
  review rounds for the current active plan without changing review CLI
  contracts or write-side behavior.
- Replaced the `Review` placeholder with an overview-first round browser that
  lists rounds, summarizes round status, and centers reviewer task/result
  content instead of raw artifact management.
- Added conservative degraded-state handling for missing, incomplete, and
  malformed review artifacts, including semantic validation for required JSON
  fields and reviewer warnings when ledger/submission state disagrees.
- Kept manifest, ledger, aggregate, and submission payloads available as
  supporting evidence while making their summaries visible directly in the UI
  so damaged-artifact diagnosis does not require raw JSON spelunking.
- Added focused backend coverage, `/api/review` integration coverage, review
  smoke automation, main UI smoke updates, and manual headed Playwright
  screenshots for final aesthetic validation.

### Not Delivered

- Review history browsing across archived or non-active plans.
- Deep finding navigation into the future `Diff` and `Files` data surfaces.
- UI-triggered review actions or command handoff affordances.
- Rich side-by-side raw artifact diffing beyond the current supporting
  evidence tabs.

### Follow-Up Issues

- Issue [#103](https://github.com/catu-ai/easyharness/issues/103) tracks the
  deferred review-browser follow-ups from this slice: cross-plan history,
  deeper finding navigation, potential review actions, and richer artifact
  inspection.
- Issue [#91](https://github.com/catu-ai/easyharness/issues/91) remains the
  related dependency for wiring finding locations into future `Diff` / `Files`
  views.
