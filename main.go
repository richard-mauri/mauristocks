package main

// Personal, on-demand quote dashboard. The Yahoo chart endpoint is unofficial:
// it can change, block requests, and may be subject to restrictions on automated use.
// Review provider terms before deploying. No quote is guaranteed real-time.

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var defaultSymbols = []string{"MU", "AMAT", "LRCX", "SPCX", "AVGO", "TSM", "TSLA", "NVDA", "JOBY", "PLTR", "LMT", "AMGN", "SNA", "FTEC", "JEPQ", "VOO", "IAU", "FSTA", "IBIT", "O"}
var symbolRE = regexp.MustCompile(`^[A-Z0-9][A-Z0-9.^=-]{0,17}$`)
var eastern, _ = time.LoadLocation("America/New_York")

// All times in API responses are RFC3339 UTC. Prices are USD unless currency says otherwise.
type Quote struct {
	Symbol          string     `json:"symbol"`
	Name            string     `json:"name,omitempty"`
	Currency        string     `json:"currency,omitempty"`
	RegularPrice    *float64   `json:"regular_price,omitempty"`
	PreviousClose   *float64   `json:"previous_close,omitempty"`
	DayLow          *float64   `json:"day_low,omitempty"`
	DayHigh         *float64   `json:"day_high,omitempty"`
	DayRangeDate    string     `json:"day_range_date,omitempty"`
	Week52Low       *float64   `json:"week_52_low,omitempty"`
	Week52High      *float64   `json:"week_52_high,omitempty"`
	ExtendedPrice   *float64   `json:"extended_price,omitempty"`
	ExtendedTime    *time.Time `json:"extended_time,omitempty"`
	ExtendedSession string     `json:"extended_session,omitempty"`
	DisplayPrice    *float64   `json:"display_price,omitempty"`
	DisplayTime     *time.Time `json:"display_time,omitempty"`
	Session         string     `json:"session"`
	Change          *float64   `json:"change,omitempty"`
	ChangePercent   *float64   `json:"change_percent,omitempty"`
	Source          string     `json:"source"`
	Note            string     `json:"note,omitempty"`
	Error           string     `json:"error,omitempty"`
}
type Response struct {
	FetchedAt time.Time `json:"fetched_at"`
	Quotes    []Quote   `json:"quotes"`
}
type period struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}
type yahooChart struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol               string   `json:"symbol"`
				ShortName            string   `json:"shortName"`
				LongName             string   `json:"longName"`
				Currency             string   `json:"currency"`
				RegularMarketPrice   *float64 `json:"regularMarketPrice"`
				PreviousClose        *float64 `json:"previousClose"`
				ChartPreviousClose   *float64 `json:"chartPreviousClose"`
				CurrentTradingPeriod struct {
					Pre     period `json:"pre"`
					Regular period `json:"regular"`
					Post    period `json:"post"`
				} `json:"currentTradingPeriod"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Close []*float64 `json:"close"`
					High  []*float64 `json:"high"`
					Low   []*float64 `json:"low"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error json.RawMessage `json:"error"`
	} `json:"chart"`
}

var client = &http.Client{Timeout: 12 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
	if len(via) > 3 {
		return errors.New("too many redirects")
	}
	return nil
}}

// Historical unadjusted daily bar highs/lows. One extra request per symbol,
// On refresh; historical data uses separate caching to avoid repeated annual requests.
type historicalRange struct {
	low, high, dayLow, dayHigh *float64
	dayDate                    string
	at                         time.Time
}

var historicalCache = struct {
	sync.Mutex
	m map[string]historicalRange
}{m: make(map[string]historicalRange)}

func fetchRanges(ctx context.Context, sym string, q *Quote, today string) {
	historicalCache.Lock()
	entry, found := historicalCache.m[sym]
	historicalCache.Unlock()
	if found && time.Since(entry.at) < 10*time.Minute {
		q.Week52Low, q.Week52High = entry.low, entry.high
		if q.DayLow == nil && entry.dayLow != nil {
			q.DayLow, q.DayHigh, q.DayRangeDate = entry.dayLow, entry.dayHigh, entry.dayDate
		}
		return
	}

	endpoint := "https://query1.finance.yahoo.com/v8/finance/chart/" + url.PathEscape(sym) + "?range=1y&interval=1d"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "MauriStockQuotes/1.0 (personal on-demand dashboard)")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		q.Note += " Historical ranges unavailable."
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		q.Note += " Historical ranges unavailable."
		return
	}
	var data yahooChart
	if json.NewDecoder(io.LimitReader(resp.Body, 3<<20)).Decode(&data) != nil || len(data.Chart.Result) == 0 {
		return
	}
	r := data.Chart.Result[0]
	if len(r.Indicators.Quote) == 0 {
		return
	}
	bars := r.Indicators.Quote[0]
	var lastDate string
	var lastLow, lastHigh *float64
	// Period daily bars can include the ongoing session. Only use the last
	// completed daily bar as a fallback when today's intraday range is absent.
	for i, stamp := range r.Timestamp {
		if i >= len(bars.Low) || i >= len(bars.High) || bars.Low[i] == nil || bars.High[i] == nil {
			continue
		}
		lo, hi := *bars.Low[i], *bars.High[i]
		if lo <= 0 || hi < lo || math.IsNaN(lo) || math.IsNaN(hi) {
			continue
		}
		day := time.Unix(stamp, 0).In(eastern).Format("2006-01-02")
		if day > today {
			continue
		}
		if q.Week52Low == nil || lo < *q.Week52Low {
			n := lo
			q.Week52Low = &n
		}
		if q.Week52High == nil || hi > *q.Week52High {
			n := hi
			q.Week52High = &n
		}
		// Skip incomplete current-day daily bar as fallback until market close.
		if day == today && q.Session != "closed" {
			continue
		}
		if day >= lastDate {
			lastDate = day
			a, b := lo, hi
			lastLow = &a
			lastHigh = &b
		}
	}
	if q.DayLow == nil && lastLow != nil {
		q.DayLow = lastLow
		q.DayHigh = lastHigh
		q.DayRangeDate = lastDate
	}
	historicalCache.Lock()
	historicalCache.m[sym] = historicalRange{q.Week52Low, q.Week52High, lastLow, lastHigh, lastDate, time.Now()}
	historicalCache.Unlock()
}

