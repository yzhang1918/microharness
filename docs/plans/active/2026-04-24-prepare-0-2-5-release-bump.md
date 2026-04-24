---
template_version: 0.2.0
created_at: "2026-04-24T23:15:29+08:00"
approved_at: "2026-04-24T23:17:24+08:00"
source_type: direct_request
source_refs: []
size: XXS
workflow_profile: lightweight
---

# Prepare the 0.2.5 release bump

<!-- If this plan uses supplements/<plan-stem>/, keep the markdown concise,
absorb any repository-facing normative content into formal tracked locations
before archive, and record archive-time supplement absorption in Archive
Summary or Outcome Summary. Lightweight plans should normally avoid
supplements. -->

## Goal

Prepare the dedicated release change for `easyharness` version `0.2.5` by
updating the repository's single tracked release source of truth from `0.2.4`
to `0.2.5`.

This is intentionally a lightweight release-bump plan: the change is XXS, is
confined to the release version file, does not alter release automation or
release-safety behavior, and should leave the existing VERSION-driven release
workflow ready to publish `v0.2.5` after the release PR merges to `main`.

## Scope

### In Scope

- Bump the root `VERSION` file from `0.2.4` to `0.2.5`.
- Run focused validation proving the version helper resolves the target release
  tag as `v0.2.5`.
- Leave a concise lightweight workflow breadcrumb for the release PR or archive
  summary so reviewers can see that the candidate is only a dedicated version
  bump.

### Out of Scope

- Changing release workflow design, tag automation, GitHub Release publishing,
  or Homebrew tap behavior.
- Manually pushing a git tag or bypassing the documented release PR and merge
  path.
- Editing release notes, changelog content, archived plans, or unrelated docs.
- Adding compatibility, migration, or fallback behavior.

## Acceptance Criteria

- [x] `VERSION` contains exactly `0.2.5`.
- [x] `scripts/read-release-version --tag` returns exactly `v0.2.5`.
- [x] The final diff is confined to the lightweight plan/package lifecycle and
      the dedicated `VERSION` bump, with no release automation changes.
- [x] The candidate is ready for the normal release PR path, where merge to
      `main` creates the `v0.2.5` tag and dispatches the release workflow.

## Deferred Items

- Any changelog, announcement, or release-note packaging for `0.2.5`.
- Any changes to release automation or Homebrew publishing policy.

## Work Breakdown

### Step 1: Bump and validate the release version

- Done: [x]

#### Objective

Update `VERSION` to `0.2.5` and prove the existing release helper resolves the
matching `v0.2.5` tag.

#### Details

This qualifies for the lightweight path because it is a human-requested XXS
release bump with a narrow single-file product change. It does not change
runtime behavior, schema meaning, harness workflow semantics, release safety
logic, or publishing automation. The release itself remains governed by the
documented PR merge path in `docs/releasing.md`.

#### Expected Files

- `VERSION`
- `docs/plans/active/2026-04-24-prepare-0-2-5-release-bump.md`

#### Validation

- Run `scripts/read-release-version --tag` and confirm it prints `v0.2.5`.
- Review `git diff --stat` and the focused diff to confirm no unrelated
  changes entered the release candidate.

#### Execution Notes

Updated the root `VERSION` file from `0.2.4` to `0.2.5`. Focused validation
passed with `scripts/read-release-version --tag`, which returned `v0.2.5`.
The focused product diff contains only the one-line `VERSION` bump; the rest of
the candidate is the harness plan lifecycle for this lightweight release bump.
TDD was not applicable because this step is a release source-of-truth update,
not a behavior change.

#### Review Notes

NO_STEP_REVIEW_NEEDED: this lightweight XXS step changed only the release
version source of truth and was validated with the existing release helper.

## Validation Strategy

- Use the existing release helper as the focused validation for the VERSION to
  tag mapping.
- Cold-read the final diff before archive/publish handoff to confirm the change
  remains a dedicated lightweight release bump.

## Risks

- Risk: A release bump could accidentally pull in unrelated local edits and
  make the release PR harder to review.
  - Mitigation: inspect the final diff and keep the candidate confined to the
    active plan lifecycle plus the `VERSION` update.
- Risk: `VERSION` could be bumped without proving the automated tag name that
  will be created after merge.
  - Mitigation: run `scripts/read-release-version --tag` and record the exact
    `v0.2.5` output before archive.

## Validation Summary

- `scripts/read-release-version --tag` -> `v0.2.5`
- Focused diff review confirmed the product change is confined to `VERSION`
  (`0.2.4` -> `0.2.5`), with no release automation edits.

## Review Summary

No separate step review was needed for this XXS lightweight release bump. The
controller confirmed the diff and helper output before archive.

## Archive Summary

PENDING_UNTIL_ARCHIVE

## Outcome Summary

### Delivered

- Root `VERSION` now targets `0.2.5`.
- Existing release helper resolves the matching release tag as `v0.2.5`.
- Candidate remains on the normal release PR path; merge to `main` should let
  the existing VERSION-driven automation create the tag and dispatch release
  publishing.

### Not Delivered

NONE

### Follow-Up Issues

NONE
