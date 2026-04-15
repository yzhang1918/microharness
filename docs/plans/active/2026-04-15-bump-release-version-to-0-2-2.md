---
template_version: 0.2.0
created_at: "2026-04-15T09:31:33+08:00"
approved_at: "2026-04-15T09:32:38+08:00"
source_type: direct_request
source_refs: []
size: XXS
workflow_profile: lightweight
---

# Bump Release Version To 0.2.2

<!-- If this plan uses supplements/<plan-stem>/, keep the markdown concise,
absorb any repository-facing normative content into formal tracked locations
before archive, and record archive-time supplement absorption in Archive
Summary or Outcome Summary. Lightweight plans should normally avoid
supplements. -->

## Goal

Prepare the dedicated release PR input for `v0.2.2` by updating the root
`VERSION` file from `0.2.1` to `0.2.2`.

This plan is intentionally narrow. The motivating issue is that the current
public Homebrew release still lags behind the repository's newer release
contract, but this slice only bumps the release seed version. It does not fix
the dev-only wrapper or change release automation behavior.

## Scope

### In Scope

- Update the root `VERSION` file to `0.2.2`.
- Keep the execution slice limited to the version bump plus required plan
  lifecycle bookkeeping.
- Record that the human explicitly approved `workflow_profile: lightweight`
  for this `XXS` release-preparation slice.

### Out of Scope

- Any wrapper or CLI behavior changes.
- Any release workflow, Homebrew tap, or GitHub Actions changes.
- Any tag creation, release publication, or manual Homebrew operations.
- Any README, spec, or release-note edits beyond what the future release PR may
  choose to add separately.

## Acceptance Criteria

- [x] `VERSION` contains `0.2.2`.
- [x] Execution changes stay limited to `VERSION` and plan lifecycle updates.
- [x] The resulting branch is suitable to use as the dedicated release PR input
      that later merge automation can turn into `v0.2.2`.
- [x] No wrapper fix or release-pipeline changes are included in this slice.

## Deferred Items

- None.

## Work Breakdown

### Step 1: Bump the tracked release seed to 0.2.2

- Done: [x]

#### Objective

Change the root `VERSION` file from `0.2.1` to `0.2.2` so the next dedicated
release PR advertises the intended stable patch release.

#### Details

The human explicitly approved `workflow_profile: lightweight` for this `XXS`
slice even though release-adjacent work is usually kept on the standard path.
That exception is acceptable here because execution is confined to one tracked
version file, does not modify runtime behavior, and leaves the existing
tag/release/Homebrew automation untouched. The wrapper mismatch discussed in
discovery is motivation only and remains out of scope for this plan.

#### Expected Files

- `VERSION`

#### Validation

- Confirm `VERSION` reads `0.2.2`.
- Run `scripts/read-release-version --tag` and confirm it reports `v0.2.2`.
- Inspect the diff to confirm execution stayed within the planned narrow scope.
- No automated tests are required because this slice does not change runtime
  or workflow logic.

#### Execution Notes

Updated `VERSION` from `0.2.1` to `0.2.2` and confirmed the release helper now
reports `v0.2.2` through `scripts/read-release-version --tag`. Diff inspection
confirmed the execution slice stayed within the planned narrow scope: the
release seed plus tracked plan lifecycle notes only.

When execution finishes, leave the required lightweight breadcrumb in the PR
body or another approved review surface before treating the candidate as ready
to wait for merge approval.

#### Review Notes

NO_STEP_REVIEW_NEEDED: this slice is a single tracked version-field bump with
no code, contract, or automation logic changes.

## Validation Strategy

- Use direct file inspection plus `scripts/read-release-version --tag`.
- Review the final diff to ensure no accidental scope expansion occurred.

## Risks

- Risk: A broader release-readiness concern could get conflated with this tiny
  version bump.
  - Mitigation: Keep the plan explicit that this slice only updates `VERSION`
    and leaves all release mechanics and wrapper behavior unchanged.
- Risk: Merging the release PR before maintainers are ready would publish
  `v0.2.2` too early through the existing automation.
  - Mitigation: Human merge approval still gates release timing after this
    lightweight slice is prepared.

## Validation Summary

PENDING_UNTIL_ARCHIVE

## Review Summary

PENDING_UNTIL_ARCHIVE

## Archive Summary

PENDING_UNTIL_ARCHIVE

## Outcome Summary

### Delivered

PENDING_UNTIL_ARCHIVE

### Not Delivered

PENDING_UNTIL_ARCHIVE

### Follow-Up Issues

NONE
