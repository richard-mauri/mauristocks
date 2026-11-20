# Current work — MauriStocks

**Baseline:** Verified merged `main` at `de8c19aac48a25e8b8af787c747bf936aeba6e37` (PR #32 merged). The former shared-banner Issue #31 / PR #32 draft handoff was superseded by that merge.

**Active:** [Issue #33](https://github.com/richard-mauri/mauristocks/issues/33), branch `feat/33-account-settings-diff`, [draft PR #34](https://github.com/richard-mauri/mauristocks/pull/34) targeting `main` (`de8c19a`). The latest head and CI must be verified live after each push.

**Candidate:** Show side-by-side, read-only differences between browser settings and the saved Homepage account before restoring. Named-watchlist differences expose symbol additions/removals, and preferences use human-readable labels. Keep Browser Settings retains both snapshots; Restore Account Settings retains the existing full localStorage replacement/reload and rollback. No Homepage endpoint/schema, session, server-side settings, quote or heatmap change.

**Safety:** Fresh clean scratch clone of merged `main` used for work; owner Mac-local uncommitted/unpushed state remains inaccessible. No live watchlists, accounts, tokens, saved data, build counters, environment secrets, listeners, processes or tunnels accessed or modified. No services restarted.

**Verification:** `git diff --check` and Node syntax checking of the embedded JavaScript passed in scratch. Go tests/build are unverified here because `/usr/bin/go` returns `Go: Unknown option: test`. GitHub exact-head CI and owner Mac/Safari acceptance remain required; consult the live PR for run and status.

**Owner handoff:** From `~/projects/mauristocks`, inspect `git status -sb` and stop if dirty, fetch, checkout and fast-forward `feat/33-account-settings-diff`, then run only `./devbuild.sh` (already tests and builds) and `./mauristocks -version`. `./stocksgo.sh` restarts the app and tunnel and requires separate authorization. Follow the Safari comparison checklist in `DEVELOPMENT.md`.

**Promotion:** Draft PR only. No FINISH, merge, Issue closure, remote branch deletion, deployment, tag or release authorized.
