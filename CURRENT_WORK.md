# Current work — MauriStocks

## Research destination synchronization (November 12, 2026)

- **Merged baseline:** GitHub main contains PR #15, market Top 30 + heatmap. Work only in the original Mac directory `~/projects/mauristocks`; use `bash ./devbuild.sh` for native local builds.
- **Active branch:** `fix/heatmap-research-links`, draft PR pending. No merge/deployment/restart.
- **Issue:** Changing Symbol links from Google Finance to Stock Analysis, Corporate Website, or Investor Relations only called `renderQuotes()`; heatmap tiles reuse existing anchor elements and were not repainted until the next quote/metadata event. Thus tile clicks could open the previously selected provider.
- **Fix:** Shared dropdown change handler calls both `renderQuotes()` and `paintHeatmap()`, so existing anchor hrefs update immediately without recreating tile geometry, altering colors, fetching quotes, or changing selected account/watchlist.
- **Regression:** Embedded JS source contract checks both view updates on the research-site change event. Original shared palette remains unchanged; V8 JS parser passes. GitHub Go CI and owner Safari manual verification pending.
- **Owner boundary:** Keep draft until approved. No merge, version bump, build counter change, release, deployment, or MauriStocks/Cloudflare restart. For Safari validation on Mac, use original checkout and `bash ./devbuild.sh`, port 8083 only if authorized.
