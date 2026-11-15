# MauriStocks — agent entry point

**Mandatory:** Read the complete [generated local suite policy](docs/suite-development-policy.md) before development. It is the generated offline-readable copy of Homepage's [authoritative shared policy](https://github.com/richard-mauri/homepage/blob/main/SUITE_DEVELOPMENT_POLICY.md). Run `python3 scripts/suite_policy.py --offline` to verify local integrity and `python3 scripts/suite_policy.py` when online to check freshness. Do not edit the copy by hand; `--sync` regenerates it for a reviewed PR. Missing/stale shared policy blocks shared-process edits.

Also read [README.md](README.md), [ARCHITECTURE.md](ARCHITECTURE.md), [DEVELOPMENT.md](DEVELOPMENT.md), and [CURRENT_WORK.md](CURRENT_WORK.md); verify live main, PR/head and local changes. The suite policy defines workflow, six-part status/handoff checklist and FINISH. DEVELOPMENT retains MauriStocks-specific CI, build, Mac and Safari instructions.

Preserve quote retrieval/provider assumptions, table columns and responsive widths, sorting, symbol links, stars, browser watchlists, regular/extended-session prices, Homepage identity and revisioned user settings, local environment and API contracts. Check call sites before edits. Never alter the owner's stocks launcher, build counter, tunnel or another service by implication. Do not claim Linux CI proves Safari behavior.
