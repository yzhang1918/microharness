# Interactive Headed Playwright Validation

Date: 2026-04-14

## Purpose

Provide a repo-visible record for the headed Playwright browser pass used to
confirm the shared workbench shell stays visually stable while navigating among
compatible pages.

## Flow

1. Opened `harness ui` in a headed Playwright browser on `/timeline`.
2. Cleared `localStorage`.
3. Dragged the `Timeline width` separator to widen the explorer.
4. Navigated through the real rail links in this order:
   `Timeline -> Review -> Status -> Plan`.
5. Recorded the rendered workbench grid columns and the shared persisted width
   after each page switch.

## Observed Result

The headed Playwright run reported the following values:

```json
{
  "routeWidths": {
    "Timeline": "360px 8px 784px",
    "Review": "360px 8px 784px",
    "Status": "360px 8px 784px",
    "Plan": "360px 8px 784px"
  },
  "stored": "360",
  "finalUrl": "http://127.0.0.1:4310/plan#overview"
}
```

## Conclusion

The explorer/inspector split stayed at the same `360px` width across all four
shell-compatible pages during a headed browser session, and the browser stored
that width once under `harness-ui:workbench-explorer-width`.
