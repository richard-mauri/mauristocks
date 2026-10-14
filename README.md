# Mauri's Stock Quotes

A lightweight, self-hosted stock watchlist written in Go. It displays regular-session, premarket, and after-hours prices in a compact browser table, without advertisements or a heavyweight charting interface.

**Current application version:** `1.0.0` · **Build:** `v34` (October 2026).

## Features

- **Session-dependent quote table:** The table has nine columns in extended-hours sessions (Symbol, Exchange, Name, Price, Premarket/After-hours, Change, Change %, Day Range, and 52-Week Range) and eight during regular trading, when the entire extended-hours column—including its heading—is hidden. **Price** remains the regular-session price. The extended column reappears for premarket and after-hours observations, with the corresponding heading. Existing columns keep their fixed widths when the extended column is hidden.
- **Session-aware price changes:** During regular trading, Change and Change % measure the regular price versus the previous close. During premarket and after-hours sessions, they measure the extended-session price versus the regular price shown in the adjacent Price column. If extended data is missing, session change values are blank. Go computes extended changes (including `extended_change` and `extended_change_percent` in JSON).
- **Multiple named watchlists:** Create, select, modify, and delete locally saved lists using **Save As New**, **Save Changes**, and **Delete List**. The permanent, read-only **Default (20 symbols)** list contains `MU, AMAT, LRCX, SPCX, AVGO, TSM, TSLA, NVDA, JOBY, PLTR, LMT, AMGN, SNA, FTEC, JEPQ, VOO, IAU, FSTA, IBIT, O`.
- **Most Active — Top 20 (Dynamic):** A read-only watchlist ranked by reported daily **share volume** using Yahoo's unofficial market screener. The ranking is cached server-side for approximately 15 minutes. It depends on provider availability; a failed ranking request reports an error rather than inventing symbols.
- **Immediate initial loading:** The active watchlist is restored and quotes are requested automatically when the page opens, including when Auto Refresh is **Never**.
- **Manual or timed refresh:** Never, 5 seconds, 30 seconds, 1 minute, or 5 minutes. Automatic polling is skipped while the page is hidden; overlapping requests are prevented. Prices remain visible while fetching, and the **Refresh quotes** button keeps its label and dimensions.
- **Sortable table:** Click a column heading to toggle direction. Price, extended-hours, and change columns use numeric comparisons; range columns sort by the displayed price's normalized position between their lows and highs. Sorting is retained on refresh, except that sorting by extended price switches to descending Price sorting when the extended column becomes hidden.
- **Graphical ranges:** Compact day and 52-week low-to-high bars with diamond indicators. Both price labels are placed above their bars.
- **Independent price shading:** The regular **Price** background indicates short-term momentum between distinct timestamped observations. When displayed, the **After-hours/Premarket** background instead indicates the extended price's difference from the regular price displayed in the table. It is neutral when values match **to the displayed cent**; this avoids false shading from sub-cent discrepancies.
- **Exchange and company names:** Exchange is separately sortable. Long exchange labels are shortened for display, and long company names stay on one line with an ellipsis; hover over a name to see it in full.
- **Configurable symbol links:** Click a symbol to open Google Finance, Stock Analysis, Corporate Website, or Investor Relations in a new tab. The preferred destination persists in browser storage. Financial research links are built using symbol/exchange metadata; official websites and IR pages are opened only when a URL is available.
- **Dynamic corporate website discovery:** With a `FINNHUB_API_KEY`, Go can resolve corporate URLs on demand through Finnhub company profiles and cache successful results for seven days. Verified built-in mappings remain a fallback. Corporate lookups do not run every quote refresh.
- **Compact mobile notes:** The formerly lengthy footer is now a native **About / Data Notes** disclosure, collapsed by default on both desktop and mobile, so the note text does not fill the iPhone screen. Tap to expand it.
- **Pacific-time header:** Quote and retrieval timestamps use PST/PDT as appropriate.
- **REST API:** Retrieve JSON quotes and the dynamic Most Active ranking.

The service uses Go's standard library, and the UI uses lightweight JavaScript without live charts or advertisements.

## Requirements

- Go installed on the machine hosting the service.
- Network access to the upstream quote provider.
- A current browser such as Safari.

