# Current work — MauriStocks

**Baseline:** Verified merged `main` at `de8c19aac48a25e8b8af787c747bf936aeba6e37` (PR #32 merged). The former shared-banner Issue #31 / PR #32 draft handoff was superseded by that merge.

**Active:** [Issue #33](https://github.com/richard-mauri/mauristocks/issues/33), branch `feat/33-account-settings-diff`, [draft PR #34](https://github.com/richard-mauri/mauristocks/pull/34) targeting `main` (`de8c19a`). The latest head and CI must be verified live after each push.

**Candidate:** Owner Safari testing exposed unreadable horizontal scrolling, misleading “No unique symbols,” and a redundant Candidates / Last saved symbols diff after removing `DRAM` and clicking Save Changes. The dialog now uses narrow-screen-safe vertically stacked cards, explicitly identifies symbols absent or present on each side, and suppresses the legacy duplicate only when it matches the active named-list difference. Explicit Save Changes now starts account persistence immediately rather than waiting for the preference debounce, while UI feedback distinguishes local save, pending account save, confirmed account save, browser-only mode, and errors/declined overwrite. Existing full localStorage replacement/reload, rollback, server revision/overwrite guards and account schema are unchanged.

**Safety:** Fresh clean scratch clone of merged `main` used for work; owner Mac-local uncommitted/unpushed state remains inaccessible. No live watchlists, accounts, tokens, saved data, build counters, environment secrets, listeners, processes or tunnels accessed or modified. No services restarted.

**Verification:** Node syntax and simulated Candidates/DRAM comparison passed in scratch; `git diff --check` required at commit. Go tests/build are unverified here because `/usr/bin/go` returns `Go: Unknown option: test`. GitHub exact-head CI and fresh owner Mac/Safari acceptance remain required; consult the live PR for run and status. Earlier owner screenshot verified the original horizontal-scrolling defect, not the fix.

**Owner handoff:** From `~/projects/mauristocks`, inspect `git status -sb` and stop if dirty, fetch, checkout and fast-forward `feat/33-account-settings-diff`, then run only `./devbuild.sh` (already tests and builds) and `./mauristocks -version`. `./stocksgo.sh` restarts the app and tunnel and requires separate authorization. Follow the Safari comparison checklist in `DEVELOPMENT.md`.

**Promotion:** Draft PR only. No FINISH, merge, Issue closure, remote branch deletion, deployment, tag or release authorized.
