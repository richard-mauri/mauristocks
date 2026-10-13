# Mauri's Stock Quotes

A lightweight, self-hosted stock watchlist written in Go. It displays regular-session, premarket, and after-hours prices in a compact browser table, without advertisements or a heavyweight charting interface.

**Current application version:** `1.0.0` · **Build:** `v17` (October 2026).

## Features

- **Session-aware prices:** Shows a regular-session price during market hours and an available extended-session price during premarket or after-hours trading. Prices are labeled by session; extended-session values are based on available intraday bars and are **not guaranteed real-time trade quotes**.
- **Multiple named watchlists:** The built-in **Default** list contains 20 symbols: `MU, AMAT, LRCX, SPCX, AVGO, TSM, TSLA, NVDA, JOBY, PLTR, LMT, AMGN, SNA, FTEC, JEPQ, VOO, IAU, FSTA, IBIT, O`. Create additional named lists, save edits, and switch between them. Lists and the active selection persist in the browser's local storage; no accounts or database are needed.
- **Restore Defaults:** A confirmation-protected control switches back to the original built-in 20-symbol list without deleting named lists. The Default list cannot be overwritten or deleted.
- **Manual and optional automatic refresh:** Never (manual only), 5 seconds, 30 seconds, 1 minute, or 5 minutes. The selected interval is saved in the browser. The app skips automatic refresh when the page is hidden.
- **Sortable table:** Symbol, Name, Price, Change, Change %, Day Range, and 52-Week Range. The chosen sort remains active as new quotes arrive; rows can reorder after refresh.
- **Graphical ranges:** Low-to-high day and 52-week scales with a marker showing the displayed price's relative position. Range sorting uses that relative position, not absolute price.
- **Price movement shading:** A subtle green or red **background** behind the price indicates an upward or downward move between successive timestamped quotes. Stronger movement produces slightly stronger shading. The price text itself does not flash or change size; no shading appears until enough distinct observations exist. This short-term state lives in the browser, not in a database.
- **Header timestamps:** Latest quote time and retrieval time are shown in Pacific time (`PDT` or `PST`, as appropriate), rather than repeated in every row.
- **REST endpoint:** Read the same quote data as JSON.

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

Open <http://localhost:8082> and click **Refresh quotes**. There is no automatic quote polling unless you select an auto-refresh interval. Stop the server with **Control-C**.

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

Enter comma-separated symbols and click **Refresh quotes**. Click a column heading to toggle ascending and descending sorting. The selected sort order is applied again on each refresh, so rows may move as prices change.

### Saving and restoring named watchlists

1. Enter or edit the symbols in the comma-separated input field.
2. Click **Save As New** to name and save a new list (names must be unique, up to 60 characters).
3. Use the **Watchlist** dropdown to restore the built-in **Default (20 symbols)** or any of your saved named lists. Selecting a list fetches its quotes.
4. For an existing named list, edit its symbols and click **Save Changes** to update its saved contents.
5. Click **Delete List** to delete the active named list after confirmation. The built-in Default list is protected.
6. Click **Restore Defaults** to confirm switching back to the original 20-symbol list; this does not erase any separately saved named lists.

The application remembers the active named list and restores its saved symbols on subsequent visits to the same browser and site address. **Editing the symbol field alone does not update a named list:** use **Save Changes** or **Save As New** to retain it. Switching lists prompts before discarding unsaved edits.

**Upgrade from v16:** On first launch of v17, if the browser previously stored a customized v16 symbol list, the app migrates it into **My Watchlist**. If the old symbols match the built-in defaults, no extra list is created. Migration is local to each browser and site address.

Use **Auto refresh** to choose *Never*, *5 seconds*, *30 seconds*, *1 minute*, or *5 minutes*. A 5-second interval for a 20-stock watchlist can create substantial traffic to the upstream service and may trigger throttling. **Never** or **5 minutes** is the safer starting point.

When fetching, the **Refresh quotes** label remains fixed; its text color changes briefly. Existing quotes remain visible while a new request is in flight.

