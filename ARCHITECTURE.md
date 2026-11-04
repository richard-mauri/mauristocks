# MauriStocks architecture

MauriStocks is currently a single-file Go service: [main.go](main.go) contains the HTTP server, Yahoo Finance request/cache and quote transformation logic, and an embedded HTML/CSS/JavaScript dashboard. [go.mod](go.mod) declares the standard-library-only module. This document describes current ownership, not a completed package refactor.

## Responsibilities

The server handles network retrieval, caching, validation, quote/session calculation and API routes. The embedded browser client handles watchlists, column show/hide preferences, sort selection, flags, refresh controls, symbol research links and DOM rendering. Browser state is localStorage, not a server-side database. The quote table sizes visible columns from content (small horizontal padding); range graphs have a minimum usable width and long names may be ellipsized. No draggable/manual width state is used. Regular-session conditions still suppress the extended-price column.

The browser must not receive API secrets. `FINNHUB_API_KEY` is a server-side optional setting. Yahoo Finance interfaces used here are unofficial and subject to upstream availability. Retain loopback binding by default.

## Boundaries and changes

Keep UI-only adjustments inside the embedded markup/script without changing quote data semantics or endpoints. Changes to retrieval, caching or pricing require independent review and regression coverage. A future package split should place UI, HTTP transport, provider integration and quote/domain calculation behind clear interfaces, rather than moving code merely to create directories. See [DEVELOPMENT.md](DEVELOPMENT.md) for verification.
## Return to Homepage navigation
MauriStocks's browser UI exposes a visible native anchor `<a href="https://home.themauris.org/">← Back to Homepage</a>` with no `target` and no JavaScript history manipulation. It navigates in the current tab and can be used regardless of account sign-in. The link adds no new server endpoints, exposes no credentials, and does not modify app-local state. This is coordinated with Homepage Issue #11 / draft PR #12, where card clicks now navigate in the same tab.

The suite-wide return button uses the same `.suite-home-link` presentation in all four apps: 38px minimum height, 8px/12px padding, 10px radius, muted border at rest, accent border and darker fill on hover, and a visible keyboard focus outline. Place one link at the upper-left above the primary content/heading.

The shared suite navigation style is now intentionally a quiet 12px text link directly below each page's primary heading, with subtle hover background and underline and keyboard focus outline (rather than the former 38px bordered button). NewsGo does not render an additional signed-out Homepage action; only the account Renew control appears during an active token session.

## Twelve-hour Homepage sign-in handoff
MauriStocks uses a short-lived transfer bearer linked to Homepage's single twelve-hour session. The bearer cannot remain valid after Homepage sign-out/session expiry. No separate app token renewal UI exists. On a Homepage `#suite_mode=local` launch, MauriStocks clears its cached bearer and enters browser-only mode without deleting watchlists or browser preferences. A valid `#suite_token=...` launch replaces it; an ordinary refresh can reuse a previously delivered bearer subject to server validation. Coordinated with Homepage draft PR #12; do not deploy independently prior to integrated Safari acceptance.
