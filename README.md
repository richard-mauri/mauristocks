# Mauri's Stock Quotes (MauriStocks)

**Version 1.0.0** (development build ID is generated locally, not a release number).

A small, self-hosted stock-watchlist dashboard written in Go. It emphasizes a clear, compact table of regular-session and available extended-hours prices without heavyweight charts or advertising.

**Source:** https://github.com/richard-mauri/mauristocks  
**Local URL (default):** http://localhost:8082/

> **Developers:** See [AGENTS.md](AGENTS.md), [ARCHITECTURE.md](ARCHITECTURE.md), [DEVELOPMENT.md](DEVELOPMENT.md) and [CURRENT_WORK.md](CURRENT_WORK.md). The README covers product features and usage.

## Features

MauriStocks displays stock symbols, exchanges, company names, regular-session prices, premarket/after-hours prices when available, changes, and graphical day/52-week price ranges. Click a column header to sort; a visual indicator identifies the active sort. The extended-hours column disappears during regular trading and returns during the applicable extended session when supported by provider data.

The **Price** column represents regular-session pricing, while **Premarket Price** or **After-hours Price** is shown separately. Extended-session Change $ and Change % compare the extended price against the adjacent regular price; regular-session changes compare the regular price with the previous close. Price coloring may indicate movement between distinct timestamped observations; extended-price coloring compares that displayed price with the regular price.

Use the permanent **Default** watchlist, create and edit named personal lists, or choose **Most Active — Top 20 (Dynamic)**, which is ranked by share volume when Yahoo's unofficial screener data is available. Personal watchlists, selected list, sort choices and refresh preferences are saved automatically in the current browser, including when signed out. Edits to named lists save after 600 ms of inactivity or on field change; Save As New creates a new personal list, and the Default and dynamic lists remain fixed. Flags are separate: clicking ☆ turns it into ★ and highlights the ticker's entire row. Browser-only changes remain local until sign-in. While signed in, user settings are also automatically saved to the corresponding Homepage account, with existing overwrite confirmation and revision-conflict safeguards.

You can refresh quotes manually or choose Never, 5 seconds, 30 seconds, 1 minute or 5 minutes. The page loads the selected list on opening. Automatic requests pause when the page is hidden. Very short intervals may trigger provider rate limits.

Click a ticker to open the selected research destination: Google Finance, Stock Analysis, a known company website or an investor-relations page when a verified destination is available. Company website discovery can optionally be enabled using Finnhub.

A **Columns** menu inside the Table panel (and hidden in Heatmap view) lets you show or hide individual columns and reset their visibility. Visible columns automatically fit their content with modest padding; manual resizing is not necessary. Column visibility preferences are saved in this browser (previously saved manual widths are ignored). The Symbol column stays visible, and extended-session pricing still disappears automatically during regular market sessions. The Symbol links destination selector is in the shared toolbar alongside Market and View.

An **About / Data Notes** section expands for more explanation and is collapsed by default on phones and computers. The table may scroll horizontally on a narrow screen.

## Market index selections (draft, PR #15)

The shared **Market** dropdown offers **My Watchlists**, **S&P 500 · Top 30**, **NASDAQ-100 · Top 30**, and **Dow Jones · 30**. Each selection works in both **Table** and **Heatmap** view. Selecting an index temporarily makes the personal watchlist controls read-only; switching back restores the previous personal watchlist without replacing or saving its symbols.

Top 30 means the 30 constituents with the largest available **USD market capitalization**, not the entire index. For Dow Jones, all 30 member tickers are retained and ordered by capitalization even if a weight cannot be retrieved; Dow's actual index is price-weighted. Heatmap tile area consistently represents available **market capitalization**, not official index weight. If a capitalization is unavailable the company remains unweighted rather than receiving an invented number.

The Go server obtains S&P membership from zyhe16's regularly updated S&P list, NASDAQ-100/Dow membership from yfiua's generally monthly index constituent snapshots, and market capitalization from zyhe16's Nasdaq screener-based daily dataset. These are unofficial third-party snapshots, **not real-time or guaranteed current**. The server refreshes and shares them across market selections on a six-hour TTL; the browser reuses the included capitalization values immediately, avoiding Finnhub's sequential per-symbol delay on market selections. A temporary upstream outage may use an already cached snapshot. Quotes still come from the existing Yahoo endpoint and the current refresh schedule. Large watchlists at very short refresh intervals can trigger unofficial quote-provider limits.

**Local validation:** after checkout, use `bash ./devbuild.sh` in the original `~/projects/mauristocks` repository. Do not restart `stocksgo.sh` or the tunnel without approval.

## Automatic watchlist and account persistence

MauriStocks has no manual **Save Changes** or **Load from account** controls. Named personal watchlists save automatically in the browser after 600 ms of typing inactivity, on field change, before changing watchlists/markets, and on quote refresh. **Save As New** remains available to name a new list. When a user is signed in, watchlists and all changed settings are automatically synchronized with the personal Homepage account. At sign-in, the app compares saved settings with the browser; if they differ, restoration requires explicit confirmation. Existing server overwrite consent and revision-conflict safeguards remain in place. On unsigned localhost, changes save locally without server calls.

