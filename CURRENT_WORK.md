# Current work — MauriStocks

## Issue #19 — Homepage suite activity reporting (November 12, 2026)

- **Merged baseline:** GitHub `main` includes prior merged PR #16; all activity work is on `feat/19-suite-activity`, draft [PR #20](https://github.com/richard-mauri/mauristocks/pull/20), related to [Issue #19](https://github.com/richard-mauri/mauristocks/issues/19). This is independent of owner-active heatmap/persistence draft PR #18 on `feat/adjustable-heatmap-weighting`; its source is not modified or merged by activity work.
- **Scope:** optional per-tab anonymous/unverified visibility and interaction pings, rolling 2/5/15/60-minute observed presence, and bounded in-flight API request accounting for quotes, market caps/top, Most Active and company lookup. One server-only aggregate report per 60 seconds to Homepage collector with `stocks` identity, keyed by `HOMEPAGE_ACTIVITY_KEY_STOCKS` and URL `STOCKS_ACTIVITY_URL`. Unconfigured reporter has no page script or outbound reports.
- **Security:** the browser receives no collector key; telemetry sends no symbol, account, portfolio, price, tab ID or session secret to Homepage. Signed-in tabs are not independently verified for telemetry and therefore count as anonymous/unverified. Reporter failures do not affect quote or watchlist behavior. Independent tunnels and per-user settings are unchanged.
- **Validation:** Go tests cover protocol, presence windows, same-origin rejection, in-flight work and disabled config. Exact-head GitHub CI is pending. Mac/Safari/real Homepage collector integration **not run**. No deployment, restart, version bump or merge.
- **Next:** check exact-head CI; run detached isolated Mac build on port 8086 with private matching service key and the existing test Homepage on 8788; verify the MauriStocks card in Safari. Preserve the separate heatmap PR and seek explicit owner approval before any production promotion.

## Owner Mac acceptance (requires explicit permission before production restart)

```sh
cd ~/projects/mauristocks
git status -sb
git fetch origin
git checkout feat/19-suite-activity
git pull --ff-only origin feat/19-suite-activity
bash ./devbuild.sh
./mauristocks -version
./stocksgo.sh
```

The launcher restarts the installed service and tunnel. Do not use it for isolated smoke testing or until owner authorization. Use a detached worktree to avoid disturbing the owner's active PR #18 checkout.
