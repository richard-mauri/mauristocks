# Current work — MauriStocks

**Verified baseline:** GitHub merged `main` `de8c19aac48a25e8b8af787c747bf936aeba6e37`; [Issue #33](https://github.com/richard-mauri/mauristocks/issues/33), branch `feat/33-account-settings-diff`, [draft PR #34](https://github.com/richard-mauri/mauristocks/pull/34) targeting main. Verify latest head SHA and exact-head Go CI before handing off.

**Current owner-required phase:** Explicit Save Changes for refresh frequency, columns, symbols and all persisted preferences, replacing automatic background sync. UI interactions now stage changes in memory rather than localStorage or Homepage; Save Changes is enabled for changes and submits one browser commit and (if signed in) one authenticated Homepage PUT. A failed or declined remote save is retryable via Save Changes with the browser copy retained; beforeunload warns for still-pending edits. Save As New / confirmed Delete List remain explicit commit actions. Default built-in list is protected. Earlier browser/account review corrections (matched named lists, single desktop heading row, narrow-screen labels and fixed action footer) remain integrated in this same PR.

**Latest owner acceptance:** In Safari Build 000098 the owner observed two simultaneously visible success messages—`Settings saved in this browser.` beside the buttons and `Settings saved to your account.` below. The draft now routes local and server confirmations to one `#watchlist-status` next to Save Changes; removes `#suite-sync-status`; uses combined `Settings saved to browser and account.` after authenticated success, with distinct unsaved/saving/browser-only/failed-retry messages. Market/list selection no longer blanks status. UI-only change, no new PUT path or data mutation. Latest exact-head CI and owner Safari testing still required.

**Data safety:** No account data, browser state, private preferences, tokens, local owner files, build counters, Homepage code, Cloudflare/tunnels, processes or live service changed by this development. Owner Mac-local uncommitted/unpushed state unavailable. No schema change or migration of account records.

**Documentation:** Homepage merged-main `SUITE_DEVELOPMENT_POLICY.md` read; MauriStocks `AGENTS.md`, README, ARCHITECTURE, DEVELOPMENT and CURRENT_WORK reviewed; updated behavior documented in README, ARCHITECTURE, DEVELOPMENT and CURRENT_WORK. AGENTS unchanged.

**Verification:** Go/Node CI on exact latest PR head required; JavaScript regression test covers staging vs persistence, and Go tests now assert explicit save and no automatic save dispatcher. Native Mac build, Safari and verified account PUT remain unverified until owner tests. The earlier live CORS preflight HTTP 204 establishes only permission for OPTIONS, not actual authenticated PUT acceptance.

**Owner Mac handoff after CI passes:** `cd ~/projects/mauristocks`; `git status -sb` (STOP if dirty); `git fetch origin`; `git checkout feat/33-account-settings-diff`; `git pull --ff-only origin feat/33-account-settings-diff`; `./devbuild.sh`; `./mauristocks -version`. The dedicated `./stocksgo.sh` launcher restarts app and Cloudflare tunnel and must not be executed without separate owner approval.

**Promotion:** PR #34 remains draft; no FINISH, merge, Issue closure, remote branch deletion, restart, deploy, release, tag or private build-counter change authorized.
