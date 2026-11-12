# Current work — MauriStocks

## Adjustable heatmap tile weighting (November 12, 2026)

- **Authoritative baseline:** GitHub `main` includes merged PR #16. PR #17 (`feat/automatic-watchlist-persistence`) is an independent draft and is **not** included in this branch. Owner’s original checkout is `~/projects/mauristocks`; build via `bash ./devbuild.sh`.
- **Active branch:** `feat/adjustable-heatmap-weighting`, draft PR pending, not merged or deployed.
- **Scope:** Heatmap Tile weighting slider from 0.50 to 1.00, step 0.05, default 0.65. It changes binary treemap area to `market_cap ** exponent` while retaining company market-cap ranking, actual USD capitalization tooltips, quote prices and change colors. Label accurately distinguishes adjusted market caps from exact areas at 1.00. Persist independently in browser localStorage `mauristocks_heat_weight_v1`, including unsigned sessions.
- **Implementation:** Embedded HTML/CSS/JS in `main.go`; new slider above heatmap, preserved existing semantic palette, exponent in `layoutKey` triggers synchronous relayout without network requests. Browser preference kept separate from Homepage account schema.
- **Owner Safari report:** Feature branch flashed at each refresh while merged main did not. Fixed full-container `ha.replaceChildren();tiles.clear()` relayout; `paintHeatmap()` now reuses tile anchors for unchanged symbols and updates positions in place, so colors/links/focus survive refreshes or slider changes. Added Go regression check rejecting destructive panel clearing. Safari re-test pending.
- **Validation:** Embedded JavaScript syntax and Go source-contract regression; latest PR head CI and owner Mac/Safari checks pending.
- **Boundaries:** Keep separate from PR #17. No merge, version bump, release, deployment, service or Cloudflare tunnel restart without owner approval. Test in original Mac checkout using `bash ./devbuild.sh`, isolated port 8083, then return to main as needed.