func fetchQuote(ctx context.Context, sym string) Quote {
	q := Quote{Symbol: sym, Source: "Yahoo Finance (unofficial chart endpoint)"}
	endpoint := "https://query1.finance.yahoo.com/v8/finance/chart/" + url.PathEscape(sym) + "?range=2d&interval=1m&includePrePost=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		q.Error = err.Error()
		return q
	}
	req.Header.Set("User-Agent", "MauriStockQuotes/1.0 (personal on-demand dashboard)")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		q.Error = "Quote service unavailable: " + err.Error()
		return q
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		q.Error = fmt.Sprintf("Provider returned HTTP %d", resp.StatusCode)
		return q
	}
	var data yahooChart
	if err = json.NewDecoder(io.LimitReader(resp.Body, 3<<20)).Decode(&data); err != nil {
		q.Error = "Invalid provider response: " + err.Error()
		return q
	}
	if len(data.Chart.Result) == 0 {
		q.Error = "Symbol not found or no quote data"
		return q
	}
	r := data.Chart.Result[0]
	q.Name = r.Meta.ShortName
	if q.Name == "" {
		q.Name = r.Meta.LongName
	}
	q.Currency = r.Meta.Currency
	q.RegularPrice = r.Meta.RegularMarketPrice
	q.PreviousClose = r.Meta.PreviousClose
	if q.PreviousClose == nil {
		q.PreviousClose = r.Meta.ChartPreviousClose
	}
	now := time.Now()
	nowET := now.In(eastern)
	today := nowET.Format("2006-01-02")
	periods := r.Meta.CurrentTradingPeriod
	// Yahoo's declared trading periods guard against weekends and holidays.
	// Never label a previous trading day's price as today's live quote.
	phase := "closed"
	unixNow := now.Unix()
	if within(unixNow, periods.Regular) {
		phase = "regular"
	}
	if within(unixNow, periods.Pre) {
		phase = "premarket"
	}
	if within(unixNow, periods.Post) {
		phase = "after-hours"
	}
	q.Session = phase
	type candidate struct {
		price   *float64
		ts      time.Time
		session string
	}
	var latestRegular, latestExtended *candidate
	var intradayLow, intradayHigh *float64
	if len(r.Indicators.Quote) > 0 {
		closes := r.Indicators.Quote[0].Close
		for i, unix := range r.Timestamp {
			if i >= len(closes) || closes[i] == nil || math.IsNaN(*closes[i]) || *closes[i] <= 0 {
				continue
			}
			t := time.Unix(unix, 0).UTC()
			if t.After(now.Add(2*time.Minute)) || t.In(eastern).Format("2006-01-02") != today {
				continue
			}
			if within(unix, periods.Regular) {
				bars := r.Indicators.Quote[0]
				if i < len(bars.Low) && i < len(bars.High) && bars.Low[i] != nil && bars.High[i] != nil {
					lo, hi := *bars.Low[i], *bars.High[i]
					if lo > 0 && hi >= lo && !math.IsNaN(lo) && !math.IsNaN(hi) {
						if intradayLow == nil || lo < *intradayLow {
							v := lo
							intradayLow = &v
						}
						if intradayHigh == nil || hi > *intradayHigh {
							v := hi
							intradayHigh = &v
						}
					}
				}
				if latestRegular == nil || t.After(latestRegular.ts) {
					latestRegular = &candidate{closes[i], t, "regular"}
				}
			} else if within(unix, periods.Pre) || within(unix, periods.Post) {
				session := "premarket"
				if within(unix, periods.Post) {
					session = "after-hours"
				}
				if latestExtended == nil || t.After(latestExtended.ts) {
					latestExtended = &candidate{closes[i], t, session}
				}
			}
		}
	}
	// Prefer the requested session's quote and never show premarket as
	// the primary price when the regular market is open.
	switch phase {
	case "regular":
		q.DisplayPrice = q.RegularPrice
		if latestRegular != nil {
			q.DisplayPrice = latestRegular.price
			q.DisplayTime = &latestRegular.ts
		}
		q.Note = "Regular-session price; provider data can be delayed."
		if q.PreviousClose != nil && q.DisplayPrice != nil && *q.PreviousClose > 0 {
			d := *q.DisplayPrice - *q.PreviousClose
			q.Change = &d
			pct := d / *q.PreviousClose * 100
			q.ChangePercent = &pct
		}
	case "premarket", "after-hours":
		if latestExtended != nil && latestExtended.session == phase {
			q.ExtendedPrice = latestExtended.price
			q.ExtendedTime = &latestExtended.ts
			q.ExtendedSession = latestExtended.session
			q.DisplayPrice = latestExtended.price
			q.DisplayTime = &latestExtended.ts
			base := q.RegularPrice
			if phase == "premarket" {
				base = q.PreviousClose
			}
			if base != nil && *base > 0 {
				d := *q.DisplayPrice - *base
				q.Change = &d
				pct := d / *base * 100
				q.ChangePercent = &pct
			}
			q.Note = "Extended-hours price uses the latest available 1-minute bar close; bar time is not a guaranteed trade time."
		} else {
			q.Note = "No current-session extended-hours bar available; no extended-hours price inferred."
		}
	default:
		q.DisplayPrice = q.RegularPrice
		q.Note = "Market closed; last regular-session price (may be from a prior trading day)."
	}

	if intradayLow != nil && intradayHigh != nil {
		q.DayLow = intradayLow
		q.DayHigh = intradayHigh
		q.DayRangeDate = today
	}
	fetchRanges(ctx, sym, &q, today)
	return q
}
func within(t int64, p period) bool {
	return p.Start > 0 && p.End > p.Start && t >= p.Start && t < p.End
}

