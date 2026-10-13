# Mauri's Stock Quotes

A lightweight, self-hosted stock watchlist written in Go. It displays regular-session, premarket, and after-hours prices in a compact browser table, without advertisements or a heavyweight charting interface.

**Current application version:** `1.0.0` · **Build:** `v24` (October 2026).

## Features

- **Eight-column quote table:** Symbol, Name, Price, After-hours/Premarket, Change, Change %, Day Range, and 52-Week Range. The **Price** column displays the regular-session price; a separate column displays an available extended-session price. The extended column switches its heading between **Premarket** and **After-hours** as appropriate.
- **Distinct price comparisons:** Change and Change % reflect the regular price against the preceding regular close. The extended price can differ from the regular price; its background shows whether it is above or below the relevant close.
- **Multiple named watchlists:** Create, select, modify, and delete locally saved lists using **Save As New**, **Save Changes**, and **Delete List**. The permanent, read-only **Default (20 symbols)** list contains `MU, AMAT, LRCX, SPCX, AVGO, TSM, TSLA, NVDA, JOBY, PLTR, LMT, AMGN, SNA, FTEC, JEPQ, VOO, IAU, FSTA, IBIT, O`.
- **Most Active — Top 20 (Dynamic):** A read-only watchlist ranked by reported daily **share volume** using Yahoo's unofficial market screener. The ranking is cached server-side for approximately 15 minutes. It depends on provider availability; a failed ranking request reports an error rather than inventing symbols.
- **Immediate initial loading:** The active watchlist is restored and quotes are requested automatically when the page opens, including when Auto Refresh is **Never**.
- **Manual or timed refresh:** Never, 5 seconds, 30 seconds, 1 minute, or 5 minutes. Automatic polling is skipped while the page is hidden; overlapping requests are prevented. Prices remain visible while fetching, and the **Refresh quotes** button keeps its label and dimensions.
- **Sortable table:** Click a column heading to toggle direction. Price, After-hours, and change columns use numeric comparisons; range columns sort by the displayed price's normalized position between their lows and highs. Sorting is retained on refresh.
- **Graphical ranges:** Compact day and 52-week low-to-high bars with diamond indicators. Both price labels are placed above their bars.
- **Independent price shading:** The regular **Price** background indicates short-term momentum between distinct timestamped observations. The **After-hours** background instead indicates the extended price's difference from the regular close (or previous close in premarket). It is neutral when values match **to the displayed cent**; this avoids false shading from sub-cent discrepancies.
- **Pacific-time header:** Quote and retrieval timestamps use PST/PDT as appropriate.
- **REST API:** Retrieve JSON quotes and the dynamic Most Active ranking.

The service uses Go's standard library, and the UI uses lightweight JavaScript without live charts or advertisements.

## Requirements

- Go installed on the machine hosting the service.
- Network access to the upstream quote provider.
- A current browser such as Safari.

The application uses only the Go standard library; there is no database, API key, or external Go dependency.

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

### Git housekeeping

Commit `go.mod` along with `main.go` and `README.md`. To avoid committing the compiled local executable, add the following to `.gitignore`:

```gitignore
/mauristocks
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

### Display and appearance

The app uses a compact light theme with black regular text, alternating pale rows, vivid red and green changes, and fixed column widths. Sorting can move rows after a refresh.

| Column | Meaning |
| --- | --- |
| Price | Latest available regular-session price or regular close, independently of the extended-hours column |
| After-hours / Premarket | Available extended-session price derived from the provider's intraday bars; `—` when unavailable |
| Change / Change % | Difference of regular-session price from the preceding regular-session close |
| Day Range | Regular-session low and high (previous available session before the open) |
| 52-Week Range | Approximate year-long low and high based on available daily bars |

The **Price** background shows short-term direction and rate of movement between successive observations with distinct quote timestamps. No false movement is inferred when a refreshed response has the same timestamp.

The **After-hours** (or **Premarket**) background measures its price against the applicable regular close, **not** the difference from the previous refresh. Red indicates below, green indicates above, with bounded shading intensity. If the two displayed prices are equal after rounding to cents, the extended price cell remains unshaded. An unavailable extended price is shown as `—`, not silently replaced by a regular price.

The range markers show the displayed price's relative position in the recorded range; markers are clamped to the scale ends when necessary. Day and year ranges are not guaranteed to incorporate extended-session activity.

### Data freshness

Quote data is requested on page open, on manual Refresh, and optionally at a selected interval. **Retrieved** indicates when the service fetched data, not when the exchange traded. **Latest quote time** is based on the provider's reported price timestamps. Both display in Pacific time.

Provider data may be delayed, incomplete, or unavailable. A returned one-minute extended-hours bar close is **not** a guarantee of a live trade quote. A 5-second refresh setting with 20 symbols can generate substantial upstream traffic and may trigger rate limiting; use longer intervals where possible.

## HTTP endpoints

| Endpoint | Purpose |
| --- | --- |
| `GET /` | Browser dashboard |
| `GET /api/quotes` | JSON for the default watchlist |
| `GET /api/quotes?symbols=MU,NVDA,AVGO` | JSON for the specified symbols |\n| `GET /api/most-active` | Dynamic list of most-active stock symbols by share volume |
| `GET /healthz` | Basic health check (`ok`) |

Examples:

```sh
curl 'http://localhost:8082/api/quotes?symbols=MU,NVDA,AVGO'
curl 'http://localhost:8082/api/most-active'
```

The quotes JSON includes a `fetched_at` timestamp and a `quotes` array. A quote may contain symbol, name, currency, session, display price/time, regular and extended-session fields, changes, range values, data-source details, and a note or error. API timestamps use UTC; the browser converts displayed header times to Pacific time.

## Quote source and limitations

The service currently retrieves data from Yahoo Finance's **unofficial** chart endpoint (`query1.finance.yahoo.com/v8/finance/chart/`). Intraday requests include premarket and after-hours bars; daily historical data is used for 52-week ranges.

**Important:** This is an undocumented interface, **not an official or guaranteed free market-data API**. Its access policies, usage rights, rate limits, format, accuracy, availability, and coverage may change without notice. Check the provider's terms before deploying or sharing the service. Requests may fail or be blocked, particularly from cloud-hosting IP addresses. Do not use it as the sole basis for trading decisions.

The Go server uses a short quote cache (about **4 seconds**), a separate longer cache for historical range data (about **10 minutes**), and a roughly **15-minute** cache for the Most Active ranking. Caching reduces redundant upstream calls but does not make delayed data live. The application does not require or store Yahoo credentials.

## Hosting and security

The same Go executable can run on a Mac, a server, or a Go-compatible hosting platform. The default local bind address is `127.0.0.1:8082`. For hosted deployments, set `PORT` (which also changes the default host to `0.0.0.0`), or provide explicit `-host` and `-port` flags.

A **Cloudflare Tunnel** can expose a locally running instance without opening a router port, but the Mac must remain awake and the Go process and tunnel must keep running. The tunnel does not host the app itself.

For a remotely accessible deployment, protect the service with access controls (for example, Cloudflare Access). The dashboard and JSON endpoint do not implement user authentication or request quotas. Avoid exposing unrestricted public access to the quote-fetching endpoint.

## GitHub setup

If `git init` was already run locally:

```sh
git add main.go go.mod README.md .gitignore
git commit -m "Build v24: extended-hours column, dynamic watchlist, and README"
```

If the `origin` remote already points to your existing GitHub repository, push the current branch:

```sh
git push origin main
```

If you have not yet configured a remote, use the URL shown by your GitHub repository page. Choose public or private visibility intentionally, and do not commit secrets. No credentials are required by the current code.

## Disclaimer

For personal informational use. Prices may be stale, incomplete, delayed, or unavailable. This software does not provide investment advice or an execution-grade market-data feed.
