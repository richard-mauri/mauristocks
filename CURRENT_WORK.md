# Current work — MauriStocks

## Issue #22 — verified Homepage guest/session and activity

- **Merged baseline:** GitHub `main` contains the heatmap/autosave and optional suite activity work (PRs #18 and #20). Source branch `feat/22-session-aware-activity` starts from `main`; draft PR pending. Issue: https://github.com/richard-mauri/mauristocks/issues/22.
- **Scope:** Reuse existing Homepage child bearer; validate it server-side against `/api/session/profile` before attributing browser presence. Deduplicate multiple tabs by Homepage's origin-scoped pseudonymous parent session alias; return signed-in sessions, unique users and bounded user summaries to the existing private collector. Revoked bearers return 401, temporarily unavailable Homepage returns 503, and anonymous pings remain distinct. No bearer or alias is included in transmitted telemetry.
- **Browser UX:** When optional activity monitoring is enabled, show verified named or one-hour guest identity and a non-dismissible sign-in dialog for invalid/local sessions with a deliberate Go to Homepage action. Guest settings remain browser-local. Named-user settings retain existing Homepage storage and conflict protection. Homepage remains sole sign-in authority. There is no automatic redirect on expiry.
- **Preserved:** quotes, table/heatmap, market/index calculation, watchlist autosave, browser-only persistence, environment secrets, independent tunnel and installed service. No semantic version bump.
- **Validation to finish:** Go compiler unavailable in agent's temporary environment; run GitHub Go CI at exact draft PR head. JavaScript syntax and Git diff whitespace checks passed before push. Owner Mac/Safari sign-in, guest, expiry/revocation and activity acceptance remain unverified. No merge, release, tag, restart or tunnel change authorized.

## Owner Mac review (launcher requires separate restart approval)

```sh
cd ~/projects/mauristocks
git status -sb
git fetch origin
git checkout feat/22-session-aware-activity
git pull --ff-only origin feat/22-session-aware-activity
go test -count=1 ./...
bash ./devbuild.sh
./mauristocks -version
./stocksgo.sh
```

Do not run the last command until approving a restart of MauriStocks and its tunnel. First inspect local changes and consider an isolated worktree to preserve the owner's checkout. Safari tests must verify Homepage named/guest launch, guest-only browser settings, revoked/expired modal, one user's multiple tabs counted once, and unchanged market/watchlist behavior.
