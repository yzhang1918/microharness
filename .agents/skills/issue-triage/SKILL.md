---
name: issue-triage
description: Triage GitHub issues for the easyharness repository using the local backlog policy, concrete version milestones, and required rationale comments. Use when reviewing a new issue, backfilling labels on existing issues, revisiting deferred or needs-info backlog items, or deciding whether an issue should stay open, move into a version milestone, or close as not planned.
---

# Issue Triage

## Overview

Use this skill only for `easyharness` repository backlog work. Follow the
policy in [docs/issue-triage.md](../../../docs/issue-triage.md)
and leave a short rationale comment whenever you first triage an issue or
change its triage state later.

## Workflow

1. Read the issue body, relevant comments, and any linked plan or release
   context before choosing labels.
2. Apply or correct the default GitHub type label when the issue clearly fits
   `bug`, `enhancement`, `documentation`, or `question`.
3. Decide whether the issue belongs in one of the open-backlog states:
   - `state/accepted`
   - `state/needs-info`
   - `state/deferred`
4. If the issue truly belongs to a concrete release scope, use a version
   milestone such as `v0.2.2` instead of a `state/*` label.
5. Close the issue as not planned when that is the honest outcome instead of
   inventing another open-state label.
6. Leave a short rationale comment that records the judgment, the main reason,
   and what would cause a revisit when that matters.
7. When revisiting `state/deferred` or `state/needs-info`, read the earlier
   rationale comment first and update the state only when the earlier reason no
   longer holds.

## Decision Rules

- Use at most one `state/*` label on an open issue.
- Remove the `state/*` label when assigning a concrete version milestone.
- Treat any `state/*` label or version milestone as "triaged".
- Do not add `priority/*`, release-bucket labels, or project-board structure
  unless a human explicitly asks for a broader system.
- Do not leave reviewed negative outcomes open; close the issue instead when it
  is genuinely not planned.

## Rationale Comments

Use the short comment shapes in
[references/rationale-comments.md](./references/rationale-comments.md) when
helpful, but adapt them to the actual issue. Keep the comment short and
specific. Future backlog sweeps should be able to answer two questions from it:

- why was this state chosen then?
- what would make it reasonable to revisit now?

## Guardrails

- Keep this skill repo-local. Do not add `easyharness-managed` metadata or move
  it into `assets/bootstrap/` unless the repository explicitly decides to ship
  it later.
- Do not rely on labels alone when a short rationale comment would prevent
  future confusion.
- Do not assign a milestone unless the issue genuinely belongs in that version.
