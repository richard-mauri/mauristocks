# Current work — MauriStocks

## Automatic browser persistence (November 12, 2026)

- **Baseline:** GitHub main includes merged PR #16, research links synchronize with heatmap. Owner Mac checkout `~/projects/mauristocks` uses `bash ./devbuild.sh` for local build; no extra worktree.
- **Feature branch:** `feat/automatic-watchlist-persistence`, draft PR pending, not merged or deployed.
- **User intent:** Remove Save Changes and Load from account UI buttons. Anonymous users' personal watchlist edits and preferences should save automatically in browser localStorage; signed-in users keep account syncing without risky silent overwrite.
- **Implementation:** Named watchlist symbol edits autosave locally after 600 ms of inactivity, on field change, on Refresh, or before changing list/market. Save As New remains for the initial creation of a named list; Default and Most Active are fixed templates. Remove both manual controls. Existing preference autosaves remain. Authenticated account reconciliation reads the saved record once per account identity and asks before restoring divergent settings. Existing account overwrite approval and revision conflict checks remain. No change to quote APIs, CSS palette, market source, or login token handling.
- **Testing:** Go regression source-contract checks and embedded JavaScript syntax. GitHub Actions CI exact PR head and Mac/Safari tests pending.
- **Owner boundary:** Do not merge, release, deploy, restart MauriStocks or the Cloudflare tunnel, or bump semantic version without explicit approval. On the Mac, use original directory, `bash ./devbuild.sh`, and test on separate port if appropriate.
