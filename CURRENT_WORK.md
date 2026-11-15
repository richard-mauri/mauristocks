# Current work — MauriStocks

**Active:** Homepage [Issue #45](https://github.com/richard-mauri/homepage/issues/45) policy adoption on `docs/45-suite-policy-adoption` from GitHub `main`; [draft PR #26](https://github.com/richard-mauri/mauristocks/pull/26) open against `main`. Documentation-only entry points to Homepage's single authoritative policy; generated copies and synchronization checks removed. No consumer merge authorized.

**Merged context:** Homepage-verified guest/named sessions, per-user revision-controlled settings, pseudonymous optional activity and non-dismissible sign-in are implemented. Invalid/revoked identity returns 401 and verification outage 503. Quote retrieval, heatmaps, watchlists, responsive columns, stars, sorting, provider behavior and local browser data are outside this scope.

**Next:** validate Homepage policy link and docs, confirm exact draft PR head and CI, and coordinate owner review with NewsGo, ScraperGo and SailingGo PRs. No Mac build/restart or Safari retest for docs-only work. Check local uncommitted/unpushed work before future runtime changes; no deployment, tag, release, version bump or branch deletion.
