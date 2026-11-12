# Current work — MauriStocks

## Issue #19 — Homepage suite activity reporting (November 12, 2026)

- **Verified merged baseline:** GitHub `main` is at PR #18's merged heatmap/autosave commit `b92dc5ea6fc4821c7ca6a3896373794a2758f16e`. GitHub comparison `main...feat/19-suite-activity` shows **8 commits ahead and 0 behind**, with merge base exactly that merged PR #18 commit. Activity PR #20 **already contains** the adjustable heatmap weighting, autosave, per-user settings sync, and flicker fixes. No code merge/rebase or separate reconciliation is required. Activity work is on `feat/19-suite-activity`, draft [PR #20](https://github.com/richard-mauri/mauristocks/pull/20), Issue #19.
- **Scope:** optional per-tab anonymous/unverified visibility and interaction pings, rolling 2/5/15/60-minute observed presence, and bounded in-flight API request accounting for quotes, market caps/top, Most Active and company lookup. One server-only aggregate report per 60 seconds to Homepage collector with `stocks` identity, keyed by `HOMEPAGE_ACTIVITY_KEY_STOCKS` and URL `STOCKS_ACTIVITY_URL`. Unconfigured reporter has no page script or outbound reports.
- **Security:** the browser receives no collector key; telemetry sends no symbol, account, portfolio, price, tab ID or session secret to Homepage. Signed-in tabs are not independently verified for telemetry and therefore count as anonymous/unverified. Reporter failures do not affect quote or watchlist behavior. Independent tunnels and per-user settings are unchanged.
- **Validation:** Go tests cover protocol, presence windows, same-origin rejection, in-flight work and disabled config. GitHub Go CI #127 passed at code head `560156c25eacc8155cf4cd60364c9427ced3acc5`; this documentation commit requires its own latest-head CI check. Mac/Safari/real Homepage collector integration **not run**. No deployment, restart, version bump or merge.
- **Next:** check documentation-head CI; run isolated Mac test on port 8086 with private matching service key and existing test Homepage on 8788, verifying MauriStocks activity and the already-included merged heatmap features in Safari. Owner must authorize any production installation/restart.

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

The launcher restarts the installed service and tunnel. Do not use it for isolated smoke testing or until owner authorization. The owner's checkout was confirmed clean on `main`. The activity branch already includes merged PR #18; use an isolated worktree for smoke testing or intentionally switch this clean checkout after confirming no new changes.
