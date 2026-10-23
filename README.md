# Mauri's Stock Quotes (MauriStocks)

**Version 1.0.0 · Build v40 release candidate (PR #1; not yet merged)**

A small, self-hosted stock-watchlist dashboard written in Go. It emphasizes a clear, compact table of regular-session and available extended-hours prices without heavyweight charts or advertising.

**Source:** https://github.com/richard-mauri/mauristocks  
**Local URL (default):** http://localhost:8082/

> **Developers and future AI agents:** Read [AGENTS.md](AGENTS.md) for the authoritative development workflow, testing rules, Git pre-commit audit and current chat handoff. This README describes the application and how to use it.

## Features

MauriStocks displays stock symbols, exchanges, company names, regular-session prices, premarket/after-hours prices when available, changes, and graphical day/52-week price ranges. Click a column header to sort; a visual indicator identifies the active sort. The extended-hours column disappears during regular trading and returns during the applicable extended session when supported by provider data.

The **Price** column represents regular-session pricing, while **Premarket Price** or **After-hours Price** is shown separately. Extended-session Change $ and Change % compare the extended price against the adjacent regular price; regular-session changes compare the regular price with the previous close. Price coloring may indicate movement between distinct timestamped observations; extended-price coloring compares that displayed price with the regular price.

Use the permanent **Default** watchlist, create and edit named personal lists, or choose **Most Active — Top 20 (Dynamic)**, which is ranked by share volume when Yahoo's unofficial screener data is available. Personal watchlists, selected list, sort choices and refresh preferences are saved in the current browser. Flags are separate: clicking ☆ turns it into ★ and highlights the ticker's entire row. Browser-saved settings do not synchronize between devices.

You can refresh quotes manually or choose Never, 5 seconds, 30 seconds, 1 minute or 5 minutes. The page loads the selected list on opening. Automatic requests pause when the page is hidden. Very short intervals may trigger provider rate limits.

Click a ticker to open the selected research destination: Google Finance, Stock Analysis, a known company website or an investor-relations page when a verified destination is available. Company website discovery can optionally be enabled using Finnhub.

A **Columns** menu above the quote table lets you show or hide individual columns, reset column settings, and resize columns using the permanently visible grips on each column heading. Width and visibility preferences are saved in this browser. The Symbol column stays visible, and extended-session pricing still disappears automatically during regular market sessions. The Symbol links destination selector is next to Columns.

An **About / Data Notes** section expands for more explanation and is collapsed by default on phones and computers. The table may scroll horizontally on a narrow screen.

## Requirements and local use

You need Go, a current browser and upstream network access. The program uses the Go standard library and does not require a database or third-party Go module. From the documented local repository folder:

```sh
cd ~/projects/mauristocks
go test -count=1 ./...
go build -o mauristocks .
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
| `/api/company?symbol=MU` | Resolved company website metadata |
| `/company?symbol=MU` | Redirect to resolved company website when known |
| `/healthz` | Health response |

For example, `curl 'http://localhost:8082/api/quotes?symbols=MU,NVDA'` requests two quotes.

## Market-data notes

Quotes are retrieved through Yahoo Finance's **unofficial, undocumented** chart endpoint. The dynamic Most Active ranking uses Yahoo's unofficial screener. Their formats, rate limits, terms and availability can change, and returned data may be delayed or incomplete. An available intraday bar is not necessarily a real-time trade. The service uses short quote caching and longer range/most-active caches; caching does not make delayed data live.

The displayed retrieval time indicates when the service fetched information, not when the exchange last traded. Quote timestamps and retrieved times are displayed in Pacific time. Range indicators summarize provider data and may not include extended-session extremes.

This tool is for personal informational use, not trading execution or investment advice. For engineering changes, deployments, version history or cross-chat continuation, see [AGENTS.md](AGENTS.md).
