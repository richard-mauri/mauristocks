# MauriStocks — current work

**As of 2026-10-26.** Authoritative accepted baseline: GitHub `main`, before the present UI/docs PR. Active development branch: `ui/content-fit-columns-and-docs` ([draft PR #2](https://github.com/richard-mauri/mauristocks/pull/2)). No merge, tag, release or deployment authorized.

## Goal and decisions

Remove painful per-header drag/keyboard resizing. Visible table columns should fit their content with modest padding; longer company names can truncate at 340 px, and range cells retain a usable graphical minimum. Keep Columns show/hide, browser-persisted hidden column selections and the nonhideable Symbol column. Discard legacy saved widths without deleting user hidden selections; retain session-based extended-price suppression. Align GitHub Actions triggers/Go steps with ScraperGo and separate root documents by responsibility.

## Acceptance / open work

GitHub Actions CI needs verification at the newest PR head after edits. Owner-side Mac/Safari verification is pending. No evidence of local installation, browser acceptance or release. Cross-check dark/light styling, sort and regular/extended sessions. The embedded page remains monolithic; structural Go-package refactoring is out of scope. Update this snapshot when PR state, evidence or next actions change.