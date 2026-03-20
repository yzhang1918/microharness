---
name: harness-execute
description: Use when a tracked harness plan has been approved and the controller agent should drive implementation, review, CI, sync, publish handoff, closeout, and archive work until the archived candidate is genuinely ready to wait for merge approval. This is the main controller skill for day-to-day execution after approval.
---

# Harness Execute

## Purpose

Use this skill after plan approval to drive the repository from active work to
an archived, merge-ready candidate.

ALWAYS use `harness-execute` whenever `harness status` resolves a current
approved tracked plan and the lifecycle does not clearly call for a different
skill. That includes fresh sessions, resumed sessions after compaction, and
pick-up work where the safest next move is to follow the current plan instead
of improvising a new workflow.

The controller agent stays in `harness-execute` for the whole execution loop,
including review orchestration. Do not switch the controller into
`harness-reviewer`; that skill is only for spawned reviewer subagents assigned
to specific review slots.

Keep exactly one active review round at a time. The detailed review rules live
in [review-orchestration.md](references/review-orchestration.md).

For behavior-changing work, default to Red/Green/Refactor TDD. Only skip TDD
when it is genuinely impractical, and record the reason in the step's
`Execution Notes`.

## Start Here

1. Run `harness status`.
2. If `harness status` points to a current tracked plan that is already
   approved for execution, stay in `harness-execute` and open that plan from
   `plan_path`.
3. Identify the active or next plan step.
4. Use the status output to answer four questions:
   - which tracked plan is current
   - which lifecycle it is in
   - which step is active or next
   - whether local state already shows review, CI, or conflict work in flight
5. If `harness` is unavailable or resolves to the wrong binary, first follow
   the repository's documented setup path. If no setup path is documented, ask
   the human to install or expose the correct `harness` command.
6. Read only the references needed for the current part of the loop.

## Lifecycle Hints

- `awaiting_plan_approval`
  - wait for approval or update the plan if scope changed
- `executing`
  - continue the current plan step and use `step_state` as a local hint
- `blocked`
  - resolve the blocker or get human input
- `awaiting_merge_approval`
  - read `handoff_state` from `harness status`; finish publish or CI handoff
    first, and only wait for merge approval once the archived candidate is
    actually ready

## Reference Guide

- Read [step-inner-loop.md](references/step-inner-loop.md) when implementing or
  validating the current plan step.
- Read [review-orchestration.md](references/review-orchestration.md) whenever a
  review round is active or about to start.
- Read [publish-ci-sync.md](references/publish-ci-sync.md) when publish, CI, or
  remote-sync work becomes relevant.
- Read [closeout-and-archive.md](references/closeout-and-archive.md) before any
  archive attempt.

## Exit Criteria

Execute is done when:

- the plan is archived
- lifecycle is `awaiting_merge_approval`
- `harness status` no longer reports archived handoff follow-up such as
  `pending_publish`, `waiting_post_archive_ci`, or `followup_required`
- durable closeout summaries are written into the tracked plan

## Do Not

- Do not ask the human to micromanage routine execution once the plan is
  approved.
- Do not bypass lifecycle gates just because the next action feels obvious.
- Do not skip TDD for behavior changes without documenting why the usual
  Red/Green/Refactor loop was not practical.
- Do not rely on chat memory when `harness status`, the tracked plan, or local
  artifacts can tell you the truth more directly.
- Do not archive based on memory alone; use the current plan plus `.local`
  artifacts.
