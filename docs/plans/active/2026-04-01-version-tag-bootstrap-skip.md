---
template_version: 0.2.0
created_at: 2026-04-01T20:44:49+08:00
source_type: direct_request
source_refs: []
workflow_profile: lightweight
---

# Cut Release VERSION to 0.1.0-alpha.6

## Goal

Ship the next alpha release by advancing the root `VERSION` file from
`0.1.0-alpha.5` to `0.1.0-alpha.6` in a dedicated release PR. This intentionally
avoids changing the bootstrap workflow logic; instead, it uses the newly
VERSION-driven release path the way maintainers expect to use it going
forward.

The outcome should be a merge-ready release PR that bumps the version, updates
any release-facing docs that still mention the old example version, and leaves
the repository ready for automation to create the new `v0.1.0-alpha.6` tag and
dispatch the existing `Release` workflow after merge.

## Scope

### In Scope

- Update the root `VERSION` file to `0.1.0-alpha.6`.
- Refresh maintainer-facing release docs where the example next version should
  now point to `0.1.0-alpha.6`.
- Run the relevant local validation for the release bump PR.
- Carry the change through review, archive, publish, and wait-for-merge
  readiness.

### Out of Scope

- Fixing the one-time bootstrap workflow behavior for already-published tags.
- Changing release cadence, stable-version policy, or workflow semantics.
- Any non-release product or harness behavior changes.

## Acceptance Criteria

- [ ] The root `VERSION` file is updated to `0.1.0-alpha.6`.
- [ ] Maintainer docs accurately describe `0.1.0-alpha.6` as the next release
      example where the old example would now be stale or confusing.
- [ ] Relevant local validation passes for the release bump candidate.
- [ ] The candidate reaches archived, published, merge-ready state for a
      dedicated alpha.6 release PR.

## Deferred Items

- Consider a separate follow-up only if the one-time bootstrap workflow failure
  needs to be cleaned up later for repository hygiene.

## Work Breakdown

### Step 1: Prepare and publish the alpha.6 release bump

- Done: [ ]

#### Objective

Update `VERSION`, refresh any adjacent release guidance, validate the release
bump, and carry the dedicated alpha.6 candidate through the lightweight
publish path.

#### Details

Keep the diff intentionally narrow and release-focused. The release PR should
remain separate from feature work and should not change release automation
logic. This plan uses `workflow_profile: lightweight` as a user-approved
exception to the repository's default lightweight eligibility guidance because
the actual code change is intentionally minimal: a one-line release version
bump plus matching release-doc examples.

#### Expected Files

- `VERSION`
- `docs/releasing.md`
- `README.md`
- `docs/plans/active/2026-04-01-version-tag-bootstrap-skip.md`

#### Validation

- `VERSION` resolves to `v0.1.0-alpha.6`.
- Docs remain internally consistent about the release flow and examples.
- Relevant local tests or release checks pass.
- Review and archive artifacts accurately reflect that this is a dedicated
  alpha.6 release bump candidate.

#### Execution Notes

Updated `VERSION` from `0.1.0-alpha.5` to `0.1.0-alpha.6` and refreshed the
README release-process example to match. `docs/releasing.md` already used the
alpha.6 examples, so no change was needed there. Validated the bump with
`scripts/read-release-version --tag` and `go test ./tests/smoke -run
ReleaseVersion -count=1`.

#### Review Notes

PENDING_STEP_REVIEW

## Validation Strategy

- Run the narrowest relevant local validation for the release bump candidate.
- Re-read the docs and version file together to confirm the examples and tag
  mapping remain correct.

## Risks

- Risk: The version bump could still leave stale release examples or summaries
  that mention alpha.5.
  - Mitigation: Update the few maintainer-facing docs that are part of the
    release path and verify the rendered version examples match `0.1.0-alpha.6`.

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
