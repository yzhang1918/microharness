---
template_version: 0.2.0
created_at: "2026-04-11T21:38:59+08:00"
source_type: issue
source_refs:
    - https://github.com/catu-ai/easyharness/issues/36
size: S
---

# Close Issue 36 With Focused Parsing Coverage

## Goal

Close `#36` by adding focused fuzz or property-style coverage for the
highest-value parsing-heavy harness paths without expanding into deterministic
resilience work, repo-level lifecycle E2E, or broader concurrency hardening.

This slice should make a clear repository-level judgment about which candidate
paths deserve fuzz/property investment now. The expected closeout is: plan
markdown parsing and command-input schema decoding gain meaningful new
coverage, while review artifact readers and historical evidence-record readers
are explicitly judged out of scope for this issue because their current
deterministic coverage is already stronger than the remaining risk reduction
available from a first fuzzing pass.

## Scope

### In Scope

- Add package-level Go fuzz tests for the plan markdown parsing surface in
  `internal/plan`, centered on `LintFile`, `LoadFile`, and the shared parsing
  helpers they exercise.
- Add property-style or seed-based invariants in `internal/plan` that check
  stable relationships between linting and document loading on canonical plan
  inputs.
- Add focused fuzz or property-style coverage for `internal/inputschema`
  normalization logic, especially JSON-pointer rendering, quoted-property
  extraction, and parent-issue pruning.
- Keep evidence command coverage aligned with the schema layer where that helps
  prove schema-derived error paths still surface correctly through a real
  command entrypoint.
- Leave an execution trail strong enough that archive or issue closeout can say
  `#36` was intentionally closed after evaluating plan lint, review artifacts,
  and evidence payload decoding rather than only touching one of them.

### Out of Scope

- `tests/resilience/`, deterministic failure-path coverage, or any work owned
  by `#37`.
- Repo-level lifecycle E2E coverage, fixture expansion, or changes under
  `tests/support/`.
- Broader concurrency or lock-behavior coverage owned by `#56`.
- Deep fuzzing of `internal/reviewui` artifact recovery or `internal/evidence`
  historical record loading beyond what is needed to justify why those readers
  are not the primary targets for closing `#36`.
- Schema redesigns, command-shape changes, or new issue/follow-up creation.

## Acceptance Criteria

- [ ] `internal/plan` has new package-level fuzz or property-style coverage for
      parsing-heavy plan inputs, and the targeted surfaces do not panic when
      fed arbitrary data plus seeded canonical plan examples.
- [ ] Canonical valid-plan seeds assert at least one stable plan invariant such
      as: lint success and document loading stay aligned, current-step
      detection remains deterministic, or archive-readiness helpers stay
      coherent after successful parsing.
- [ ] `internal/inputschema` has new fuzz or property-style coverage for path
      rendering and validation-error normalization, including nested-array
      paths, quoted-property extraction, and parent-issue pruning behavior.
- [ ] Any touched `internal/evidence` regression coverage stays narrowly tied
      to schema-decoding behavior and does not expand into resilience-style
      malformed-artifact recovery.
- [ ] The slice documents, through plan execution notes and closeout, that
      review artifact readers and historical evidence-record readers were
      evaluated but were not required additions for closing `#36`.

## Deferred Items

- None.

## Work Breakdown

### Step 1: Add focused fuzz coverage for plan markdown parsing

- Done: [ ]

#### Objective

Introduce high-signal fuzz or property-style tests in `internal/plan` that
exercise the mixed YAML and Markdown parsing surface without widening into
repo-level fixtures or resilience infrastructure.

#### Details

Target the shared parser surface behind `LintFile` and `LoadFile`, not a new
parallel helper API. Seed the fuzz corpus with valid rendered plans plus a
small number of invalid structured variants so the engine learns both success
and failure shapes. Prefer invariants that stay useful under future template
evolution, such as no panic, stable error/result shape expectations for seeded
examples, and coherence between successful lint and successful document load
for canonical valid plans.

#### Expected Files

- `internal/plan/lint_test.go`
- `internal/plan/document_test.go`
- `internal/plan/*_test.go`

#### Validation

- `go test ./internal/plan`
- `go test -fuzz=Fuzz -run=^$ ./internal/plan` for a bounded fuzz pass

#### Execution Notes

Added `internal/plan/fuzz_test.go` with canonical seed properties that keep
`LintFile` and `LoadFile` aligned across active, archived, and archived
lightweight plans; added a tracked-plan corpus check that every repository plan
that lints cleanly also loads cleanly; and added a bounded file-based fuzz
target asserting `LintFile` success implies `LoadFile` success while document
helper methods stay panic-free on arbitrary inputs. Validated with
`go test ./internal/plan` and
`go test -run=^$ -fuzz=FuzzLintFileAndLoadFileAgreement -fuzztime=2s ./internal/plan`.

#### Review Notes

PENDING_STEP_REVIEW

### Step 2: Cover schema-driven input decoding and keep issue closeout bounded

- Done: [ ]

#### Objective

Add focused fuzz or property-style coverage for command-input schema
normalization in `internal/inputschema`, then keep any supporting evidence
tests tightly limited to proving those normalized errors still reach a real
command surface.

#### Details

Concentrate on the path-shaping logic that easyharness owns locally:
`renderInstanceLocation`, `propertiesFromValidationMessage`,
`renderIssueDetails`, and `pruneParentIssues`. Use schema-backed seeds so the
tests stay grounded in real command inputs instead of arbitrary generated
structures. If `internal/evidence/service_test.go` needs small updates, keep
them at the command-boundary level and do not broaden into malformed historical
record loading or status-side conservative behavior. During execution, record
the explicit judgment that `internal/reviewui` already has strong deterministic
coverage for malformed and partial artifacts, so it is not the first fuzzing
target needed to close `#36`.

#### Expected Files

- `internal/inputschema/validator_test.go`
- `internal/inputschema/*_test.go`
- `internal/evidence/service_test.go`

#### Validation

- `go test ./internal/inputschema ./internal/evidence`
- `go test -fuzz=Fuzz -run=^$ ./internal/inputschema` for a bounded fuzz pass

#### Execution Notes

PENDING_STEP_EXECUTION

#### Review Notes

PENDING_STEP_REVIEW

## Validation Strategy

- Use package-level `go test` runs for the touched units instead of repo-level
  E2E suites.
- Run bounded Go fuzz passes only in `internal/plan` and `internal/inputschema`
  so the work stays isolated from `tests/resilience/`, `tests/support/`, and
  other worktrees handling `#37` plus `#56`.
- Treat closeout as incomplete unless the final validation record can explain
  why review artifact readers and historical evidence readers were evaluated
  but not made primary fuzz targets for this issue.

## Risks

- Risk: Fuzz targets around file-based plan parsing can become flaky or too
  coupled to temporary filesystem setup.
  - Mitigation: Seed from deterministic tempdir fixtures and keep invariants
    focused on no-panic and stable parser relationships rather than brittle
    exact-error text for random inputs.
- Risk: The slice could drift into resilience or malformed-artifact hardening
  already separated into `#37`.
  - Mitigation: Keep all new work inside package tests for `internal/plan`,
    `internal/inputschema`, and narrowly scoped evidence regressions.
- Risk: Closing `#36` could look premature if the plan does not explicitly
  justify why review artifact readers were not fuzzed.
  - Mitigation: Make that judgment explicit in execution and archive summaries
    and only archive once the added coverage plus rationale would let a cold
    reviewer understand the decision.

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