The application uses only the Go standard library; there is no database or external Go dependency. Yahoo quote retrieval does not require an API key. A **Finnhub API key is optional** for dynamic corporate-website lookup.

## Build and run locally

From your repository directory (`~/mauristocks`):

```sh
cd ~/mauristocks
```

The project uses only the Go standard library. A `go.mod` file is recommended for reproducible builds and deployment. If one does not exist, initialize it **once**:

```sh
go mod init github.com/richard-mauri/mauristocks
```

Build and start the executable:

```sh
go build -o mauristocks .
./mauristocks
```

Alternatively, for a source-only run:

```sh
go run main.go
```

Open <http://localhost:8082>. Quotes load automatically on page open. There is no recurring polling unless you select an auto-refresh interval. Stop the server with **Control-C**.

### Command-line flags

```text
./mauristocks -h
./mauristocks --help
./mauristocks -version
./mauristocks -port 8090
./mauristocks -host 0.0.0.0 -port 8082
./mauristocks -debug
```

| Flag | Default | Description |
| --- | --- | --- |
| `-h`, `--help` | — | Print available flags and exit (provided by Go's standard `flag` package) |
| `-port` | `8082`, or `PORT` if set | TCP listening port |
| `-host` | `127.0.0.1` locally; `0.0.0.0` when `PORT` is set | Bind address |
| `-version` | `false` | Print semantic version and build ID, then exit |
| `-debug` | `false` | Log HTTP method, request path, and elapsed time |

Use `-port` to override the `PORT` environment variable:

```sh
PORT=9090 ./mauristocks
./mauristocks -port 8083
```

**Local safety:** The default `127.0.0.1` host binds only to your Mac. `-host 0.0.0.0` makes the application reachable from other network interfaces; use it only with appropriate access controls. When a hosting platform supplies `PORT`, the app automatically defaults to `0.0.0.0` so the platform can reach it. An explicit `-host` overrides that behavior.

### Optional Finnhub company-profile lookup

Corporate Website links can be resolved dynamically using a free Finnhub company-profile API key. Pass the key to the server through an environment variable (never put it in source code or browser storage):

```sh
export FINNHUB_API_KEY="YOUR_FINNHUB_API_KEY"
./mauristocks
```

When set, the server logs that Finnhub company profiles are enabled. When absent, built-in verified URL mappings remain usable. The key is only used server-side; company-profile lookups are separate from quote refreshes and successful lookups are cached for approximately seven days. Company-profile data may not cover every symbol or ETF. Investor Relations links remain curated mappings.

### Git housekeeping

Keep local `.env` and `.cloudflared/` token files out of version control, as well as the compiled binary.

Commit `go.mod` along with `main.go` and `README.md`. To avoid committing the compiled local executable, add the following to `.gitignore`:

```gitignore
/mauristocks
.env
.cloudflared/
.DS_Store
*.log
```

## How to use it

On opening the dashboard, the browser restores the last selected watchlist and immediately requests its prices. Click **Refresh quotes** at any time to request updated values. Use the **Auto refresh** menu to select *Never*, *5 seconds*, *30 seconds*, *1 minute*, or *5 minutes*. The default is Never unless a prior preference is stored.

### Named and dynamic watchlists

The **Watchlist** selector includes the read-only **Default (20 symbols)** list, individually saved named lists, and **Most Active — Top 20 (Dynamic)**. The old standalone **Restore Defaults** button has been removed: select **Default (20 symbols)** to return to the originals.

To save a personalized list, enter comma-separated symbols and click **Save As New**; give it a unique name. For a saved named list, use **Save Changes** to persist edits, or **Delete List** to remove it after confirmation. The Default and dynamic lists are protected; the most-active symbols are supplied by the upstream ranking and are not manually edited in place. Switching lists can prompt about unsaved changes.

On upgrade from earlier releases, legacy locally stored custom symbols can be migrated to **My Watchlist**. Named lists, the active selection, sort selection, and refresh interval are retained in the browser's `localStorage`, scoped to the browser and website origin. They do **not** synchronize across devices, and clearing site data removes them.

The dynamic Most Active list ranks by reported share count traded, **not** dollar volume. Its membership can change as new rankings are retrieved. Yahoo access or rate-limiting can prevent that ranking from loading.

### Symbol links and company details

As of Build v32, the link menu intentionally excludes **Yahoo Finance** and **Finviz** because of their ad-heavy public websites. This does not change the Yahoo-based quote data feed.

Use the **Symbol links** selector to choose Google Finance, Stock Analysis, Corporate Website, or Investor Relations. A click on a symbol opens a new browser tab; the dashboard stays open. The selection is stored per browser and origin. Previously saved Yahoo Finance or Finviz destinations automatically fall back to Google Finance. Yahoo is still used for **background quote retrieval**, not as a symbol-link destination. The Exchange column displays the provider's exchange information, with shortened labels where needed; it is also sortable.

When Corporate Website is selected, the app can send the click to its `/company` resolver, which requests the company profile from Finnhub if the server is configured with `FINNHUB_API_KEY`. Successful URLs are cached, and known corporate website mappings are used as fallback. If a URL cannot be established, the link is unavailable rather than guessed. Investor Relations links use the verified mappings built into this release.

### Display and appearance

The app uses a compact light theme with black regular text, alternating pale rows, vivid red and green changes, and fixed column widths. Sorting can move rows after a refresh.

| Column | Meaning |
| --- | --- |
| Price | Latest available regular-session price or regular close, independently of the extended-hours column |
| After-hours / Premarket | Available extended-session price derived from the provider's intraday bars; `—` when unavailable |
| Change / Change % | Session-aware difference: extended-session price versus the regular price during premarket/after-hours; regular price versus previous close during normal trading |
| Day Range | Regular-session low and high (previous available session before the open) |
| 52-Week Range | Approximate year-long low and high based on available daily bars |

The **Price** background shows short-term direction and rate of movement between successive observations with distinct quote timestamps. No false movement is inferred when a refreshed response has the same timestamp.

The **After-hours** (or **Premarket**) background measures its price against the regular price displayed in the Price column, **not** the difference from the previous refresh. Red indicates below, green indicates above, with bounded shading intensity. If the two displayed prices are equal after rounding to cents, the extended price cell remains unshaded. An unavailable extended price is shown as `—`, not silently replaced by a regular price.

The range markers show the displayed price's relative position in the recorded range; markers are clamped to the scale ends when necessary. Day and year ranges are not guaranteed to incorporate extended-session activity.

### Session-dependent extended-hours column (Build v34)

During normal **regular trading**, the application hides the extended-hours header, every extended-hours price cell, and the corresponding table column. During **premarket** it displays **Premarket**; during **after-hours** it displays **After-hours**. Table width shrinks by 110 pixels during a regular session, while the other columns retain their widths. The column's visibility is driven by the *session values returned by the Go server*, not by the browser clock, and updates on the next successful quotes retrieval (manual or scheduled). There is no need to reload the page. The source still includes extended-hours fields in JSON even when the browser hides the column.

If the active sort was extended price when the column becomes hidden, the dashboard switches to **Price**, descending; other sorts are preserved. When no data has been fetched or session results are mixed/unavailable, the hide condition may not apply. On weekends, holidays, or overnight, the column's presence and label follow available provider session classifications; the application does **not** implement an independent exchange-calendar-based visibility scheduler.

Build v33 introduced this behavior. **Build v34 fixes the extended-hours header not hiding along with the data cells** (an HTML `class` attribute issue in v33).

### Collapsed About / Data Notes

The long explanatory footer was replaced beginning with Build v29 by the **About / Data Notes** disclosure. It is **collapsed by default**, especially useful when opening the dashboard on an iPhone. Click or tap the heading to show details about momentum shading, price ranges, extended-hours behavior, data availability, and the JSON API. Tap again to close. This native HTML disclosure does not require extra JavaScript.

The fixed-width quote table can scroll horizontally on a narrow phone screen; during regular trading the table is 110 pixels narrower. A fully responsive mobile table is not yet implemented.

### Data freshness

Quote data is requested on page open, on manual Refresh, and optionally at a selected interval. **Retrieved** indicates when the service fetched data, not when the exchange traded. **Latest quote time** is based on the provider's reported price timestamps. Both display in Pacific time.

Provider data may be delayed, incomplete, or unavailable. A returned one-minute extended-hours bar close is **not** a guarantee of a live trade quote. A 5-second refresh setting with 20 symbols can generate substantial upstream traffic and may trigger rate limiting; use longer intervals where possible.

## HTTP endpoints

| Endpoint | Purpose |
| --- | --- |
| `GET /` | Browser dashboard |
| `GET /api/quotes` | JSON for the default watchlist |
| `GET /api/quotes?symbols=MU,NVDA,AVGO` | JSON for the specified symbols |
| `GET /api/most-active` | Dynamic list of most-active stock symbols by share volume |
| `GET /api/company?symbol=MU` | Resolved corporate website metadata, with source and cache details |
| `GET /company?symbol=MU` | Redirect to the resolved corporate website, if available |
| `GET /healthz` | Basic health check (`ok`) |

Examples:

```sh
curl 'http://localhost:8082/api/quotes?symbols=MU,NVDA,AVGO'
curl 'http://localhost:8082/api/most-active'
curl 'http://localhost:8082/api/company?symbol=MU'
```

The quotes JSON includes a `fetched_at` timestamp and a `quotes` array. A quote may contain symbol, name, currency, session, display price/time, regular and extended-session fields, changes, range values, data-source details, and a note or error. API timestamps use UTC; the browser converts displayed header times to Pacific time.

## Quote source and limitations

The service currently retrieves data from Yahoo Finance's **unofficial** chart endpoint (`query1.finance.yahoo.com/v8/finance/chart/`). Intraday requests include premarket and after-hours bars; daily historical data is used for 52-week ranges.

**Important:** This is an undocumented interface, **not an official or guaranteed free market-data API**. Its access policies, usage rights, rate limits, format, accuracy, availability, and coverage may change without notice. Check the provider's terms before deploying or sharing the service. Requests may fail or be blocked, particularly from cloud-hosting IP addresses. Do not use it as the sole basis for trading decisions.

The Go server uses a short quote cache (about **4 seconds**), a separate longer cache for historical range data (about **10 minutes**), and a roughly **15-minute** cache for the Most Active ranking. Caching reduces redundant upstream calls but does not make delayed data live. The application does not require or store Yahoo credentials. An optional `FINNHUB_API_KEY` enables company-profile lookup; do not publish it or commit it to Git.

### Testing from an iPhone on the same Wi-Fi

Start the server listening on your Mac's network interface:

```sh
./mauristocks -host 0.0.0.0 -port 8082
```

Find its local address with `ipconfig getifaddr en0`, then open `http://MAC_IP_ADDRESS:8082` in iPhone Safari while both devices use the same trusted Wi-Fi network. Your iPhone has its own browser-local saved watchlists; they do not automatically sync with Safari on your Mac. Ensure the macOS firewall permits the connection. Do not use open listening on untrusted networks without protection.

## Hosting and security

The same Go executable can run on a Mac, a server, or a Go-compatible hosting platform. The default local bind address is `127.0.0.1:8082`. For hosted deployments, set `PORT` (which also changes the default host to `0.0.0.0`), or provide explicit `-host` and `-port` flags.

A **Cloudflare Tunnel** can expose a locally running instance without opening a router port, but the Mac must remain awake and the Go process and tunnel must keep running. The tunnel does not host the app itself.

For a remotely accessible deployment, protect the service with access controls (for example, Cloudflare Access). The dashboard and JSON endpoint do not implement user authentication or request quotas. Avoid exposing unrestricted public access to the quote-fetching endpoint.

## GitHub setup

If `git init` was already run locally:

```sh
git add main.go go.mod README.md .gitignore
git commit -m "Build v34: session-dependent extended-hours column"
```

If the `origin` remote already points to your existing GitHub repository, push the current branch:

```sh
git push origin main
```

If you have not yet configured a remote, use the URL shown by your GitHub repository page. Choose public or private visibility intentionally, and do not commit secrets. Only the optional Finnhub API key needs configuration for corporate-profile discovery. Never commit API keys or local `.env` files.

## Disclaimer

For personal informational use. Prices may be stale, incomplete, delayed, or unavailable. This software does not provide investment advice or an execution-grade market-data feed.
