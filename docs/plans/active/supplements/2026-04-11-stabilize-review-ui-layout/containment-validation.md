# Review Explorer Containment Validation

Current validated head: `bca2720`
Original containment fix commit: `ef1b89a`

This note records the direct 220px-width validation used to close the final
review explorer containment repair for
`2026-04-11-stabilize-review-ui-layout`, plus the revision-2 revalidation after
the clean `origin/main` sync merge.

## Browser Evidence

- Explorer row screenshot:
  `docs/plans/active/supplements/2026-04-11-stabilize-review-ui-layout/review-explorer-row-220.png`
- Header metadata screenshot:
  `docs/plans/active/supplements/2026-04-11-stabilize-review-ui-layout/review-header-meta.png`
- Revision-2 review smoke overview:
  `docs/plans/active/supplements/2026-04-11-stabilize-review-ui-layout/review-smoke-rev2-initial.png`
- Revision-2 review smoke active-row capture:
  `docs/plans/active/supplements/2026-04-11-stabilize-review-ui-layout/review-smoke-rev2-row-active.png`

The 220px containment screenshots were captured from the rebuilt embedded
`harness ui` binary with `harness-ui:explorer-width:review=220` in browser
local storage. Revision 2 then reran `scripts/ui-playwright-review-smoke`
successfully on merged head `bca2720` and copied two representative smoke
screenshots into this tracked supplement. The sync merge from `920deea` to
`bca2720` did not touch `web/src` or `internal/ui`, so the direct 220px
containment captures remain the current UI-specific evidence for this slice.

## DOM Measurements

The direct browser probe against the selected review row at the 220px explorer
width reported:

- button: `clientWidth=219`, `scrollWidth=219`
- main column: `clientWidth=182`, `scrollWidth=182`
- title row: `clientWidth=182`, `scrollWidth=182`
- subtitle row: `clientWidth=182`, `scrollWidth=182`
- metadata cell: `clientWidth=151`, `scrollWidth=151`
- title text: `clientWidth=182`, `scrollWidth=241`

Interpretation:

- the row container and subtitle layout are fully contained at the minimum
  explorer width because their `clientWidth` and `scrollWidth` values match
- the long title still overflows its own text box, but that overflow is now
  handled by ellipsis inside the title cell rather than by pushing the row
  wider than the explorer item

## Supporting Checks

- `pnpm --dir web check`
- `pnpm --dir web build`
- `scripts/ui-playwright-review-smoke`
- `git diff --check`