Named watchlists, last-selected watchlist, sort preferences, and refresh interval are saved **per browser and site origin** using local storage. They are not synced across devices or backed up on the server. Clearing browser site data removes them. Accessing the same service at `localhost:8082` versus a hosted hostname uses separate storage.

## Appearance

Build v17 uses a compact light theme with black text, alternating pale rows, red/green change figures, and price-cell background shading for short-term momentum. Range low/high labels sit above their graphical scales to prevent overlap.

## Price and range interpretation

| Display | Interpretation |
| --- | --- |
| Price | Session-appropriate reported price, if one is available |
| Change / Change % | Change relative to the applicable reference close |
| Day Range | Regular-session low–high; before market open, the preceding available regular session |
| 52-Week Range | Low–high calculated from available daily bars over roughly the prior year |
| Price shading | Recent direction and approximate speed of movement between distinct timestamped observations |

The diamond on a range bar indicates the price's **normalized position** between its low and high. When an extended-hours price falls outside a regular-session day range, the marker is clamped to the corresponding end of the bar. The range itself does not expand to include that extended-hours move.

Missing prices or ranges are shown as unavailable rather than invented. Timestamps matter: a fresh web request does **not** guarantee that the provider has supplied a new trade or price bar.

## HTTP endpoints

| Endpoint | Purpose |
| --- | --- |
| `GET /` | Browser dashboard |
| `GET /api/quotes` | JSON for the default watchlist |
| `GET /api/quotes?symbols=MU,NVDA,AVGO` | JSON for the specified symbols |
| `GET /healthz` | Basic health check (`ok`) |

Example:

```sh
curl 'http://localhost:8082/api/quotes?symbols=MU,NVDA,AVGO'
```

The JSON includes a `fetched_at` timestamp and a `quotes` array. A quote may contain symbol, name, currency, session, display price/time, regular and extended-session fields, changes, range values, data-source details, and a note or error. API timestamps use UTC; the browser converts displayed header times to Pacific time.

## Quote source and limitations

The service currently retrieves data from Yahoo Finance's **unofficial** chart endpoint (`query1.finance.yahoo.com/v8/finance/chart/`). Intraday requests include premarket and after-hours bars; daily historical data is used for 52-week ranges.

**Important:** This is an undocumented interface, **not an official or guaranteed free market-data API**. Its access policies, usage rights, rate limits, format, accuracy, availability, and coverage may change without notice. Check the provider's terms before deploying or sharing the service. Requests may fail or be blocked, particularly from cloud-hosting IP addresses. Do not use it as the sole basis for trading decisions.

The Go server uses a short quote cache (about **4 seconds**) and a separate longer cache for historical range data (about **10 minutes**). Caching reduces redundant upstream calls but does not make delayed data live. The application does not require or store Yahoo credentials.

## Hosting and security

The same Go executable can run on a Mac, a server, or a Go-compatible hosting platform. The default local bind address is `127.0.0.1:8082`. For hosted deployments, set `PORT` (which also changes the default host to `0.0.0.0`), or provide explicit `-host` and `-port` flags.

A **Cloudflare Tunnel** can expose a locally running instance without opening a router port, but the Mac must remain awake and the Go process and tunnel must keep running. The tunnel does not host the app itself.

For a remotely accessible deployment, protect the service with access controls (for example, Cloudflare Access). The dashboard and JSON endpoint do not implement user authentication or request quotas. Avoid exposing unrestricted public access to the quote-fetching endpoint.

## GitHub setup

If `git init` was already run locally:

```sh
git add main.go go.mod README.md .gitignore
git commit -m "Update stock quotes to Build v17 with named watchlists"
```

If the `origin` remote already points to your existing GitHub repository, push the current branch:

```sh
git push origin main
```

If you have not yet configured a remote, use the URL shown by your GitHub repository page. Choose public or private visibility intentionally, and do not commit secrets. No credentials are required by the current code.

## Disclaimer

For personal informational use. Prices may be stale, incomplete, delayed, or unavailable. This software does not provide investment advice or an execution-grade market-data feed.
