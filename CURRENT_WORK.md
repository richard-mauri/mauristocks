# Current work — MauriStocks

**Active:** [Issue #31](https://github.com/richard-mauri/mauristocks/issues/31), branch `ui/31-shared-banner`, [draft PR #32](https://github.com/richard-mauri/mauristocks/pull/32) targeting merged `main` `bf7be70b06203fe1bbbe7aceb5927e5e074e9feb`. Homepage shared foundation draft PR #55 remains open; MauriStocks pins immutable module `v0.0.0-20261120172629-80deb3b13f6a`.

**Candidate:** Replace only the custom masthead with the common, viewport-aligned suite banner and Homepage's exact original two blurred teal ambient shapes. Preserve the app title and executable version/build, same-tab Homepage link, and verified Homepage identity via the existing `suite-auth-status` JavaScript hook. Tanner JPEG is embedded and served locally. The shared button contract and the stock content inside `.stocks-content` remain independent; quote retrieval, sorting, watchlists, heatmap, dialogs, persistence and session recovery are unchanged.

**Safety:** Fresh clean scratch clone of merged `main` used for work; owner Mac-local uncommitted/unpushed state remains inaccessible. No watchlists, accounts, tokens, saved data, build counters, environment secrets, listeners, processes or tunnels accessed or modified. No services restarted.

**Verification:** Regression tests cover shared HTML/CSS, original ambient shapes, local Tanner, same-tab return and existing account updates. GitHub Go CI [44786390872](https://github.com/richard-mauri/mauristocks/actions/runs/44786390872) passed on implementation head `4d6037d52b701e6ae7a8e4fef49cc275abc4a177`; recheck exact-head CI after this documentation commit. This workspace's `/usr/bin/go` is not a functioning compiler, so local Go test/build unverified. Owner Mac/Safari dark/light, responsive widths, quotes, heatmap, watchlists and account behavior remain unverified.

**Owner handoff:** From `~/projects/mauristocks`, inspect `git status -sb` and stop if dirty, fetch, checkout and fast-forward `ui/31-shared-banner`, then run only `./devbuild.sh` (already tests and builds) and `./mauristocks -version`. `./stocksgo.sh` restarts the app and tunnel and requires separate authorization.

**Promotion:** Draft PR only. No FINISH, merge, Issue closure, remote branch deletion, deployment, tag or release authorized.