func normalize(input string) ([]string, error) {
	parts := strings.Split(input, ",")
	if len(parts) > 50 {
		return nil, errors.New("maximum 50 symbols per request")
	}
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		s := strings.ToUpper(strings.TrimSpace(part))
		if s == "" {
			continue
		}
		if !symbolRE.MatchString(s) {
			return nil, fmt.Errorf("invalid symbol %q", s)
		}
		if !seen[s] {
			out = append(out, s)
			seen[s] = true
		}
	}
	if len(out) == 0 {
		return nil, errors.New("supply at least one symbol")
	}
	return out, nil
}

// Small in-memory TTL reduces accidental duplicate clicks without polling.
type cacheEntry struct {
	quote Quote
	at    time.Time
}

var cache = struct {
	sync.Mutex
	m map[string]cacheEntry
}{m: make(map[string]cacheEntry)}

func cachedQuote(ctx context.Context, s string) Quote {
	cache.Lock()
	e, ok := cache.m[s]
	cache.Unlock()
	if ok && time.Since(e.at) < 4*time.Second {
		return e.quote
	}
	q := fetchQuote(ctx, s)
	if q.Error == "" {
		cache.Lock()
		cache.m[s] = cacheEntry{q, time.Now()}
		cache.Unlock()
	}
	return q
}

func quotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	input := r.URL.Query().Get("symbols")
	if input == "" {
		input = strings.Join(defaultSymbols, ",")
	}
	symbols, err := normalize(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
	defer cancel()
	result := make([]Quote, len(symbols))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for j := 0; j < 4; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				result[i] = cachedQuote(ctx, symbols[i])
			}
		}()
	}
	for i := range symbols {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(Response{FetchedAt: time.Now().UTC(), Quotes: result})
}

