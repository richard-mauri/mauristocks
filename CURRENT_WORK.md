# Current work — MauriStocks

## Issue #13 — watchlist heatmap (November 12, 2026)

- **Baseline:** GitHub main merge `ead1e7b2f7c4ff99b652aaf1c13ddf881fba5321` (PR #12 merged; previous #11 handoff documentation was stale).
- **Scoped branch and draft PR:** `feat/13-watchlist-heatmap`; https://github.com/richard-mauri/mauristocks/pull/14. Not merged or deployed.
- **Implemented:** Table/Heatmap UI switch for the selected watchlist, weighted binary treemap, regular-price versus previous-close color, native research links and titles, browser-local view preference, clearly marked unweighted securities.
- **Refresh:** Heatmap updates from the existing `refreshQuotes()` and in-memory `quotes`, with the same Never/5/30/60/300-second interval, manual refresh and hidden-tab pause. No second quote loop. Stable tile geometry for changing prices. Separate capitalization lookup only when switching/selecting symbols or opening the heatmap.
- **Capitalization:** Additive `GET /api/market-caps?symbols=...` route, server-side Finnhub Profile 2 using optional `FINNHUB_API_KEY`, validation of ticker/USD/positive finite capitalization, conversion from USD millions, sequential rate-spaced requests, 24-hour positive and 1-hour missing-value caches. Without a verified, configured API key the heatmap shows the unweighted list and does not fabricate data. The Finnhub free-plan quota/real-world profile coverage still require owner-side verification.
- **Source files changed:** `main.go`, `dashboard_javascript_syntax_test.go`, `README.md`, `ARCHITECTURE.md`, and this handoff. `DEVELOPMENT.md` and `AGENTS.md` remain unchanged.
- **Tests added:** HTTP 503 without capitalization credential, 405 unsupported method, and 400 invalid symbols. Existing embedded JavaScript parse test continues in test suite. No actual tests/build run via the connector; GitHub CI for latest PR head is not yet reported. Mac/Safari, live Yahoo quote/provider requests and Finnhub capitalization not verified.
- **Next:** Verify exact PR-head GitHub Actions, Go build/test and JS parse; verify real Finnhub key, missing ETF securities and quota behavior; inspect Safari desktop/mobile layout; confirm watchlist change, reload, timed and manual refresh, research links, table, account load, light/dark and hidden-tab behavior. Fix any findings in same branch and update PR. Never represent CI as Mac/Safari acceptance.
- **Owner boundary:** Draft only; no merge, tag, release, semantic version bump, service/tunnel restart, or deployment without explicit owner approval.

## Mac verification (only after owner permits service restart)

```sh
cd ~/projects/mauristocks
git status -sb
git fetch origin
git switch feat/13-watchlist-heatmap
git pull --ff-only origin feat/13-watchlist-heatmap
go test -count=1 ./...
go build ./...
bash ./devbuild.sh
./mauristocks -version
# Inspect the local stocksgo.sh launcher and get approval before running it.
```