## Heatmap adjustable tile weighting

The Heatmap view offers a **Tile weighting** range slider from **0.50 (balanced)** to **1.00 (exact market-cap proportions)**, default **0.65**. Tile area is proportional to `market_cap ^ exponent`; this preserves capitalization ranking while making smaller tiles more readable. A value of 1.00 is the original exact market-cap-area sizing. At values below 1, tile areas are **adjusted for readability**, not literal index weights. The setting is saved locally under `mauristocks_heat_weight_v1` and, when signed in, is included in the user's per-account Homepage settings record. It affects both personal-watchlist and index heatmaps. It does not change market-cap amounts shown in tooltips, price-change coloring, quote refresh, or the market-cap retrieval APIs. Moving the slider recomputes layout immediately without fetching data.

## Watchlist heatmap (Issue #13)

Use the theme-matched **View** dropdown (Table or Heatmap) to switch views of the currently selected watchlist. The heatmap sits in a recessed market panel with a muted slate-teal surround, a fine border, inset depth, a watchlist title and concise coverage status. These styles use existing theme variables and do not change the red/green price colors or shared palette. The heatmap sizes company tiles in proportion to market capitalization and colors them by **regular-market price versus previous close**, independently of the table's active-session Change %. The browser measures actual tile pixel dimensions and rendered text widths to show the symbol and percentage change when each fits; even small tiles may show a ticker if space permits. Tile details appear in a readable hover/focus popup (first tap on touch devices) showing capitalization, regular price, change and quote time; tapping again follows the selected research link. Tiles open the existing selected research destination. Symbols without valid capitalization (including ETFs and some unsupported tickers) appear in a separate **Unweighted / unavailable** group rather than receiving invented weights.

Capitalization arrives in small batches so weighted tiles appear progressively on first use. A per-symbol browser cache reuses capitalization across overlapping watchlists and test-server restarts (12 hours for valid capitalization, one hour for unavailable values); it stores no API credentials. Previously cached watchlist-level results are migrated automatically. The heatmap follows the application's unchanged shared light/dark palette and cyan accents, with faint top-right/bottom-left background glows; it never displays server credential instructions to browser users.\n\nThe heatmap reuses the existing `/api/quotes` responses and the same Never / 5 seconds / 30 seconds / 1 minute / 5 minutes refresh control. It does **not** start a separate quote timer or claim exchange-level real-time updates. Only tile values and colors repaint during normal quote refreshes; tile area is recomputed if capitalization/selected symbols change. Automatic requests still pause while the page is hidden.

Weighted tiles require **`FINNHUB_API_KEY` configured server-side**. The independent `/api/market-caps?symbols=NVDA,AVGO` endpoint requests Finnhub Company Profile 2 metadata, validates the ticker and USD currency, converts the provider's USD-million capitalization to USD, rate-limits upstream requests, and caches successful values for 24 hours and unsupported values for one hour. It may be slow during an initially uncached watchlist. Missing key, coverage, quota or provider failures leave the quote table usable but may prevent weighted tiles from displaying. Do not embed the Finnhub key in browser code. This feature has not yet received owner Mac/Safari acceptance.

## Requirements and local use

You need Go, a current browser and upstream network access. The program uses the Go standard library and does not require a database or third-party Go module. From the documented local repository folder:

```sh
cd ~/projects/mauristocks
bash ./devbuild.sh
./mauristocks -version
./stocksgo.sh
```

Then open http://localhost:8082/. Use your existing launcher stop procedure when applicable. The local `stocksgo.sh` script is maintained outside this repository and must be inspected before changing its behavior. For command-line help, use `./mauristocks -h`; `-port`, `-host`, `-version` and `-debug` are available. By default the service binds to `127.0.0.1`; exposing it on a network interface should only be done with appropriate access controls.

To optionally look up company websites, configure `FINNHUB_API_KEY` as a **server environment variable**. Do not put the key in source code or browser storage. The service also uses curated website mappings when available.

## HTTP endpoints

| Endpoint | Purpose |
| --- | --- |
| `/` | Browser quote dashboard |
| `/api/quotes` | Default symbols or `?symbols=MU,NVDA,AVGO` |
| `/api/most-active` | Dynamic volume-ranked list when provider access succeeds |
| `/api/market-caps?symbols=NVDA,AVGO` | Independent cached capitalization metadata (requires `FINNHUB_API_KEY`) |
| `/api/company?symbol=MU` | Resolved company website metadata |
| `/company?symbol=MU` | Redirect to resolved company website when known |
| `/healthz` | Health response |

For example, `curl 'http://localhost:8082/api/quotes?symbols=MU,NVDA'` requests two quotes.

## Market-data notes

Quotes are retrieved through Yahoo Finance's **unofficial, undocumented** chart endpoint. The dynamic Most Active ranking uses Yahoo's unofficial screener. Their formats, rate limits, terms and availability can change, and returned data may be delayed or incomplete. An available intraday bar is not necessarily a real-time trade. The service uses short quote caching and longer range/most-active caches; caching does not make delayed data live.

