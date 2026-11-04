# MauriStocks — current work (2026-11-04)

Authoritative baseline: GitHub `main`. Active issue [#6](https://github.com/richard-mauri/mauristocks/issues/6), branch `fix/6-homepage-return`, [draft PR #7](https://github.com/richard-mauri/mauristocks/pull/7) targeting main. This is an independent part of Homepage Issue #11 / PR #12, which changes launcher navigation to same-tab. Scope: Adding a styled native same-tab Back to Homepage link in the dashboard header. Quote provider, watchlists, preference storage and optional suite token handoff stay unchanged.

Owner Mac/Safari acceptance of return link, mobile visibility, light/dark contrast and same-tab navigation is pending. Go regression test checks presence of native link. GitHub CI must be checked at the exact final PR head. No merge, release or service restart authorized. Existing local work and Cloudflare configuration are unverified and must not be overwritten.

Other work: existing UI and optional Homepage storage synchronization candidates require independent review; do not merge them into this change.

Owner Safari feedback on navigation PRs: app links worked but the return controls differed in position/hover style, and MauriStocks displayed a second 'Open Homepage' button while signed out. Latest candidate uses identical `.suite-home-link` CSS in all apps and positions the link at top-left above primary heading. MauriStocks retains only the universal return link while signed out; its account Renew button is visible only while signed in. Mac/Safari visual acceptance of this revision is outstanding.
