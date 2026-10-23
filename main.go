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
 "sort"
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
	Exchange        string     `json:"exchange,omitempty"`
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
	ExtendedChange *float64 `json:"extended_change,omitempty"`
	ExtendedChangePercent *float64 `json:"extended_change_percent,omitempty"`
	DisplayPrice    *float64   `json:"display_price,omitempty"`
	DisplayTime     *time.Time `json:"display_time,omitempty"`
	Session         string     `json:"session"`
	Change          *float64   `json:"change,omitempty"`
	ChangePercent   *float64   `json:"change_percent,omitempty"`
	Source          string     `json:"source"`
	Note            string     `json:"note,omitempty"`
	Error           string     `json:"error,omitempty"`
}

// The provider uses venue aliases; normalize only known values, never guess an exchange.
func normalizeExchange(full, code string) string {
 v := strings.ToUpper(strings.TrimSpace(full + " " + code))
 switch {
 case strings.Contains(v,"NASDAQ") || strings.Contains(v,"NMS") || strings.Contains(v,"NGM") || strings.Contains(v,"NCM"): return "NASDAQ"
 case strings.Contains(v,"NYSE ARCA") || strings.Contains(v,"NYQ ARCA") || strings.Contains(v,"PCX") || strings.Contains(v,"ARCX"): return "NYSEARCA"
 case strings.Contains(v,"NYSE AMERICAN") || strings.Contains(v,"ASE") || strings.Contains(v,"AMEX"): return "NYSEAMERICAN"
 case strings.Contains(v,"NEW YORK STOCK EXCHANGE") || strings.Contains(v,"NYQ") || strings.Contains(v,"NYSE"): return "NYSE"
 case strings.Contains(v,"BATS") || strings.Contains(v,"CBOE"): return "CBOE"
 default: return ""
 }
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
				ExchangeName         string   `json:"exchangeName"`
				FullExchangeName     string   `json:"fullExchangeName"`
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

// Yahoo's unofficial predefined screener may be blocked or changed without notice.
// Only daily share volume is used for the ranking, never dollar volume.
type mostActiveResponse struct {
 Finance struct {
  Result []struct { Quotes []struct {
   Symbol string `json:"symbol"`
   Volume *float64 `json:"regularMarketVolume"`
  } `json:"quotes"` } `json:"result"`
  Error json.RawMessage `json:"error"`
 } `json:"finance"`
}
var activeCache = struct {
 sync.Mutex
 symbols []string
 at time.Time
}{}

func topActiveSymbols(ctx context.Context) ([]string, time.Time, error) {
 activeCache.Lock()
 defer activeCache.Unlock()
 if len(activeCache.symbols)==20 && time.Since(activeCache.at)<15*time.Minute {
  return append([]string(nil), activeCache.symbols...),activeCache.at,nil
 }
 endpoint := "https://query1.finance.yahoo.com/v1/finance/screener/predefined/saved?count=100&scrIds=most_actives"
 req,err:=http.NewRequestWithContext(ctx,http.MethodGet,endpoint,nil)
 if err!=nil{return nil,time.Time{},err}
 req.Header.Set("User-Agent","Mozilla/5.0 (compatible; MauriStockQuotes/1.0)")
 req.Header.Set("Accept","application/json")
 resp,err:=client.Do(req)
 if err!=nil{return nil,time.Time{},fmt.Errorf("Yahoo Most Active unavailable: %w",err)}
 defer resp.Body.Close()
 if resp.StatusCode!=http.StatusOK {return nil,time.Time{},fmt.Errorf("Yahoo Most Active HTTP %d",resp.StatusCode)}
 var data mostActiveResponse
 if err:=json.NewDecoder(io.LimitReader(resp.Body,2<<20)).Decode(&data);err!=nil{return nil,time.Time{},fmt.Errorf("invalid screener response: %w",err)}
 if len(data.Finance.Result)==0 || len(data.Finance.Result[0].Quotes)==0 {return nil,time.Time{},errors.New("Yahoo returned no Most Active rankings")}
 type candidate struct{symbol string;volume float64}
 var candidates []candidate
 seen:=map[string]bool{}
 for _,q:=range data.Finance.Result[0].Quotes {
  sym:=strings.ToUpper(strings.TrimSpace(q.Symbol))
  if !symbolRE.MatchString(sym)||q.Volume==nil||*q.Volume<=0||seen[sym]{continue}
  seen[sym]=true
  candidates=append(candidates,candidate{sym,*q.Volume})
 }
 sort.SliceStable(candidates,func(i,j int)bool{return candidates[i].volume>candidates[j].volume})
 if len(candidates)<20{return nil,time.Time{},fmt.Errorf("screener returned only %d valid volume-ranked stocks; need 20",len(candidates))}
 symbols:=make([]string,20)
 for i:=range symbols{symbols[i]=candidates[i].symbol}
 activeCache.symbols=append([]string(nil),symbols...)
 activeCache.at=time.Now().UTC()
 return symbols,activeCache.at,nil
}
func mostActiveHandler(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet {http.Error(w,"method not allowed",http.StatusMethodNotAllowed);return}
 ctx,cancel:=context.WithTimeout(r.Context(),15*time.Second);defer cancel()
 symbols,updated,err:=topActiveSymbols(ctx)
 if err!=nil{http.Error(w,err.Error(),http.StatusBadGateway);return}
 w.Header().Set("Content-Type","application/json")
 w.Header().Set("Cache-Control","no-store")
 _=json.NewEncoder(w).Encode(map[string]any{"symbols":symbols,"ranked_at":updated,"ranking":"regular-session share volume","source":"Yahoo Finance unofficial screener"})
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
	q.Exchange = normalizeExchange(r.Meta.FullExchangeName, r.Meta.ExchangeName)
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
		} else if phase == "after-hours" {
			// Immediately after 4 PM ET, the provider may not have published
			// its first postmarket bar. Retain the last regular-session value
			// and clearly label it rather than blanking the entire watchlist.
			q.Session = "after-hours"
			q.DisplayPrice = q.RegularPrice
			if latestRegular != nil {
				q.DisplayPrice = latestRegular.price
				q.DisplayTime = &latestRegular.ts
			}
			if q.PreviousClose != nil && q.DisplayPrice != nil && *q.PreviousClose > 0 {
				d := *q.DisplayPrice - *q.PreviousClose
				q.Change = &d
				pct := d / *q.PreviousClose * 100
				q.ChangePercent = &pct
			}
			q.Note = "Regular-session price retained while waiting for the first after-hours price bar; not an after-hours quote."
		} else {
			q.Note = "No current-session premarket price bar available; no extended-hours price inferred."
		}
	default:
		q.DisplayPrice = q.RegularPrice
		q.Note = "Market closed; last regular-session price (may be from a prior trading day)."
	}

	// The Price column always represents the regular session; the separate
	// after-hours column uses ExtendedPrice only when an actual postmarket
	// bar was returned. Never substitute the closing price for extended data.
	if phase == "after-hours" || phase == "regular" || phase == "closed" {
		q.Session = phase
		if latestRegular != nil {
			q.RegularPrice = latestRegular.price
			q.DisplayTime = &latestRegular.ts
		}
		q.DisplayPrice = q.RegularPrice
		q.Change, q.ChangePercent = nil, nil
		if q.PreviousClose != nil && q.DisplayPrice != nil && *q.PreviousClose > 0 {
			d := *q.DisplayPrice - *q.PreviousClose
			pct := d / *q.PreviousClose * 100
			q.Change, q.ChangePercent = &d, &pct
		}
	}
	if phase == "premarket" {
		// Premarket is represented in the separate extended-hours column,
		// and the regular price remains the prior close.
		q.DisplayPrice = q.RegularPrice
		q.Change, q.ChangePercent = nil, nil
	}

	if intradayLow != nil && intradayHigh != nil {
		q.DayLow = intradayLow
		q.DayHigh = intradayHigh
		q.DayRangeDate = today
	}
	// Extended-session comparison is based on the same regular price that
	// appears in the Price column. Compare rounded cents to match the UI.
	// Keep a true zero change in JSON (pointers avoid omitempty dropping it).
	if q.ExtendedPrice != nil && q.DisplayPrice != nil && *q.DisplayPrice > 0 {
		baseCents := math.Round(*q.DisplayPrice * 100)
		extCents := math.Round(*q.ExtendedPrice * 100)
		if baseCents > 0 {
			delta := (extCents - baseCents) / 100
			pct := 100 * (extCents - baseCents) / baseCents
			q.ExtendedChange = &delta
			q.ExtendedChangePercent = &pct
		}
	}
	// Change and ChangePercent track the active session. During premarket
	// or after-hours they describe the extended quote versus the same regular
	// closing price displayed in Price. Missing extended data remains unknown.
	// At all other times they retain the regular-session change versus the
	// previous closing price. ExtendedChange fields remain independently
	// available to API clients for explicit comparisons.
	if phase == "premarket" || phase == "after-hours" {
		q.Change, q.ChangePercent = q.ExtendedChange, q.ExtendedChangePercent
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
:root{color-scheme:light dark;font-family:system-ui,-apple-system,sans-serif}body{max-width:1200px;margin:2rem auto;padding:0 1rem}h1{margin-bottom:.25rem}.muted{opacity:.68;font-size:.88rem}.controls{display:flex;gap:.6rem;flex-wrap:wrap;margin:1.4rem 0}input{flex:1;min-width:260px;padding:.7rem;border:1px solid #888;border-radius:6px;font:inherit}button{padding:.7rem 1.1rem;border:0;border-radius:6px;background:#2868bd;color:white;font:inherit;cursor:pointer}button:disabled{cursor:wait}#refresh{min-width:148px;white-space:nowrap}#refresh.fetching{color:#ffe08a}#refresh.fetching:disabled{opacity:1}table{width:1180px;min-width:1180px;table-layout:fixed;border-collapse:collapse;font-variant-numeric:tabular-nums}th,td{padding:.65rem .45rem;text-align:left;border-bottom:1px solid #8885}th{font-size:.85rem}th[data-sort]{cursor:pointer;user-select:none} .sort-button{border:0;background:transparent;color:inherit;padding:.25rem .1rem;font:inherit;font-weight:700;cursor:pointer;width:100%;text-align:left}th.num .sort-button{text-align:right}.sort-button:hover{text-decoration:underline}.sort-button:focus-visible{outline:2px solid #2868bd;outline-offset:2px}.sort-indicator{display:inline-block;margin-left:.2em;color:#8795a4;font-weight:500}.sort-indicator.active{color:#1763b1;font-size:1.15em;font-weight:900}th[data-sort][aria-sort="ascending"] .sort-indicator,th[data-sort][aria-sort="descending"] .sort-indicator{color:#1763b1}th[data-sort]:hover{text-decoration:underline}th[data-sort]:focus-visible{outline:2px solid #2868bd;outline-offset:-2px}th[data-sort][aria-sort="ascending"],th[data-sort][aria-sort="descending"]{text-decoration:underline}td.num,th.num{text-align:right}.up{color:#168243}.down{color:#d54848}.tablewrap{overflow-x:auto}.small{font-size:.8rem;opacity:.75}#message{margin:1rem 0}footer{margin:1.5rem 0;line-height:1.5}footer details summary{cursor:pointer;display:list-item;font-weight:600;width:max-content;max-width:100%;padding:.35rem 0}footer details .notes-content{padding:.45rem 0;max-width:1100px}tr.error td{opacity:.7}a{color:inherit} @media(max-width:650px){body{margin:1rem auto}th,td{font-size:.8rem;padding:.55rem .25rem}}
 .controls{align-items:center}.refresh-label{align-self:center;font-size:.88rem;white-space:nowrap}.controls select{font:inherit;padding:.65rem .5rem;border-radius:6px;border:1px solid #888;background:Canvas;color:CanvasText;cursor:pointer}
.fetch-status{position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0}.range-cell{padding:.35rem .5rem!important}.range-labels{display:flex;justify-content:space-between;gap:.3rem;font-size:.68rem;line-height:1.25;font-variant-numeric:tabular-nums;white-space:nowrap;opacity:.85}.range-track{position:relative;height:5px;background:color-mix(in srgb,CanvasText 17%,transparent);border-radius:5px;margin:9px 6px 5px}.range-fill{position:absolute;inset:0;border-radius:inherit;background:color-mix(in srgb,CanvasText 35%,transparent)}.price-momentum{font-weight:700}.range-marker{position:absolute;top:50%;width:10px;height:10px;background:#2f89d9;border:1px solid #eef5fc;transform:translate(-50%,-50%) rotate(45deg);box-shadow:0 0 0 1px #2868bd;z-index:1}

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
/* v25 link controls */
.research-control{display:flex;align-items:center;gap:.4rem}
.research-control select{font:inherit;border:1px solid #a1adba;border-radius:6px;background:white;color:#111;padding:.6rem .45rem}
.symbol-link{color:#145ea8;text-decoration:none;font-weight:650}
.symbol-link:hover,.symbol-link:focus-visible{text-decoration:underline}
.symbol-unavailable{color:#111;font-weight:650}
/* v28: keep exchange/name text inside fixed-width cells. */
tbody td.exchange-cell,tbody td.company-name{
  white-space:nowrap;overflow:hidden;text-overflow:ellipsis;
}
tbody td.exchange-cell{font-size:.88rem}
tbody td.company-name{max-width:0}
/* v34: remove the extended column entirely during regular trading.
   Fixed widths of the remaining columns do not change. */
table.regular-session{width:1070px;min-width:1070px}
table.regular-session col.extended-column{visibility:collapse}
table.regular-session th.extended-column,
table.regular-session td.extended-column{display:none}
/* v35: mark any symbol across watchlists and sorting, while allowing price momentum text colors to remain visible. */
tbody tr.flagged-row,tbody tr.flagged-row:nth-child(even),tbody tr.flagged-row:hover{background:#fff5d6}
.symbol-cell{white-space:nowrap}
.symbol-flag{background:transparent;border:0;color:#727985;cursor:pointer;font-size:16px;line-height:1;padding:2px 3px;margin-right:2px;vertical-align:middle;border-radius:3px}
.symbol-flag[aria-pressed="true"]{color:#946900}
.symbol-flag:hover,.symbol-flag:focus-visible{background:#ffe38b;outline:1px solid #aa8a35}



/* Shared dashboard layout: NewsGo fills its content grid; ScraperGo keeps
   page background continuous. Allocate *all* table width to visible columns.
   Percentage col tracks override the legacy pixel-width HTML attributes. */
.tablewrap{width:100%;max-width:100%;overflow-x:auto}
#quotes-table{width:100%;min-width:1180px;table-layout:fixed}
#quotes-table col:nth-child(1){width:6%!important}
#quotes-table col:nth-child(2){width:7%!important}
#quotes-table col:nth-child(3){width:25%!important}
#quotes-table col:nth-child(4){width:9%!important}
#quotes-table col:nth-child(5){width:9%!important}
#quotes-table col:nth-child(6){width:9%!important}
#quotes-table col:nth-child(7){width:9%!important}
#quotes-table col:nth-child(8){width:13%!important}
#quotes-table col:nth-child(9){width:13%!important}
/* In regular session the fifth column is removed: redistribute its space
   among the eight displayed columns, retaining a 1070px scroll threshold. */
#quotes-table.regular-session{width:100%;min-width:1070px}
#quotes-table.regular-session col:nth-child(1){width:7%!important}
#quotes-table.regular-session col:nth-child(2){width:8%!important}
#quotes-table.regular-session col:nth-child(3){width:29%!important}
#quotes-table.regular-session col:nth-child(4){width:10%!important}
#quotes-table.regular-session col:nth-child(5){width:0!important}
#quotes-table.regular-session col:nth-child(6){width:10%!important}
#quotes-table.regular-session col:nth-child(7){width:10%!important}
#quotes-table.regular-session col:nth-child(8){width:13%!important}
#quotes-table.regular-session col:nth-child(9){width:13%!important}

/* Shared semantic palette: ScraperGo / NewsGo. Preserves financial status colors. */
:root{
 color-scheme:light dark;
 --bg:light-dark(#f5f5f0,#0e141a);
 --panel:light-dark(#ffffff,#19232b);
 --ink:light-dark(#182631,#eef2f4);
 --muted:light-dark(#60707b,#a5b3bb);
 --border:light-dark(#dce2e1,#34434a);
 --accent:light-dark(#126877,#69c6d3);
 --soft:light-dark(#e6f3f3,#203c43);
 --row-alt:light-dark(#f2f7f6,#202d35);
 --row-hover:light-dark(#e5f0ef,#29414a);
 --starred:light-dark(#fff5d6,#4b3a20);
 --gain:light-dark(#087b38,#72d89b);
 --loss:light-dark(#b42334,#ff9696);
}
/* Override the old :root light canvas as well as the centered body's paint. */
:root{background:var(--bg);color:var(--ink)}
html,body{background:var(--bg);color:var(--ink)}
body{min-height:100vh}
h1{color:var(--ink)}
.muted,.small,#message,footer,#watchlist-status{color:var(--muted);opacity:1}
table{background:var(--panel);border:1px solid var(--border);border-radius:8px}
th{background:var(--soft);color:var(--ink);border-bottom:1px solid var(--border)}
th,td{color:var(--ink)}
tbody tr{background:var(--panel)}
tbody tr:nth-child(even){background:var(--row-alt)}
tbody tr:hover{background:var(--row-hover)}
input,.controls select,.watchlist-controls select,.research-control select{
 background:var(--panel);color:var(--ink);border:1px solid var(--border)
}
button{background:light-dark(#126877,#24454e);color:light-dark(#fff,#eef2f4);border:1px solid var(--border);border-radius:6px}
button:not(:disabled):hover{background:light-dark(#0d5360,#305c66)}
.watchlist-controls button{background:var(--soft);color:var(--ink);border:1px solid var(--border)}
.sort-button,.sort-button:hover,.symbol-flag{background:transparent;color:var(--ink);border:0}
.sort-indicator.active,th[data-sort][aria-sort="ascending"] .sort-indicator,th[data-sort][aria-sort="descending"] .sort-indicator{color:var(--accent)}
input:focus-visible,select:focus-visible,button:focus-visible,a:focus-visible,th[data-sort]:focus-visible{outline:2px solid var(--accent);outline-offset:2px}
a,.symbol-link{color:var(--accent)}
.up{color:var(--gain)}.down{color:var(--loss)}
.range-labels,.range-labels span{color:var(--ink)}
.range-track,.range-fill{background:var(--border)}
.range-marker{background:var(--accent);border-color:var(--panel);box-shadow:0 0 0 1px var(--accent)}
tbody tr.flagged-row,tbody tr.flagged-row:nth-child(even),tbody tr.flagged-row:hover{background:var(--starred)}
.symbol-flag[aria-pressed="true"]{color:light-dark(#946900,#ffd27d)}
.symbol-flag:hover,.symbol-flag:focus-visible{background:var(--soft);outline:1px solid var(--accent)}

</style></head><body><h1>Mauri's Stock Quotes <span class="small">v1.0.0 · Build v39</span></h1><p class="muted">Optional timed refresh • No ads or live charts • Unofficial quote source</p><div class="controls"><input id="symbols" aria-label="Comma-separated symbols" spellcheck="false"><button id="refresh">Refresh quotes</button><span id="fetch-status" class="fetch-status" role="status" aria-live="polite"></span><label class="refresh-label" for="refresh-interval">Auto refresh</label><select id="refresh-interval" aria-label="Automatic refresh interval"><option value="0">Never</option><option value="5">5 seconds</option><option value="30">30 seconds</option><option value="60">1 minute</option><option value="300">5 minutes</option></select><label class="research-control" for="research-site">Symbol links <select id="research-site" aria-label="Stock symbol link destination"><option value="google">Google Finance</option><option value="stockanalysis">Stock Analysis</option><option value="corporate">Corporate Website</option><option value="ir">Investor Relations</option></select></label></div><div class="watchlist-controls"><label for="watchlist-select">Watchlist</label><select id="watchlist-select" aria-label="Saved watchlist"></select><button id="watchlist-new" type="button">Save As New</button><button id="watchlist-save" type="button">Save Changes</button><button id="watchlist-delete" type="button">Delete List</button><span id="watchlist-status" role="status" aria-live="polite"></span></div><p id="message" class="muted">Edit symbols, then Save As New or Save Changes to keep a named watchlist. Quotes refresh on demand.</p><div class="tablewrap"><table id="quotes-table"><colgroup><col style="width:75px"><col style="width:80px"><col style="width:255px"><col style="width:100px"><col class="extended-column" style="width:110px"><col style="width:100px"><col style="width:100px"><col style="width:180px"><col style="width:180px"></colgroup><thead><tr><th data-sort="symbol" tabindex="0" role="button" aria-sort="none">Symbol</th><th data-sort="exchange" tabindex="0" role="button" aria-sort="none">Exchange</th><th class="optional" data-sort="name" tabindex="0" role="button" aria-sort="none">Name</th><th class="num" data-sort="display_price" tabindex="0" role="button" aria-sort="none">Price</th><th class="num extended-column" data-sort="extended_price" tabindex="0" role="button" aria-sort="none" id="extended-heading">After-hours Price</th><th class="num" data-sort="change" tabindex="0" role="button" aria-sort="none" id="change-heading">Change $</th><th class="num" data-sort="change_percent" tabindex="0" role="button" aria-sort="none" id="change-percent-heading">Change %</th><th class="num" data-sort="day_position" tabindex="0" role="button" aria-sort="none">Day Range</th><th class="num" data-sort="week_52_position" tabindex="0" role="button" aria-sort="none">52-Week Range</th></tr></thead><tbody id="rows"></tbody></table></div><footer class="muted"><details><summary>About / Data Notes</summary><div class="notes-content">Click the star beside a symbol to toggle persistent pale-yellow highlighting across watchlists and sorting; price momentum colors remain visible. Price text color reflects short-term momentum between distinct quotes. After-hours text color reflects the percentage difference from the regular close (premarket from the previous close). Green indicates above the close; red indicates below. Neutral text color when an extended quote is unavailable or matches the close to the displayed cent. Range bars show the low-to-high span, with a diamond marking the displayed price (clamped to the bar if outside the range). Day Range uses the latest regular-session high/low (previous session before market opens); 52-Week Range uses available daily bars over the past year. Price always shows the regular-session value; the separate After-hours Price (or Premarket Price) column shows an extended-hours 1-minute bar close only when available, <strong>not guaranteed live trade quotes</strong>. Change $ and Change % follow the active session: extended price versus the regular Price column before/after market hours, and regular price versus the previous close during normal hours. If extended prices are unavailable, the session changes are blank. Blank means no applicable session price was returned. Data may be delayed, incomplete, or unavailable. Quotes are for personal informational use. Endpoint: <code>/api/quotes?symbols=MU,NVDA</code>.</div></details></footer>
<script>
const defaults='MU,AMAT,LRCX,SPCX,AVGO,TSM,TSLA,NVDA,JOBY,PLTR,LMT,AMGN,SNA,FTEC,JEPQ,VOO,IAU,FSTA,IBIT,O';
const input=document.getElementById('symbols'),button=document.getElementById('refresh'),intervalSelect=document.getElementById('refresh-interval'),rows=document.getElementById('rows'),message=document.getElementById('message'),fetchStatus=document.getElementById('fetch-status');
const wlSelect=document.getElementById('watchlist-select');
const wlNew=document.getElementById('watchlist-new'),wlSave=document.getElementById('watchlist-save'),wlDelete=document.getElementById('watchlist-delete');
const wlStatus=document.getElementById('watchlist-status');
const dynamicName="Most Active — Top 20 (Dynamic)";
const wlKey='mauri_stock_watchlists_v1', activeKey='mauri_stock_active_watchlist';
const flagStorageKey='mauri_stock_flagged_symbols_v1';
function loadFlags(){try{const x=JSON.parse(localStorage.getItem(flagStorageKey));if(Array.isArray(x))return new Set(x.filter(s=>typeof s==='string'&&/^[A-Z0-9][A-Z0-9.^=-]{0,17}$/.test(s)));}catch(e){}return new Set()}
const flaggedSymbols=loadFlags();
function toggleFlag(symbol){if(flaggedSymbols.has(symbol))flaggedSymbols.delete(symbol);else flaggedSymbols.add(symbol);try{localStorage.setItem(flagStorageKey,JSON.stringify([...flaggedSymbols].sort()));}catch(e){}renderQuotes()}

function normalizeSymbols(value){return [...new Set(String(value).toUpperCase().split(',').map(x=>x.trim()).filter(Boolean))].join(',')}
function loadLists(){try{const value=JSON.parse(localStorage.getItem(wlKey));if(value&&typeof value==='object'&&!Array.isArray(value))return Object.fromEntries(Object.entries(value).filter(([k,v])=>k&&k!=='Default'&&typeof v==='string'));}catch(e){}return {}}
let watchlists=loadLists();
// Migrate the v16 custom watchlist, preserving its symbols on first launch.
if(localStorage.getItem(wlKey)===null){const legacy=localStorage.getItem('mauri_stock_symbols');if(legacy&&normalizeSymbols(legacy)!==defaults){watchlists['My Watchlist']=normalizeSymbols(legacy)}localStorage.setItem(wlKey,JSON.stringify(watchlists))}
let activeWatchlist=localStorage.getItem(activeKey)|| (watchlists['My Watchlist']?'My Watchlist':'Default');
if(activeWatchlist!=='Default'&&activeWatchlist!==dynamicName&&!Object.hasOwn(watchlists,activeWatchlist))activeWatchlist='Default';
function persistLists(){localStorage.setItem(wlKey,JSON.stringify(watchlists))}
function updateWatchlistButtons(){const readOnly=activeWatchlist==='Default'||activeWatchlist===dynamicName;wlSave.disabled=readOnly;wlDelete.disabled=readOnly;input.readOnly=activeWatchlist===dynamicName}
function updateWatchlistMenu(){wlSelect.replaceChildren();for(const n of ['Default',dynamicName,...Object.keys(watchlists).sort((a,b)=>a.localeCompare(b))]){const op=document.createElement('option');op.value=n;op.textContent=n==='Default'?'Default (20 symbols)':n;wlSelect.appendChild(op)}wlSelect.value=activeWatchlist;updateWatchlistButtons()}
function selectWatchlist(name,fetchNow=true){if(name!=='Default'&&name!==dynamicName&&!Object.hasOwn(watchlists,name))return;activeWatchlist=name;input.value=name==='Default'?defaults:name===dynamicName?'':watchlists[name];localStorage.setItem(activeKey,name);localStorage.setItem('mauri_stock_symbols',input.value);updateWatchlistMenu();wlStatus.textContent='';if(fetchNow)refreshQuotes()}
updateWatchlistMenu();selectWatchlist(activeWatchlist,false);
wlSelect.addEventListener('change',()=>{if(button.disabled){wlSelect.value=activeWatchlist;return}if(activeWatchlist!==dynamicName&&input.value!== (activeWatchlist==='Default'?defaults:watchlists[activeWatchlist])&&!window.confirm('Discard unsaved symbol edits and open another watchlist?')){wlSelect.value=activeWatchlist;return}selectWatchlist(wlSelect.value)});
wlNew.addEventListener('click',()=>{const name=window.prompt('Name the new watchlist:');if(name===null)return;const n=name.trim();if(!n||n.length>60||(n.toLowerCase()==='default'||n.toLowerCase()===dynamicName.toLowerCase())){wlStatus.textContent='Choose a valid name (up to 60 characters).';return}if(Object.keys(watchlists).some(k=>k.toLowerCase()===n.toLowerCase())){wlStatus.textContent='That watchlist already exists.';return}const symbols=normalizeSymbols(input.value);if(!symbols){wlStatus.textContent='Enter symbols before saving.';return}watchlists[n]=symbols;persistLists();selectWatchlist(n,false);wlStatus.textContent='Saved '+n+'.'});
wlSave.addEventListener('click',()=>{if(activeWatchlist==='Default'||activeWatchlist===dynamicName)return;const symbols=normalizeSymbols(input.value);if(!symbols){wlStatus.textContent='Enter symbols before saving.';return}watchlists[activeWatchlist]=symbols;input.value=symbols;persistLists();localStorage.setItem('mauri_stock_symbols',symbols);wlStatus.textContent='Saved '+activeWatchlist+'.'});
wlDelete.addEventListener('click',()=>{if(activeWatchlist==='Default'||activeWatchlist===dynamicName||button.disabled)return;if(!window.confirm('Delete watchlist "'+activeWatchlist+'"?'))return;delete watchlists[activeWatchlist];persistLists();selectWatchlist('Default');wlStatus.textContent='Deleted watchlist.'});
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
const pacificTimeZone='America/Los_Angeles';
function barTime(v){if(!v)return '—';return new Date(v).toLocaleString('en-US',{timeZone:pacificTimeZone,month:'short',day:'numeric',hour:'numeric',minute:'2-digit',timeZoneName:'short'})}
function retrievedTime(fetchedAt,latestQuote){
 const fetched=new Date(fetchedAt);
 if(!Number.isFinite(fetched.getTime()))return '—';
 const quote=latestQuote?new Date(latestQuote):null;
 const dateKey=d=>d.toLocaleDateString('en-CA',{timeZone:pacificTimeZone,year:'numeric',month:'2-digit',day:'2-digit'});
 const sameDate=quote&&Number.isFinite(quote.getTime())&&dateKey(quote)===dateKey(fetched);
 const options={timeZone:pacificTimeZone,hour:'numeric',minute:'2-digit',second:'2-digit',timeZoneName:'short'};
 if(!sameDate){options.month='short';options.day='numeric';}
 return fetched.toLocaleString('en-US',options);
}
let quotes=[];
// Browser-only state: compare successive, distinctly timestamped quotes for each symbol.
// Sorting and repeated responses with the same quote timestamp never alter momentum.
const lastObserved=new Map();
const priceMomentum=new Map();
function updateChannelMomentum(nextQuotes,priceField,timeField,sessionField,observed,momentum){
 for(const q of nextQuotes){
  const price=q[priceField], time=q[timeField], session=q[sessionField];
  if(!Number.isFinite(price)||!time){
   // No new usable observation. For the extended column, don't show a
   // trend left over from a previous session or missing price.
   if(priceField==='extended_price')momentum.delete(q.symbol);
   continue;
  }
  const ts=Date.parse(time);
  if(!Number.isFinite(ts))continue;
  const previous=observed.get(q.symbol);
  if(previous && ts>previous.ts && previous.session===session && previous.price>0){
   const rate=(100*(price-previous.price)/previous.price)/((ts-previous.ts)/60000);
   const strength=Math.min(1,Math.abs(rate)/0.05);
   momentum.set(q.symbol,{rate,alpha:0.12+0.23*strength});
  }else if(!previous || previous.session!==session || ts<previous.ts){
   momentum.delete(q.symbol);
  }
  if(!previous || ts>previous.ts || previous.session!==session || ts<previous.ts){
   observed.set(q.symbol,{price,ts,session});
  }
 }
}
function updateMomentum(nextQuotes){
 updateChannelMomentum(nextQuotes,'display_price','display_time','session',lastObserved,priceMomentum);
}
function shadeMomentum(cell,momentum,price){
 if(!momentum || !Number.isFinite(price) || Math.abs(momentum.rate)<=1e-10)return;
 const rising=momentum.rate>0;
 cell.classList.add('price-momentum');
 cell.style.color=rising?'#137a3b':'#bf3037';
 cell.title+=(cell.title?' • ':'')+(rising?'Rising ':'Falling ')+Math.abs(momentum.rate).toFixed(3)+'%/min since preceding distinct quote';
}
// The extended column color measures performance relative to the regular-session
// price displayed in the Price column, not the change between refreshes.
function shadeExtendedVsClose(cell,q){
 const pct=q.extended_change_percent;
 if(!Number.isFinite(pct)||Math.abs(pct)<1e-10)return;
 // Go computes percentage change from the displayed regular price,
 // rounded to cents. The browser only chooses how to paint the cell.
 cell.classList.add('price-momentum');
 cell.style.color=pct>0?'#137a3b':'#bf3037';
 cell.title+=(cell.title?' • ':'')+(pct>0?'+':'')+pct.toFixed(2)+'% versus displayed regular price';
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
const numericKeys=new Set(['display_price','extended_price','change','change_percent','day_position','week_52_position']);
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
const researchSelect=document.getElementById('research-site');
const researchSites=new Set(['google','stockanalysis','corporate','ir']);
const savedResearchSite=localStorage.getItem('mauri_research_site');
if(researchSites.has(savedResearchSite))researchSelect.value=savedResearchSite;
researchSelect.addEventListener('change',()=>{localStorage.setItem('mauri_research_site',researchSelect.value);renderQuotes()});
const irURLs={
 MU:'https://investors.micron.com', AMAT:'https://ir.appliedmaterials.com',
 LRCX:'https://investor.lamresearch.com', SPCX:'https://ir.spacex.com/',
 NVDA:'https://investor.nvidia.com', TSLA:'https://ir.tesla.com',
 AVGO:'https://investors.broadcom.com', O:'https://www.realtyincome.com/investors',
 TSM:'https://investor.tsmc.com/', PLTR:'https://investors.palantir.com/',
 LMT:'https://investors.lockheedmartin.com/', AMGN:'https://investors.amgen.com/',
 SNA:'https://investors.snapon.com/', JOBY:'https://ir.jobyaviation.com/'
};
function researchURL(q){
 const sym=String(q.symbol||'').toUpperCase();
 if(!/^[A-Z0-9][A-Z0-9.^=-]{0,17}$/.test(sym))return '';
 switch(researchSelect.value){
 case 'google':return q.exchange?'https://www.google.com/finance/quote/'+encodeURIComponent(sym)+':'+encodeURIComponent(q.exchange):'';
 case 'stockanalysis':return 'https://stockanalysis.com/'+(['FTEC','JEPQ','VOO','IAU','FSTA','IBIT'].includes(sym)?'etf/':'stocks/')+encodeURIComponent(sym.toLowerCase())+'/';
 case 'corporate':return '/company?symbol='+encodeURIComponent(sym);
 case 'ir':return irURLs[sym]||'';
 default:return '';
 }
}
function symbolCell(tr,q){
 const td=document.createElement('td');td.className='symbol-cell';
 const symbol=String(q.symbol||'').toUpperCase();
 const flag=document.createElement('button');flag.type='button';flag.className='symbol-flag';
 flag.textContent=flaggedSymbols.has(symbol)?'★':'☆';
 flag.setAttribute('aria-pressed',String(flaggedSymbols.has(symbol)));
 flag.setAttribute('aria-label',(flaggedSymbols.has(symbol)?'Unflag ':'Flag ')+symbol);
 flag.title=(flaggedSymbols.has(symbol)?'Remove highlight from ':'Highlight ')+symbol;
 flag.addEventListener('click',()=>toggleFlag(symbol));td.appendChild(flag);
 const link=researchURL(q);
 if(link){const a=document.createElement('a');a.className='symbol-link';a.href=link;a.target='_blank';a.rel='noopener noreferrer';a.textContent=q.symbol;a.title='Open '+q.symbol+' on '+researchSelect.selectedOptions[0].textContent;td.appendChild(a)}
 else {const span=document.createElement('span');span.className='symbol-unavailable';span.textContent=q.symbol;span.title='No verified link for this stock and research destination';td.appendChild(span)}
 tr.appendChild(td);
}
function renderQuotes(){
 rows.replaceChildren();
 for(const q of [...quotes].sort(compareQuotes)){
  let tr=document.createElement('tr');if(q.error)tr.classList.add('error');if(flaggedSymbols.has(String(q.symbol||'').toUpperCase()))tr.classList.add('flagged-row');
  symbolCell(tr,q);
  // Keep full exchange metadata for research URLs, but use shorter labels in the table.
  const exLabel=({'NYSEARCA':'ARCA','NYSEAMERICAN':'NYSE Am.'})[q.exchange]||q.exchange||'—';
  const exCell=cell(tr,exLabel,'exchange-cell');
  exCell.title=q.exchange||'Exchange unavailable';
  const nameCell=cell(tr,q.name||'','optional company-name');
  nameCell.title=q.name||'';

  const priceCell=cell(tr,money(q.display_price,q.currency),'num');
  const extendedCell=cell(tr,money(q.extended_price,q.currency),'num extended-column');
  if(q.extended_price!=null){
   extendedCell.title=(q.extended_session||'extended')+' quote: '+barTime(q.extended_time);
  }else{extendedCell.title='No extended-hours price reported';}
  shadeMomentum(priceCell,priceMomentum.get(q.symbol),q.display_price);
  shadeExtendedVsClose(extendedCell,q);
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
  const sortButton=h.querySelector('button');
  const indicator=document.createElement('span');
  indicator.className='sort-indicator'+(active?' active':'');
  indicator.textContent=active?(sortDirection==='asc'?'↑':'↓'):'↕';
  indicator.setAttribute('aria-hidden','true');
  sortButton.replaceChildren(document.createTextNode(h.dataset.label+' '),indicator);
  sortButton.setAttribute('aria-label','Sort by '+h.dataset.label+(active?', currently '+(sortDirection==='asc'?'ascending':'descending'):', not currently sorted'));
  sortButton.title=active?'Currently sorted '+(sortDirection==='asc'?'ascending':'descending')+'; click to reverse':'Click to sort by '+h.dataset.label;
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
 if(activeWatchlist!==dynamicName&&!symbols.length){message.textContent='Enter one or more symbols.';return}
 if(activeWatchlist!==dynamicName){input.value=symbols.join(',');localStorage.setItem('mauri_stock_symbols',input.value)}
 button.disabled=true;button.classList.add('fetching');button.setAttribute('aria-busy','true');fetchStatus.textContent='Fetching quotes';
 try{
  if(activeWatchlist===dynamicName){
   const ranked=await fetch('/api/most-active',{cache:'no-store'});
   if(!ranked.ok)throw new Error('Most Active ranking unavailable: '+(await ranked.text()).trim());
   const listing=await ranked.json();
   if(!Array.isArray(listing.symbols)||listing.symbols.length!==20)throw new Error('Invalid Most Active ranking');
   symbols=listing.symbols;input.value=symbols.join(',');
  }
  let response=await fetch('/api/quotes?symbols='+encodeURIComponent(symbols.join(',')),{cache:'no-store'});
  if(!response.ok)throw new Error(await response.text());
  let data=await response.json();quotes=data.quotes||[];updateMomentum(quotes);
  const activeSession=quotes.find(q=>q.session==='premarket'||q.session==='after-hours')?.session;
  // Use the server's session classification, so the table changes on the
  // first successful quote retrieval after the market opens or closes.
  const regularOnly=quotes.length>0 && quotes.every(q=>q.session==='regular');
  const quoteTable=document.getElementById('quotes-table');
  quoteTable.classList.toggle('regular-session',regularOnly);
  if(regularOnly && sortKey==='extended_price'){
   // Hidden columns should not control the visible ordering.
   sortKey='display_price';sortDirection='desc';
   localStorage.setItem('mauri_stock_sort',JSON.stringify({key:sortKey,direction:sortDirection}));
  }

  // Session context is shown in the extended price header; changes have concise, consistent labels.
  for(const [id,label] of [['change-heading','Change $'],['change-percent-heading','Change %']]){
   const header=document.getElementById(id);
   header.dataset.label=label;
   header.querySelector('button').setAttribute('aria-label','Sort by '+label);
  }
  renderQuotes();
  const valid=quotes.filter(q=>q.display_price!=null);
  const extendedCount=quotes.filter(q=>q.extended_price!=null).length;
  const extHead=document.getElementById('extended-heading');extHead.dataset.label=quotes.some(q=>q.session==='premarket')?'Premarket Price':'After-hours Price';
  const sessions=[...new Set(valid.map(q=>q.session).filter(Boolean))];
  const session=sessions.length===1?sessions[0]:sessions.length?'Mixed ('+sessions.join(', ')+')':'Unavailable';
  const times=valid.map(q=>q.display_time).filter(Boolean).map(t=>new Date(t).getTime()).filter(Number.isFinite);
  const latest=times.length?new Date(Math.max(...times)).toISOString():null;
  message.textContent='Session: '+session+' • Latest quote time: '+barTime(latest)+' • Retrieved: '+retrievedTime(data.fetched_at,latest)+' • '+valid.length+'/'+quotes.length+' regular prices • '+extendedCount+' extended prices.';
 }catch(e){message.textContent='Refresh failed: '+e.message+' • Existing quotes retained.'}finally{button.disabled=false;button.classList.remove('fetching');button.removeAttribute('aria-busy');fetchStatus.textContent=''}
}
button.addEventListener('click',refreshQuotes);
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
// Populate the restored watchlist immediately, even when auto refresh is Never.
refreshQuotes();
</script></body></html>`


// Corporate website lookup is deliberately separate from the quote refresh path.
// Finnhub's API key never leaves the server. Results (including misses) are cached.
type companySite struct {
 Symbol string `json:"symbol"`
 Name string `json:"name,omitempty"`
 Website string `json:"website,omitempty"`
 Source string `json:"source,omitempty"`
 Error string `json:"error,omitempty"`
}
type siteCacheEntry struct {site companySite; expires time.Time}
var siteCache = struct {sync.Mutex; entries map[string]siteCacheEntry}{entries:make(map[string]siteCacheEntry)}
var knownCorporateSites = map[string]string{
 "MU":"https://www.micron.com", "AMAT":"https://www.appliedmaterials.com",
 "LRCX":"https://www.lamresearch.com", "SPCX":"https://www.spacex.com/",
 "AVGO":"https://www.broadcom.com", "TSM":"https://www.tsmc.com",
 "TSLA":"https://www.tesla.com", "NVDA":"https://www.nvidia.com",
 "JOBY":"https://www.jobyaviation.com", "PLTR":"https://www.palantir.com",
 "LMT":"https://www.lockheedmartin.com", "AMGN":"https://www.amgen.com",
 "SNA":"https://www.snapon.com", "O":"https://www.realtyincome.com",
 "FTEC":"https://www.fidelity.com/", "FSTA":"https://www.fidelity.com/",
 "JEPQ":"https://am.jpmorgan.com/", "VOO":"https://investor.vanguard.com/",
 "IAU":"https://www.ishares.com/", "IBIT":"https://www.ishares.com/",
}
// Refuse non-web URLs, credentials, local hosts and explicit nonstandard ports.
func safeCompanyURL(raw string) string {
 u, err := url.Parse(strings.TrimSpace(raw))
 if err != nil || (u.Scheme!="https" && u.Scheme!="http") || u.Hostname()=="" || u.User!=nil || u.Port()!="" {return ""}
 host:=strings.ToLower(u.Hostname())
 if host=="localhost" || strings.HasSuffix(host,".local") || strings.HasSuffix(host,".internal") || !strings.Contains(host,".") {return ""}
 if ip:=net.ParseIP(host);ip!=nil {return ""}
 return u.String()
}
func companyProfile(ctx context.Context, symbol string) companySite {
 siteCache.Lock()
 if entry,ok:=siteCache.entries[symbol];ok && time.Now().Before(entry.expires){siteCache.Unlock();return entry.site}
 siteCache.Unlock()
 site:=companySite{Symbol:symbol}
 ttl:=time.Hour
 key:=strings.TrimSpace(os.Getenv("FINNHUB_API_KEY"))
 if key!="" {
  endpoint:="https://finnhub.io/api/v1/stock/profile2?symbol="+url.QueryEscape(symbol)
  req,err:=http.NewRequestWithContext(ctx,http.MethodGet,endpoint,nil)
  if err==nil {
   req.Header.Set("X-Finnhub-Token",key)
   req.Header.Set("Accept","application/json")
   resp,err:=client.Do(req)
   if err==nil {
    if resp.StatusCode==http.StatusOK {
     var result struct {Ticker string `json:"ticker"`; Name string `json:"name"`; WebURL string `json:"weburl"`}
     if json.NewDecoder(io.LimitReader(resp.Body,65536)).Decode(&result)==nil && strings.EqualFold(result.Ticker,symbol) {
      if address:=safeCompanyURL(result.WebURL);address!="" {site.Website=address;site.Name=result.Name;site.Source="Finnhub";ttl=7*24*time.Hour}
     }
    } else {log.Printf("Finnhub company profile lookup for %s: HTTP %d",symbol,resp.StatusCode)}
    resp.Body.Close()
   }
  }
 }
 if site.Website=="" {if fallback:=safeCompanyURL(knownCorporateSites[symbol]);fallback!="" {site.Website=fallback;site.Source="verified mapping";ttl=7*24*time.Hour}}
 if site.Website=="" {site.Error="No corporate website available for this symbol"}
 siteCache.Lock();siteCache.entries[symbol]=siteCacheEntry{site:site,expires:time.Now().Add(ttl)};siteCache.Unlock()
 return site
}
func companyHandler(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet {http.Error(w,"GET only",http.StatusMethodNotAllowed);return}
 symbol:=strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
 if !symbolRE.MatchString(symbol){http.Error(w,"invalid symbol",http.StatusBadRequest);return}
 ctx,cancel:=context.WithTimeout(r.Context(),13*time.Second);defer cancel()
 site:=companyProfile(ctx,symbol)
 w.Header().Set("Cache-Control","private, no-store")
 w.Header().Set("Content-Type","application/json")
 if site.Website=="" {w.WriteHeader(http.StatusNotFound)}
 _=json.NewEncoder(w).Encode(site)
}
func companyRedirectHandler(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet {http.Error(w,"GET only",http.StatusMethodNotAllowed);return}
 symbol:=strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
 if !symbolRE.MatchString(symbol){http.Error(w,"invalid symbol",http.StatusBadRequest);return}
 ctx,cancel:=context.WithTimeout(r.Context(),13*time.Second);defer cancel()
 site:=companyProfile(ctx,symbol)
 if site.Website=="" {http.Error(w,"No corporate website verified for "+symbol,http.StatusNotFound);return}
 w.Header().Set("Cache-Control","no-store")
 http.Redirect(w,r,site.Website,http.StatusFound)
}

const appVersion = "1.0.0"
const buildVersion = "v39"

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
	if os.Getenv("FINNHUB_API_KEY")!="" {log.Print("Finnhub company profiles: enabled")} else {log.Print("Finnhub company profiles: disabled (verified mappings only)")}
 addr := net.JoinHostPort(*host, *port)
	tmpl := template.Must(template.New("index").Parse(page))
	mux := http.NewServeMux()
	mux.HandleFunc("/api/quotes", quotesHandler)
 mux.HandleFunc("/api/company", companyHandler)
 mux.HandleFunc("/company", companyRedirectHandler)
 mux.HandleFunc("/api/most-active", mostActiveHandler)
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