const page = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Mauri's Stock Quotes</title>
<style>
:root{color-scheme:light dark;font-family:system-ui,-apple-system,sans-serif}body{max-width:1200px;margin:2rem auto;padding:0 1rem}h1{margin-bottom:.25rem}.muted{opacity:.68;font-size:.88rem}.controls{display:flex;gap:.6rem;flex-wrap:wrap;margin:1.4rem 0}input{flex:1;min-width:260px;padding:.7rem;border:1px solid #888;border-radius:6px;font:inherit}button{padding:.7rem 1.1rem;border:0;border-radius:6px;background:#2868bd;color:white;font:inherit;cursor:pointer}button:disabled{cursor:wait}#refresh{min-width:148px;white-space:nowrap}#refresh.fetching{color:#ffe08a}#refresh.fetching:disabled{opacity:1}table{width:1130px;min-width:1130px;table-layout:fixed;border-collapse:collapse;font-variant-numeric:tabular-nums}th,td{padding:.65rem .45rem;text-align:left;border-bottom:1px solid #8885}th{font-size:.85rem}th[data-sort]{cursor:pointer;user-select:none} .sort-button{border:0;background:transparent;color:inherit;padding:.25rem .1rem;font:inherit;font-weight:700;cursor:pointer;width:100%;text-align:left}th.num .sort-button{text-align:right}.sort-button:hover{text-decoration:underline}.sort-button:focus-visible{outline:2px solid #2868bd;outline-offset:2px}th[data-sort]:hover{text-decoration:underline}th[data-sort]:focus-visible{outline:2px solid #2868bd;outline-offset:-2px}th[data-sort][aria-sort="ascending"],th[data-sort][aria-sort="descending"]{text-decoration:underline}td.num,th.num{text-align:right}.up{color:#168243}.down{color:#d54848}.tablewrap{overflow-x:auto}.small{font-size:.8rem;opacity:.75}#message{margin:1rem 0}footer{margin:1.5rem 0;line-height:1.5}tr.error td{opacity:.7}a{color:inherit} @media(max-width:650px){body{margin:1rem auto}th,td{font-size:.8rem;padding:.55rem .25rem}}
 .controls{align-items:center}.refresh-label{align-self:center;font-size:.88rem;white-space:nowrap}.controls select{font:inherit;padding:.65rem .5rem;border-radius:6px;border:1px solid #888;background:Canvas;color:CanvasText;cursor:pointer}
.fetch-status{position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0}.range-cell{padding:.35rem .5rem!important}.range-labels{display:flex;justify-content:space-between;gap:.3rem;font-size:.68rem;line-height:1.25;font-variant-numeric:tabular-nums;white-space:nowrap;opacity:.85}.range-track{position:relative;height:5px;background:color-mix(in srgb,CanvasText 17%,transparent);border-radius:5px;margin:9px 6px 5px}.range-fill{position:absolute;inset:0;border-radius:inherit;background:color-mix(in srgb,CanvasText 35%,transparent)}.price-momentum{border-radius:4px;background-clip:padding-box}.range-marker{position:absolute;top:50%;width:10px;height:10px;background:#2f89d9;border:1px solid #eef5fc;transform:translate(-50%,-50%) rotate(45deg);box-shadow:0 0 0 1px #2868bd;z-index:1}

/* v15: high-contrast light palette, compact rows, clear range labels. */
:root{color-scheme:light;background:#f4f5f7;color:#111}
html{min-height:100%;background:#f4f5f7}
body{background:#f4f5f7;color:#111}
h1{color:#111}
.muted,.small,#message,footer{color:#111;opacity:1}
table{background:#fff;border-radius:5px}
th{background:#dde3e9;color:#111}
th,td{border-bottom:0}
th,td{padding:4px 7px;color:#111}
tbody tr{background:#fff}
tbody tr:nth-child(even){background:#edf1f5}
tbody tr:hover{background:#e0e8ef}
tbody td{height:34px;box-sizing:border-box}
input,.controls select{background:#fff;color:#111;border:1px solid #8995a2}
input:focus,.controls select:focus{outline:2px solid #2868bd;outline-offset:1px}
button{background:#2868bd;color:#fff}
#restore-defaults{background:#e4e9ef;color:#17212b;border:1px solid #a1adba;white-space:nowrap}
#restore-defaults:hover{background:#d4dde6}
.sort-button{background:transparent;color:#111}
.sort-button:hover{background:transparent}
.up{color:#087b38;font-weight:600}.down{color:#d21d35;font-weight:600}
a{color:#145ea8}
.range-cell{padding:3px 7px!important;vertical-align:middle}
.range-labels{display:flex;justify-content:space-between;gap:5px;line-height:13px;font-size:11px;font-weight:500;color:#111;opacity:1;white-space:nowrap}
.range-labels span{color:#111}
.range-track{height:4px;background:#aab5c1;border-radius:4px;margin:8px 6px 3px}
.range-fill{background:#aab5c1}
.range-marker{background:#2563a6;border-color:#fff;box-shadow:0 0 0 1px #2563a6;width:9px;height:9px}
@media(max-width:650px){th,td{padding:4px 5px}.range-cell{padding:3px 5px!important}}
/* v17 named watchlists */
.watchlist-controls{display:flex;align-items:center;gap:.55rem;flex-wrap:wrap;margin:-.55rem 0 1.1rem}
.watchlist-controls select{min-width:190px;padding:.55rem;border-radius:6px;border:1px solid #a1adba;background:white;color:#111}
.watchlist-controls button{padding:.55rem .8rem;background:#e4e9ef;color:#17212b;border:1px solid #a1adba}
.watchlist-controls button:disabled{opacity:.46;cursor:default}
#watchlist-status{font-size:.84rem;color:#303946}
</style></head><body><h1>Mauri's Stock Quotes <span class="small">v1.0.0 · Build v17</span></h1><p class="muted">Optional timed refresh • No ads or live charts • Unofficial quote source</p><div class="controls"><input id="symbols" aria-label="Comma-separated symbols" spellcheck="false"><button id="refresh">Refresh quotes</button><button id="restore-defaults" type="button" title="Restore original 20-symbol watchlist">Restore Defaults</button><span id="fetch-status" class="fetch-status" role="status" aria-live="polite"></span><label class="refresh-label" for="refresh-interval">Auto refresh</label><select id="refresh-interval" aria-label="Automatic refresh interval"><option value="0">Never</option><option value="5">5 seconds</option><option value="30">30 seconds</option><option value="60">1 minute</option><option value="300">5 minutes</option></select></div><div class="watchlist-controls"><label for="watchlist-select">Watchlist</label><select id="watchlist-select" aria-label="Saved watchlist"></select><button id="watchlist-new" type="button">Save As New</button><button id="watchlist-save" type="button">Save Changes</button><button id="watchlist-delete" type="button">Delete List</button><span id="watchlist-status" role="status" aria-live="polite"></span></div><p id="message" class="muted">Edit symbols, then Save As New or Save Changes to keep a named watchlist. Quotes refresh on demand.</p><div class="tablewrap"><table><colgroup><col style="width:90px"><col style="width:285px"><col style="width:115px"><col style="width:115px"><col style="width:120px"><col style="width:205px"><col style="width:200px"></colgroup><thead><tr><th data-sort="symbol" tabindex="0" role="button" aria-sort="none">Symbol</th><th class="optional" data-sort="name" tabindex="0" role="button" aria-sort="none">Name</th><th class="num" data-sort="display_price" tabindex="0" role="button" aria-sort="none">Price</th><th class="num" data-sort="change" tabindex="0" role="button" aria-sort="none">Change</th><th class="num" data-sort="change_percent" tabindex="0" role="button" aria-sort="none">Change %</th><th class="num" data-sort="day_position" tabindex="0" role="button" aria-sort="none">Day Range</th><th class="num" data-sort="week_52_position" tabindex="0" role="button" aria-sort="none">52-Week Range</th></tr></thead><tbody id="rows"></tbody></table></div><footer class="muted">Price cell background reflects the rate of change between successive timestamped quotes (green = rising; red = falling; no shading until a second distinct quote). Range bars show the low-to-high span, with a diamond marking the displayed price (clamped to the bar if outside the range). Day Range uses the latest regular-session high/low (previous session before market opens); 52-Week Range uses available daily bars over the past year. Regular hours show regular-session prices; extended hours show the latest available extended-session bar close, <strong>not guaranteed live trade quotes</strong>. Blank means no applicable session price was returned. Data may be delayed, incomplete, or unavailable. Quotes are for personal informational use. Endpoint: <code>/api/quotes?symbols=MU,NVDA</code>.</footer>
<script>
const defaults='MU,AMAT,LRCX,SPCX,AVGO,TSM,TSLA,NVDA,JOBY,PLTR,LMT,AMGN,SNA,FTEC,JEPQ,VOO,IAU,FSTA,IBIT,O';
const input=document.getElementById('symbols'),button=document.getElementById('refresh'),intervalSelect=document.getElementById('refresh-interval'),rows=document.getElementById('rows'),message=document.getElementById('message'),fetchStatus=document.getElementById('fetch-status');
const wlSelect=document.getElementById('watchlist-select');
const wlNew=document.getElementById('watchlist-new'),wlSave=document.getElementById('watchlist-save'),wlDelete=document.getElementById('watchlist-delete');
const wlStatus=document.getElementById('watchlist-status');
const wlKey='mauri_stock_watchlists_v1', activeKey='mauri_stock_active_watchlist';
function normalizeSymbols(value){return [...new Set(String(value).toUpperCase().split(',').map(x=>x.trim()).filter(Boolean))].join(',')}
function loadLists(){try{const value=JSON.parse(localStorage.getItem(wlKey));if(value&&typeof value==='object'&&!Array.isArray(value))return Object.fromEntries(Object.entries(value).filter(([k,v])=>k&&k!=='Default'&&typeof v==='string'));}catch(e){}return {}}
let watchlists=loadLists();
// Migrate the v16 custom watchlist, preserving its symbols on first launch.
if(localStorage.getItem(wlKey)===null){const legacy=localStorage.getItem('mauri_stock_symbols');if(legacy&&normalizeSymbols(legacy)!==defaults){watchlists['My Watchlist']=normalizeSymbols(legacy)}localStorage.setItem(wlKey,JSON.stringify(watchlists))}
let activeWatchlist=localStorage.getItem(activeKey)|| (watchlists['My Watchlist']?'My Watchlist':'Default');
if(activeWatchlist!=='Default'&&!Object.hasOwn(watchlists,activeWatchlist))activeWatchlist='Default';
function persistLists(){localStorage.setItem(wlKey,JSON.stringify(watchlists))}
function updateWatchlistButtons(){const isDefault=activeWatchlist==='Default';wlSave.disabled=isDefault;wlDelete.disabled=isDefault}
function updateWatchlistMenu(){wlSelect.replaceChildren();for(const n of ['Default',...Object.keys(watchlists).sort((a,b)=>a.localeCompare(b))]){const op=document.createElement('option');op.value=n;op.textContent=n==='Default'?'Default (20 symbols)':n;wlSelect.appendChild(op)}wlSelect.value=activeWatchlist;updateWatchlistButtons()}
function selectWatchlist(name,fetchNow=true){if(name!=='Default'&&!Object.hasOwn(watchlists,name))return;activeWatchlist=name;input.value=name==='Default'?defaults:watchlists[name];localStorage.setItem(activeKey,name);localStorage.setItem('mauri_stock_symbols',input.value);updateWatchlistMenu();wlStatus.textContent='';if(fetchNow)refreshQuotes()}
updateWatchlistMenu();selectWatchlist(activeWatchlist,false);
wlSelect.addEventListener('change',()=>{if(button.disabled){wlSelect.value=activeWatchlist;return}if(input.value!== (activeWatchlist==='Default'?defaults:watchlists[activeWatchlist])&&!window.confirm('Discard unsaved symbol edits and open another watchlist?')){wlSelect.value=activeWatchlist;return}selectWatchlist(wlSelect.value)});
wlNew.addEventListener('click',()=>{const name=window.prompt('Name the new watchlist:');if(name===null)return;const n=name.trim();if(!n||n.length>60||n.toLowerCase()==='default'){wlStatus.textContent='Choose a valid name (up to 60 characters).';return}if(Object.keys(watchlists).some(k=>k.toLowerCase()===n.toLowerCase())){wlStatus.textContent='That watchlist already exists.';return}const symbols=normalizeSymbols(input.value);if(!symbols){wlStatus.textContent='Enter symbols before saving.';return}watchlists[n]=symbols;persistLists();selectWatchlist(n,false);wlStatus.textContent='Saved '+n+'.'});
wlSave.addEventListener('click',()=>{if(activeWatchlist==='Default')return;const symbols=normalizeSymbols(input.value);if(!symbols){wlStatus.textContent='Enter symbols before saving.';return}watchlists[activeWatchlist]=symbols;input.value=symbols;persistLists();localStorage.setItem('mauri_stock_symbols',symbols);wlStatus.textContent='Saved '+activeWatchlist+'.'});
wlDelete.addEventListener('click',()=>{if(activeWatchlist==='Default'||button.disabled)return;if(!window.confirm('Delete watchlist "'+activeWatchlist+'"?'))return;delete watchlists[activeWatchlist];persistLists();selectWatchlist('Default');wlStatus.textContent='Deleted watchlist.'});
function money(n,currency){return n==null?'—':new Intl.NumberFormat('en-US',{style:'currency',currency:currency||'USD',maximumFractionDigits:2}).format(n)}
function cell(tr,value,cls){let td=document.createElement('td');td.textContent=value; if(cls)td.className=cls;tr.appendChild(td);return td}
function rangeCell(tr,low,high,price,currency){
 const td=document.createElement('td');td.className='range-cell';tr.appendChild(td);
 if(!Number.isFinite(low)||!Number.isFinite(high)||high<low){td.textContent='—';return td}
 const labels=document.createElement('div');labels.className='range-labels';
 const lo=document.createElement('span');lo.textContent=money(low,currency);
 const hi=document.createElement('span');hi.textContent=money(high,currency);
 labels.append(lo,hi);
 const track=document.createElement('div');track.className='range-track';
 const fill=document.createElement('div');fill.className='range-fill';track.appendChild(fill);
 const within=Number.isFinite(price);
 if(within){
  const pos=high===low?50:Math.max(0,Math.min(100,100*(price-low)/(high-low)));
  const marker=document.createElement('span');marker.className='range-marker';marker.style.left=pos+'%';
  marker.title='Displayed price: '+money(price,currency)+(price<low?' (below range)':price>high?' (above range)':'');
  marker.setAttribute('aria-label',marker.title);
  track.appendChild(marker);
 }
 td.title='Low: '+money(low,currency)+' • High: '+money(high,currency)+(within?' • Price: '+money(price,currency):'');
 td.append(labels,track);return td;
}
function barTime(v){if(!v)return '—';return new Date(v).toLocaleString('en-US',{timeZone:'America/Los_Angeles',month:'short',day:'numeric',hour:'numeric',minute:'2-digit',timeZoneName:'short'})}
let quotes=[];
// Browser-only state: compare successive, distinctly timestamped quotes for each symbol.
// Sorting and repeated responses with the same quote timestamp never alter momentum.
const lastObserved=new Map();
const priceMomentum=new Map();
function updateMomentum(nextQuotes){
 for(const q of nextQuotes){
  if(!Number.isFinite(q.display_price)||!q.display_time)continue;
  const ts=Date.parse(q.display_time);
  if(!Number.isFinite(ts))continue;
  const previous=lastObserved.get(q.symbol);
  if(previous && ts>previous.ts && previous.session===q.session && previous.price>0){
   const pct=100*(q.display_price-previous.price)/previous.price;
   const minutes=(ts-previous.ts)/60000;
   const rate=pct/minutes;
   // 0.05%/minute produces the strongest shade for the charcoal theme.
   const strength=Math.min(1,Math.abs(rate)/0.05);
   priceMomentum.set(q.symbol,{rate,alpha:0.12+0.23*strength});
  }else if(!previous || previous.session!==q.session || ts<previous.ts){
   priceMomentum.delete(q.symbol);
  }
  if(!previous || ts>previous.ts || previous.session!==q.session || ts<previous.ts){
   lastObserved.set(q.symbol,{price:q.display_price,ts,session:q.session});
  }
 }
}
const sortHeaders=[...document.querySelectorAll('th[data-sort]')];
for(const h of sortHeaders){
 const label=h.textContent;
 const b=document.createElement('button');b.type='button';b.className='sort-button';
 b.setAttribute('aria-label','Sort by '+label);b.textContent=label+' ⇅';
 h.replaceChildren(b);
 h.dataset.label=label;
}
const savedSort=(()=>{try{return JSON.parse(localStorage.getItem('mauri_stock_sort'))}catch{return null}})();
let sortKey=sortHeaders.some(h=>h.dataset.sort===savedSort?.key)?savedSort.key:'symbol';
let sortDirection=savedSort?.direction==='desc'?'desc':'asc';
const numericKeys=new Set(['display_price','change','change_percent','day_position','week_52_position']);
// A normalized price position: 0 = range low; 100 = range high.
// Values outside the range are clamped to the nearest endpoint, matching the marker.
function rangePosition(q,lowKey,highKey){
 const low=q[lowKey],high=q[highKey],price=q.display_price;
 if(!Number.isFinite(low)||!Number.isFinite(high)||!Number.isFinite(price)||high<low)return null;
 if(high===low)return 50;
 return Math.max(0,Math.min(100,100*(price-low)/(high-low)));
}
function sortValue(q,key){
 if(key==='day_position')return rangePosition(q,'day_low','day_high');
 if(key==='week_52_position')return rangePosition(q,'week_52_low','week_52_high');
 return q[key];
}
function compareQuotes(a,b){
 const av=sortValue(a,sortKey),bv=sortValue(b,sortKey);
 if(av==null&&bv==null)return a.symbol.localeCompare(b.symbol);
 if(av==null)return 1; // missing prices always last
 if(bv==null)return -1;
 let result=numericKeys.has(sortKey)?av-bv:String(av).localeCompare(String(bv),undefined,{sensitivity:'base'});
 return (sortDirection==='asc'?result:-result)||a.symbol.localeCompare(b.symbol);
}
function renderQuotes(){
 rows.replaceChildren();
 for(const q of [...quotes].sort(compareQuotes)){
  let tr=document.createElement('tr');if(q.error)tr.className='error';
  cell(tr,q.symbol);cell(tr,q.name||'','optional');
  const priceCell=cell(tr,money(q.display_price,q.currency),'num');
  const momentum=priceMomentum.get(q.symbol);
  if(momentum && Number.isFinite(q.display_price) && Math.abs(momentum.rate)>1e-10){
   const rising=momentum.rate>0;
   priceCell.classList.add('price-momentum');
   priceCell.style.backgroundColor=(rising?'rgba(29, 142, 79, ':'rgba(194, 61, 64, ')+momentum.alpha.toFixed(3)+')';
   priceCell.title=(rising?'Rising ':'Falling ')+Math.abs(momentum.rate).toFixed(3)+'%/min since preceding distinct quote';
  }
  let delta=q.change==null?'—':(q.change>=0?'+':'')+q.change.toFixed(2);
  cell(tr,delta,'num '+(q.change>0?'up':q.change<0?'down':''));
  let change=q.change_percent==null?'—':(q.change_percent>=0?'+':'')+q.change_percent.toFixed(2)+'%';
  cell(tr,change,'num '+(q.change_percent>0?'up':q.change_percent<0?'down':''));
  rangeCell(tr,q.day_low,q.day_high,q.display_price,q.currency);
  rangeCell(tr,q.week_52_low,q.week_52_high,q.display_price,q.currency);
  if(q.day_range_date && q.day_range_date!==new Date().toLocaleDateString('en-CA',{timeZone:'America/New_York',year:'numeric',month:'2-digit',day:'2-digit'}))tr.title='Day range from '+q.day_range_date;
  if(q.error)tr.title=q.error;
  rows.appendChild(tr);
 }
 for(const h of sortHeaders){
  const active=h.dataset.sort===sortKey;
  h.setAttribute('aria-sort',active?(sortDirection==='asc'?'ascending':'descending'):'none');
  h.querySelector('button').textContent=h.dataset.label+(active?(sortDirection==='asc'?' ↑':' ↓'):' ⇅');
  if(h.dataset.sort==='day_position'||h.dataset.sort==='week_52_position')h.querySelector('button').title='Ascending: price nearest range low first; descending: nearest range high first';
 }
}
for(const h of sortHeaders){
 const activate=()=>{
  if(sortKey===h.dataset.sort)sortDirection=sortDirection==='asc'?'desc':'asc';
  else{sortKey=h.dataset.sort;sortDirection=numericKeys.has(sortKey)?'desc':'asc'}
  localStorage.setItem('mauri_stock_sort',JSON.stringify({key:sortKey,direction:sortDirection}));
  renderQuotes();
 };
 h.querySelector('button').addEventListener('click',activate);
}
renderQuotes();
async function refreshQuotes(){
 if(button.disabled)return;
 let symbols=[...new Set(input.value.toUpperCase().split(',').map(x=>x.trim()).filter(Boolean))];
 if(!symbols.length){message.textContent='Enter one or more symbols.';return}
 input.value=symbols.join(',');localStorage.setItem('mauri_stock_symbols',input.value);
 button.disabled=true;button.classList.add('fetching');button.setAttribute('aria-busy','true');fetchStatus.textContent='Fetching quotes';
 try{
  let response=await fetch('/api/quotes?symbols='+encodeURIComponent(input.value),{cache:'no-store'});
  if(!response.ok)throw new Error(await response.text());
  let data=await response.json();quotes=data.quotes||[];updateMomentum(quotes);renderQuotes();
  const valid=quotes.filter(q=>q.display_price!=null);
  const sessions=[...new Set(valid.map(q=>q.session).filter(Boolean))];
  const session=sessions.length===1?sessions[0]:sessions.length?'Mixed ('+sessions.join(', ')+')':'Unavailable';
  const times=valid.map(q=>q.display_time).filter(Boolean).map(t=>new Date(t).getTime()).filter(Number.isFinite);
  const latest=times.length?new Date(Math.max(...times)).toISOString():null;
  message.textContent='Session: '+session+' • Latest quote time: '+barTime(latest)+' • Retrieved: '+new Date(data.fetched_at).toLocaleString('en-US',{timeZone:'America/Los_Angeles',month:'numeric',day:'numeric',year:'numeric',hour:'numeric',minute:'2-digit',second:'2-digit',timeZoneName:'short'})+' • '+valid.length+'/'+quotes.length+' prices. Refresh: '+(intervalSelect.value==='0'?'Never':intervalSelect.options[intervalSelect.selectedIndex].text)+'.';
 }catch(e){message.textContent='Refresh failed: '+e.message+' • Existing quotes retained.'}finally{button.disabled=false;button.classList.remove('fetching');button.removeAttribute('aria-busy');fetchStatus.textContent=''}
}
button.addEventListener('click',refreshQuotes);
// Restore only the stock symbols; leave sorting and refresh preferences intact.
const restoreDefaultsButton=document.getElementById('restore-defaults');
restoreDefaultsButton.addEventListener('click',()=>{
 if(button.disabled)return; // Avoid racing an in-progress quote request.
 if(!window.confirm('Restore the original 20-symbol watchlist? Your current symbols will be replaced.'))return;
 selectWatchlist('Default');
});

let refreshTimer=null;
function setRefreshInterval(){
 if(refreshTimer!==null){clearInterval(refreshTimer);refreshTimer=null}
 const seconds=Number(intervalSelect.value);
 localStorage.setItem('mauri_stock_refresh_interval',String(seconds));
 if(seconds>0)refreshTimer=setInterval(()=>{if(!document.hidden)refreshQuotes()},seconds*1000);
}
const savedInterval=localStorage.getItem('mauri_stock_refresh_interval');
if(['0','5','30','60','300'].includes(savedInterval))intervalSelect.value=savedInterval;
intervalSelect.addEventListener('change',setRefreshInterval);
setRefreshInterval();
</script></body></html>`

const appVersion = "1.0.0"
const buildVersion = "v17"

func main() {
	defaultPort := os.Getenv("PORT")
	if defaultPort == "" {
		defaultPort = "8082"
	}
	defaultHost := "127.0.0.1"
	if os.Getenv("PORT") != "" {
		defaultHost = "0.0.0.0"
	} // Cloud hosting must accept external traffic.
	port := flag.String("port", defaultPort, "HTTP listening port (overrides PORT)")
	host := flag.String("host", defaultHost, "Listening interface (127.0.0.1 for local-only, 0.0.0.0 for external access)")
	version := flag.Bool("version", false, "Print application version and exit")
	debug := flag.Bool("debug", false, "Log HTTP request method, path, and duration")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Mauri's Stock Quotes %s (Build %s)\n\nUsage: %s [options]\n\nOptions:\n", appVersion, buildVersion, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if *version {
		fmt.Printf("Mauri's Stock Quotes %s (Build %s)\n", appVersion, buildVersion)
		return
	}
	if _, err := net.LookupPort("tcp", *port); err != nil {
		log.Fatalf("Invalid port %q: %v", *port, err)
	}
	addr := net.JoinHostPort(*host, *port)
	tmpl := template.Must(template.New("index").Parse(page))
	mux := http.NewServeMux()
	mux.HandleFunc("/api/quotes", quotesHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = tmpl.Execute(w, nil)
	})
	var handler http.Handler = mux
	if *debug {
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			mux.ServeHTTP(w, r)
			log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started))
		})
	}
	log.Printf("Mauri's Stock Quotes %s (Build %s) listening on http://%s", appVersion, buildVersion, addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
