# Current work — MauriStocks

**Active documentation review:** [Issue #24](https://github.com/richard-mauri/mauristocks/issues/24), branch `docs/24-session-markdown-refresh` targeting `main`, [draft PR #25](https://github.com/richard-mauri/mauristocks/pull/25). This branch changes only Markdown; exact-head CI and owner merge approval are pending.

GitHub `main` at `934415c` includes merged Issue #22 / PR #23 for Homepage-verified named/guest sessions and activity. No issues or PRs remain open and no implementation branch is active. The owner accepted Safari guest sign-in, revocation and presentation.

MauriStocks validates the Homepage child bearer for guest/named status. Guest settings remain browser-local; named-user Homepage-backed settings remain isolated and revision-controlled. Optional telemetry verifies bearers server-side, deduplicates tabs by pseudonymous parent session, and reports aggregate/private counts without leaking credentials. Invalid/revoked identity gets 401; verification outage gets 503. The sign-in-required dialog is non-dismissible and offers a deliberate Homepage action. Quote retrieval, heatmap, index calculations, watchlists and local persistence are unchanged.

**Next:** reconcile Markdown on a documentation-only branch and draft PR from `main`. Validate paths/links and latest-head CI; no Mac launcher or Safari retest is needed for docs alone. Future runtime acceptance and any merge, deploy, version change or restart remain owner-controlled. Recheck GitHub and unpushed owner-local changes before further work.
