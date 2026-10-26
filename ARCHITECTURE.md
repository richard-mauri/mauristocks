# MauriStocks architecture

MauriStocks is currently a single-file Go service: [main.go](main.go) contains the HTTP server, Yahoo Finance request/cache and quote transformation logic, and an embedded HTML/CSS/JavaScript dashboard. [go.mod](go.mod) declares the standard-library-only module. This document describes current ownership, not a completed package refactor.

## Responsibilities

The server handles network retrieval, caching, validation, quote/session calculation and API routes. The embedded browser client handles watchlists, column show/hide preferences, sort selection, flags, refresh controls, symbol research links and DOM rendering. Browser state is localStorage, not a server-side database. The quote table sizes visible columns from content (small horizontal padding); range graphs have a minimum usable width and long names may be ellipsized. No draggable/manual width state is used. Regular-session conditions still suppress the extended-price column.

The browser must not receive API secrets. `FINNHUB_API_KEY` is a server-side optional setting. Yahoo Finance interfaces used here are unofficial and subject to upstream availability. Retain loopback binding by default.

## Boundaries and changes

Keep UI-only adjustments inside the embedded markup/script without changing quote data semantics or endpoints. Changes to retrieval, caching or pricing require independent review and regression coverage. A future package split should place UI, HTTP transport, provider integration and quote/domain calculation behind clear interfaces, rather than moving code merely to create directories. See [DEVELOPMENT.md](DEVELOPMENT.md) for verification.