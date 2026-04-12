---
template_version: 0.2.0
created_at: 2026-04-12T15:01:35+08:00
source_type: direct_request
source_refs: []
size: XXS
workflow_profile: lightweight
---

# Highlight Harness UI As The Human Steering Surface In README

## Goal

Update the root `README.md` so it clearly explains that `harness ui` is the
most important built-in surface for human steering. The current README mentions
the command in a feature list, but it does not yet frame the UI as the main
place where humans inspect workflow state and steer agent work.

## Scope

### In Scope

- Add or refine README language that introduces `harness ui` as the primary
  human steering interface.
- Keep the change limited to documentation and existing product behavior.

### Out of Scope

- Any CLI, UI, or workflow behavior changes.
- Broader README restructuring unrelated to the UI steering gap.

## Acceptance Criteria

- [ ] The README explicitly tells readers that `harness ui` is the main UI
      surface for human steering.
- [ ] The updated wording stays aligned with current product behavior and keeps
      the change lightweight and documentation-only.

## Deferred Items

- None.

## Work Breakdown

### Step 1: Update README to spotlight the UI steering surface

- Done: [x]

#### Objective

Add a small README update that makes the human-steering role of `harness ui`
easy to notice and understand.

#### Details

This qualifies for the lightweight path because it is a bounded `XXS`
documentation clarification in one file, with no contract, runtime, review, or
state-machine changes.

#### Expected Files

- `README.md`

#### Validation

- Re-read the updated README section for clarity and consistency with the
  current CLI surface.
- No automated tests are required because the change is documentation-only.

#### Execution Notes

- Added a Quickstart note that points humans to `harness ui` when they want the
  main built-in steering surface.
- Clarified that `harness ui` is the local read-only workbench for inspecting
  the current plan, workflow status, and execution summaries.
- Added `harness ui` to the `Workflow Surface` core-ideas list as the primary
  built-in human steering surface.

#### Review Notes

PENDING_STEP_REVIEW

## Validation Strategy

- Inspect the edited README to confirm the UI is described as the main human
  steering surface without overstating current capabilities.

## Risks

- Risk: README wording could imply UI capabilities that do not exist yet.
  - Mitigation: Keep the language anchored to steering, visibility, and
    existing workflow inspection rather than inventing new UI behavior.

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
