# Rationale Comments

Keep triage comments short. They should record the judgment, the main reason,
and the revisit trigger when that helps future backlog sweeps.

Examples:

```text
Triaged as state/accepted. This looks like real follow-up work, but I do not
yet want to commit it to a specific release scope.
```

```text
Triaged as state/needs-info. I do not yet have enough evidence about the
problem shape or the expected UX. Revisit once there is a concrete workflow
example or tighter acceptance target.
```

```text
Triaged as state/deferred. The issue still looks worth keeping, but it is not
the current window and I do not want to lock in the wrong shape yet. Revisit
after the adjacent workflow surface settles.
```

```text
Triaged into milestone v0.2.2. This belongs in the next patch release because
it tightens an already-shipped maintainer workflow without widening scope into
broader policy work.
```
