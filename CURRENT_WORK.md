# Current work — MauriStocks

## Issue #11 — restore persistent Homepage sign-in on direct Stocks visits (November 9, 2026)

- **Merged baseline:** GitHub `main` at `4812d0b553a1bb930bd9945df91f0e11ecae23ca` (documentation PR #10 already merged; previous status on main was stale).
- **Scoped work:** [Issue #11](https://github.com/richard-mauri/mauristocks/issues/11), branch `fix/11-homepage-session-recovery`, draft PR pending, coordinated with Homepage [Issue #24](https://github.com/richard-mauri/homepage/issues/24) / [draft PR #25](https://github.com/richard-mauri/homepage/pull/25).
- **Behavior:** A direct HTTPS Stocks visit without a tab-scoped Homepage handoff navigates through Homepage's fixed `/launch/stocks` origin. Homepage alone validates its already-live durable session via retained opaque ID or host-only cookie and issues a fresh bound child bearer. A missing/revoked/expired login returns `#suite_mode=local` for the current visit without loop or changes to browser watchlists. Later direct visits can recheck. Cached handoff expiry no longer immediately clears the token: Stocks requests Homepage profile revalidation first, throttled during network failure.
- **Preserved:** Identity-only account header; no Stocks Renew/countdown; Homepage per-user storage and server authorization; quote retrieval/cache/sorting/watchlists/flags, local preferences, four independent Cloudflare tunnels and listener defaults.
- **Validation:** Confirm `go test -count=1 ./...`, `go build ./...` and GitHub CI on the exact final head. Mac/Safari, provider quotes, browser-close/reopen, and public Homepage tunnel are unverified.
- **Acceptance:** After explicit owner approval for Homepage and Stocks restarts, deploy Homepage PR #25 first. Sign in to Homepage, open Stocks, close Safari and reopen Stocks directly; verify signed-in identity returns without a new login, unchanged parent expiry and preserved watchlists. Test absent/revoked/expired sessions, browser-only launch, Homepage Renew/Disconnect and existing quote, refresh, sorting and research links.
- **Merge boundary:** Draft PR only, no merge, version bump, release, restart or tunnel change without owner authorization.

## Mac commands

```bash
cd ~/projects/mauristocks
git status -sb
git fetch origin
git checkout fix/11-homepage-session-recovery
git pull --ff-only origin fix/11-homepage-session-recovery
go test -count=1 ./...
go build ./...
bash ./devbuild.sh
./mauristocks -version
./stocksgo.sh
```
