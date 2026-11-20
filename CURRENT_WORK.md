# Current work — MauriStocks

**Baseline:** Verified GitHub merged `main` at `de8c19aac48a25e8b8af787c747bf936aeba6e37` (PR #32 merged). Latest active [Issue #33](https://github.com/richard-mauri/mauristocks/issues/33), [draft PR #34](https://github.com/richard-mauri/mauristocks/pull/34), branch `feat/33-account-settings-diff`, based on `main`. Refresh exact head/CI live after changes.

**Candidate:** Responsive browser/account comparison with named-list symbol deltas and no duplicate selected-symbol mirror. Follow-up Safari acceptance showed a Visible Columns card incorrectly reading `hidden: shown` because the saved `hidden` array was interpreted as a boolean. The branch now compares known hidden-column IDs and displays only actual differences (for example `Name: hidden` versus `Name: shown`); malformed legacy data remain safe plain text. Automatic preference updates now show pending account persistence, and profile verification no longer clears confirmed status. Keeping browser settings explicitly says the account is unchanged. Explicit symbol Save Changes still immediately syncs its committed list; other preferences still synchronize automatically. Full account-snapshot reconciliation, data schema, initial overwrite approval, per-user authentication and revision guards are unchanged.

**Integration:** Homepage [Issue #58](https://github.com/richard-mauri/homepage/issues/58) / [draft PR #59](https://github.com/richard-mauri/homepage/pull/59) corrects Safari CORS storage preflights; user live curl showed HTTP 204 after the Homepage runtime update. That proves preflight access, not a successful authenticated MauriStocks PUT. Keep PR #34 separate and draft pending Safari acceptance.

**Safety:** GitHub scoped branch only; owner Mac-local dirty/unpushed state inaccessible. No accounts, browser settings, watchlists, tokens, local counters, runtime processes, Homepage data, DNS or tunnel settings modified. No release or restart.

**Verification:** `scripts/test-account-diff.cjs` exercises column visibility and status regression, and the Go CI workflow runs it. Check the **latest exact-head CI** and report its outcome before Mac handoff. Owner Safari/account PUT 200 and local build remain unverified for the new candidate. Relevant owner steps are in `DEVELOPMENT.md`.

**Owner handoff:** From `~/projects/mauristocks`, check `git status -sb` (stop if dirty), fetch origin, checkout and fast-forward `feat/33-account-settings-diff`, run `./devbuild.sh` (already tests and builds) and `./mauristocks -version`. Do not run `stocksgo.sh`, which restarts the app and tunnel, without separate authorization. Then check Safari browser/account comparison and verified authenticated PUT.

**Promotion:** PR #34 stays draft. No FINISH, merge, Issue closure, branch deletion, deployment, tag, release or service restart authorized.
