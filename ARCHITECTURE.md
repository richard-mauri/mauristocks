# MauriStocks architecture

MauriStocks is currently a single-file Go service: [main.go](main.go) contains the HTTP server, Yahoo Finance request/cache and quote transformation logic, and an embedded HTML/CSS/JavaScript dashboard. [go.mod](go.mod) declares the standard-library-only module. This document describes current ownership, not a completed package refactor.

## Responsibilities

The server handles network retrieval, caching, validation, quote/session calculation and API routes. The embedded browser client handles watchlists, column show/hide preferences, sort selection, flags, refresh controls, symbol research links and DOM rendering. Browser state is localStorage, not a server-side database. The quote table sizes visible columns from content (small horizontal padding); range graphs have a minimum usable width and long names may be ellipsized. No draggable/manual width state is used. Regular-session conditions still suppress the extended-price column.

The browser must not receive API secrets. `FINNHUB_API_KEY` is a server-side optional setting. Yahoo Finance interfaces used here are unofficial and subject to upstream availability. Retain loopback binding by default.

## Boundaries and changes

Keep UI-only adjustments inside the embedded markup/script without changing quote data semantics or endpoints. Changes to retrieval, caching or pricing require independent review and regression coverage. A future package split should place UI, HTTP transport, provider integration and quote/domain calculation behind clear interfaces, rather than moving code merely to create directories. See [DEVELOPMENT.md](DEVELOPMENT.md) for verification.
## Return to Homepage navigation
MauriStocks's browser UI exposes a visible native anchor `<a href="https://home.themauris.org/">← Back to Homepage</a>` with no `target` and no JavaScript history manipulation. It navigates in the current tab and can be used regardless of account sign-in. The link adds no new server endpoints, exposes no credentials, and does not modify app-local state. This navigation is part of the merged five-service Homepage suite; the card links navigate in the same tab.

The suite-wide return button uses the same `.suite-home-link` presentation in all four apps: 38px minimum height, 8px/12px padding, 10px radius, muted border at rest, accent border and darker fill on hover, and a visible keyboard focus outline. Place one link at the upper-left above the primary content/heading.

The shared suite navigation style is a quiet 12px text link below each page's primary heading, with subtle hover background and underline and keyboard focus outline. There is no app-specific Renew button.

## Twelve-hour Homepage sign-in handoff
MauriStocks uses a transfer bearer linked to Homepage's single twelve-hour session. The account header displays identity only, without a countdown; its internal expiry handling remains unchanged. The bearer cannot remain valid after Homepage sign-out/session expiry. No separate app token renewal UI exists. On a Homepage `#suite_mode=local` launch, MauriStocks clears its cached bearer and enters browser-only mode without deleting watchlists or browser preferences. A valid `#suite_token=...` launch replaces it; an ordinary refresh can reuse a previously delivered bearer subject to server validation. On direct HTTPS visits without a bearer, MauriStocks navigates to the fixed Homepage `/launch/stocks` bridge: Homepage's own same-origin code uses its retained session ID or host-only cookie to POST a new parent-bound handoff, then returns `#suite_token=...` or `#suite_mode=local`. Explicit local mode suppresses further recovery only on that visit, preventing loops; later visits check Homepage again. Cached expiry triggers bounded profile revalidation before discarding a bearer, since the Homepage parent may already have been renewed. No new local login, personal token retention, or app renewal operation is introduced. These are coordinated with Homepage Issue #24. The matching Homepage PR #12 and NewsGo PR #53 are merged; the owner accepted coordinated Safari behavior.
