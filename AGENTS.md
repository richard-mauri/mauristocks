# MauriStocks — agent instructions

Read [CURRENT_WORK.md](CURRENT_WORK.md), [ARCHITECTURE.md](ARCHITECTURE.md) and [DEVELOPMENT.md](DEVELOPMENT.md) before development. Check actual GitHub branch, open PR and source; the handoff snapshot may be stale. Explain the active baseline, change scope, CI, owner-side Mac/Safari acceptance and explicit no-merge rule before edits.

Keep this file limited to permanent agent obligations and handoff maintenance. Product instructions belong in [README.md](README.md), implementation ownership in [ARCHITECTURE.md](ARCHITECTURE.md), and branch/release procedure in [DEVELOPMENT.md](DEVELOPMENT.md).

Preserve quote retrieval, provider assumptions, browser watchlists, sorting, symbol links, stars, session-specific extended prices, local environment and API contracts unless directly in scope. Review affected code before changing it. Run Go tests/build and check GitHub Actions at the exact PR head; label any unavailable checks unverified. CI does not prove Safari layout. Never merge, deploy, tag, release, restart the owner's service or change semantic version without explicit owner approval.

Maintain [CURRENT_WORK.md](CURRENT_WORK.md) as a concise portable snapshot of active branch/PR, technical decisions, evidence, blockers and next steps. Replace obsolete state instead of appending a dated history. Before PR handoff update its description to reflect actual code and test status.