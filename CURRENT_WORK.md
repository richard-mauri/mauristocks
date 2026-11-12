# Current work — MauriStocks

## Combined UI persistence + weighted heatmap (November 12, 2026)

- **Merged baseline:** GitHub main includes merged PR #16.
- **Single active test branch:** `feat/adjustable-heatmap-weighting` and draft [PR #18](https://github.com/richard-mauri/mauristocks/pull/18) now include the previously separate PR #17 feature work. PR #17 is superseded; the owner should only check out/test the single combined branch.
- **Auto-persistence:** Remove manual Save Changes and Load from account controls. Named watchlists autosave after 600 ms of input inactivity, on field change, on refresh, and before changing watchlists/markets. Save As New remains. Default and dynamic lists stay fixed.
- **Account sync:** Homepage's existing authenticated, isolated per-user schema-1 record saves all changed personal settings, including Market, View, and Tile weighting. At sign-in, different account/browser settings require explicit confirmation before account restoration; account-overwrite consent and optimistic revision checks are preserved. Anonymous/localhost use remains browser-only. No Homepage server API change.
- **Heatmap:** Tile-weight exponent 0.50–1.00, step 0.05, default 0.65. Preserve actual cap values and tile colors. Non-destructive tile reconciliation avoids the reported flash on refresh. Browser preference persists locally and, when signed in, in account settings.
- **Testing:** Combined Go regression/source-contract tests and embedded JS V8 syntax; exact head Go CI and Mac/Safari test pending.
- **Mac workflow:** Original checkout `~/projects/mauristocks` only; use `git switch feat/adjustable-heatmap-weighting`, `git pull --ff-only origin feat/adjustable-heatmap-weighting`, `bash ./devbuild.sh`, then `./mauristocks -host 127.0.0.1 -port 8083`. No additional worktree. Do **not** restart `stocksgo.sh`/Cloudflare without owner approval. No merge, release, semantic-version bump or deployment without explicit approval.