The displayed retrieval time indicates when the service fetched information, not when the exchange last traded. Quote timestamps and retrieved times are displayed in Pacific time. Range indicators summarize provider data and may not include extended-session extremes.

This tool is for personal informational use, not trading execution or investment advice. For engineering changes see [DEVELOPMENT.md](DEVELOPMENT.md); for active handoff status see [CURRENT_WORK.md](CURRENT_WORK.md).

## Build identifiers

Run `bash ./devbuild.sh` for local builds. Following ScraperGo, it tests first, reserves the next private sequential number in Git's directory (`git rev-parse --git-path mauristocks-build-counter`), then embeds a zero-padded `Build NNNNNN` label in the executable and page. An absent counter starts after historical v40, at Build 000041. Failed compilation after reservation does not reuse a number. This counter belongs to the local clone and is not a global or release version. Direct `go build` shows `Build dev`; semantic version stays 1.0.0. Do not commit the counter or manually edit labels in HTML.

### Homepage account settings (current)

Homepage is the sign-in authority for the suite. Launch MauriStocks from a signed-in Homepage card to receive an app handoff bearer bound to that single 12-hour Homepage session. The app removes the URL fragment and keeps the bearer in tab-scoped sessionStorage, revalidating it with Homepage; sign-out or expiry invalidates access. The account header shows only **Signed in as [username]** or **Not signed in**, without a separate countdown or Renew button. A direct MauriStocks visit without a tab handoff now briefly visits Homepage's fixed `/launch/stocks` session recovery page. Homepage revalidates its existing durable 12-hour session using its own retained opaque session ID or host-only cookie, then returns a fresh parent-bound bearer to Stocks in the URL fragment. If the session has expired, was revoked, or does not exist, MauriStocks remains browser-only with saved watchlists/preferences unchanged; no automatic sign-in or renewal occurs. A later direct visit can check Homepage again. An already-present bearer is revalidated with Homepage at its cached expiration before it is discarded, since Homepage may have renewed its parent. Account load/save uses Homepage's per-user storage and server-enforced roles. All user-editable persisted preferences—named watchlists, selected watchlist, flags, symbols, sorting, refresh frequency, research destination, visible table columns, Market, View and heatmap Tile weighting—are included in the same per-user settings record. Changes to Market, View and Tile weighting automatically trigger the existing debounced account save while signed in; offline/browser-only changes continue to be retained locally. The first account overwrite requires explicit approval if saved account data already exists, and a conflicting update from another device is rejected rather than silently overwritten. Quote and market-cap metadata caches are not synchronized. Homepage's storage supports the existing schema-1 JSON envelope without requiring changes to the server.

## Optional suite activity monitoring (Issue #19 draft)

MauriStocks can optionally report **aggregate, advisory activity** to Homepage's separate administrator collector. Configure `STOCKS_ACTIVITY_URL` to the explicit Homepage ingestion endpoint and `HOMEPAGE_ACTIVITY_KEY_STOCKS` to a unique server-only secret of at least 32 characters. The same secret must be configured for the `stocks` service on the Homepage collector. Without both variables, monitoring is disabled and no browser heartbeat is loaded. Never put the secret in JavaScript, saved preferences, source control or a URL. The reporter accepts HTTPS or loopback-only HTTP collector URLs, rejects redirects, and has a short network timeout.

With reporting enabled, the browser submits random per-tab presence identifiers to same-origin `POST /api/activity/ping`. The server aggregates anonymous or unverified observed tabs active over 2/5/15/60-minute windows, foreground tab observations, coarse interactions and in-flight quote, market-cap, market-top, Most Active and company-lookup requests. It reports aggregates every 60 seconds using service-specific server-side authorization and a nonce. **No portfolio symbols, price observations, personal watchlists, account identities, visitor identifiers or authorization tokens leave MauriStocks as telemetry.** A signed-in tab is still unverified at this stage and included in the anonymous count; a future separate step would require server-side Homepage identity verification. An idle period or disconnected collector is not evidence that a restart is safe. The integration is draft-only, with no runtime deployment authorized.

### Isolated test environment (no production restart)

Start from a separate worktree of `origin/feat/19-suite-activity` and launch the test binary on `127.0.0.1:8086` using `-host 127.0.0.1 -port 8086`, without `stocksgo.sh`. Point `STOCKS_ACTIVITY_URL` to `http://127.0.0.1:8788/api/admin/activity/ingest`, and configure a fresh matching `HOMEPAGE_ACTIVITY_KEY_STOCKS` in both that test process and the isolated Homepage collector running with `SUITE_DATA_DIR` and `LISTEN_ADDR=127.0.0.1:8788`. Do not use production data or tunnel configuration. Keep a local stocks tab open, then after 75 seconds refresh the Homepage admin Session Activity panel: the MauriStocks card should report fresh counts. This test has not been performed on the owner's Mac yet and must not be described as acceptance.
