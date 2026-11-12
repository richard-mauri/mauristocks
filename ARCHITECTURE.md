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

## Watchlist heatmap (Issue #13 draft)

The dashboard now has Table and Heatmap views over the same in-memory `quotes` array and existing `refreshQuotes()` timer. A successful quote refresh repaints the heatmap; switching views does not start a second quote poller. The heatmap uses regular-market price relative to previous close for color, rather than reusing the table's session-dependent change field. Area comes from independent company capitalization metadata, not share price. Tiles reuse `researchURL()` and the browser's selected research destination.

`/api/market-caps` uses Finnhub Company Profile 2 with a server-only `FINNHUB_API_KEY`. The response includes USD capitalization, source and retrieval timestamp where available. Requests are serialized and rate-spaced independently of quote retrieval. Successful metadata is cached for 24 hours; missing-cap metadata for one hour. This is an additive endpoint, not a change in the existing `/api/quotes` contract. Unavailable securities are displayed separately, never given invented market-cap weights. Metadata failures must not interrupt quote updates. Browser-only view preference uses `mauristocks_view_v1`; existing Homepage account settings keys and their per-user isolation are unchanged.

The first version uses a small alternating-axis area-preserving binary treemap with no JavaScript dependency. No major-index universe, sector map, historical change selector, or heatmap export is included. Mac/Safari and provider acceptance remain outstanding.

## Market Top 30 selections (PR #15 draft)

`market_indices.go` implements `/api/market-top?market=sp500|nasdaq100|dow30`, validating upstream CSV membership and market capitalization snapshots, ranking eligible index constituents, returning 30 symbols plus available per-symbol USD caps, and keeping a shared six-hour in-memory snapshot for all three indexes. Sources: zyhe16 daily Nasdaq screener snapshots/S&P membership, yfiua monthly constituent lists for NASDAQ-100 and Dow. These are not real-time, official index weights. Dow members are retained even when an individual cap is unavailable. The browser stores a separate `mauristocks_market_v1` selection, makes watchlist edits unavailable while viewing an index, seeds the existing per-symbol browser cap cache from each server response, and retains the exact same `/api/quotes` and refresh cycle. Account credentials and per-user watchlist storage are unchanged. Provider failures are visible to the user and are not replaced with guessed constituents.

## Automatic watchlist persistence

Named personal watchlists are autosaved to browser localStorage on a debounced symbol-input edit (600 ms), blur/change, refresh or selection change. The user no longer needs a manual Save Changes button. Default and dynamic lists are still immutable templates; new lists are explicitly created through Save As New. Unsigned users need no account access, and their browser storage remains authoritative. Signed-in users retain the existing debounced `/api/storage/mauristocks` account saves and optimistic revision checks. A one-time sign-in account reconciliation compares saved settings with the current browser and requires confirmation before replacing divergent browser settings; existing account-overwrite confirmation remains. Removed the manual Load from account button. No API schema or authentication changes.
