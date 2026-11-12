package main

// Market index selections are informational snapshots, not official index feeds.
// Membership comes from independently maintained public lists; capitalization
// comes from the daily Nasdaq screener snapshot. No third-party API key is needed.
import (
 "context"
 "encoding/csv"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "math"
 "net/http"
 "net/url"
 "sort"
 "strconv"
 "strings"
 "sync"
 "time"
)

const marketTopCount = 30

type rankedMarketCap struct {
 Symbol string `json:"symbol"`
 MarketCap float64 `json:"market_cap"`
}
type marketTopResponse struct {
 Market string `json:"market"`
 Symbols []string `json:"symbols"`
 Items []rankedMarketCap `json:"items"`
 Constituents int `json:"constituents"`
 WithCaps int `json:"with_caps"`
 RankedAt time.Time `json:"ranked_at"`
 Source string `json:"source"`
}
var topMarketCache = struct {
 sync.Mutex
 entries map[string]marketTopResponse
 expires time.Time
}{entries:make(map[string]marketTopResponse)}

var marketDataURLs = map[string]string{
 "all":"https://raw.githubusercontent.com/zyhe16/top-us-stock-tickers/main/tickers/all.csv",
 "sp500":"https://raw.githubusercontent.com/zyhe16/top-us-stock-tickers/main/tickers/sp500.csv",
 "nasdaq100":"https://yfiua.github.io/index-constituents/constituents-nasdaq100.csv",
 "dow30":"https://yfiua.github.io/index-constituents/constituents-dowjones.csv",
}

func fetchMarketCSV(ctx context.Context, endpoint string, maxBytes int64) ([][]string,error) {
 u,err:=url.Parse(endpoint)
 if err!=nil || u.Scheme!="https" || u.Host=="" {return nil,errors.New("invalid market-data source")}
 req,err:=http.NewRequestWithContext(ctx,http.MethodGet,endpoint,nil)
 if err!=nil{return nil,err}
 req.Header.Set("User-Agent","MauriStocks/1.0 (personal financial dashboard)")
 req.Header.Set("Accept","text/csv,text/plain")
 resp,err:=client.Do(req)
 if err!=nil{return nil,err}
 defer resp.Body.Close()
 if resp.StatusCode!=http.StatusOK{return nil,fmt.Errorf("market data source HTTP %d",resp.StatusCode)}
 reader:=csv.NewReader(io.LimitReader(resp.Body,maxBytes+1))
 reader.FieldsPerRecord=-1
 records,err:=reader.ReadAll()
 if err!=nil{return nil,err}
 // The cap bounds memory usage even if an upstream server ignores Content-Length.
 if resp.ContentLength>maxBytes {return nil,errors.New("market snapshot too large")}
 return records,nil
}

// parseMarketCaps parses the Nasdaq snapshot without interpreting quote prices as live.
func parseMarketCaps(records [][]string) (map[string]float64,error) {
 if len(records)<2{return nil,errors.New("empty capitalization snapshot")}
 columns:=make(map[string]int)
 for i,name:=range records[0]{columns[strings.ToLower(strings.TrimSpace(name))]=i}
 symbolCol,okSymbol:=columns["symbol"];capCol,okCap:=columns["marketcap"]
 if !okSymbol||!okCap{return nil,errors.New("market snapshot missing symbol or marketCap column")}
 caps:=make(map[string]float64)
 for _,record:=range records[1:]{
  if len(record)<=symbolCol||len(record)<=capCol{continue}
  symbol:=strings.ToUpper(strings.TrimSpace(record[symbolCol]))
  if !symbolRE.MatchString(symbol){continue}
  v,err:=strconv.ParseFloat(strings.TrimSpace(record[capCol]),64)
  if err!=nil||v<=0||math.IsNaN(v)||math.IsInf(v,0){continue}
  caps[symbol]=v
 }
 if len(caps)<100{return nil,fmt.Errorf("too few valid market-cap rows: %d",len(caps))}
 return caps,nil
}
func parseIndexSymbols(records [][]string) ([]string,error) {
 if len(records)<2{return nil,errors.New("empty index membership snapshot")}
 if !strings.EqualFold(strings.TrimSpace(records[0][0]),"symbol"){return nil,errors.New("index snapshot missing Symbol header")}
 var out []string
 seen:=make(map[string]bool)
 for _,record:=range records[1:]{
  if len(record)==0{continue}
  symbol:=strings.ToUpper(strings.TrimSpace(record[0]))
  if !symbolRE.MatchString(symbol)||seen[symbol]{continue}
  seen[symbol]=true
  out=append(out,symbol)
 }
 if len(out)<30{return nil,fmt.Errorf("too few index constituents: %d",len(out))}
 return out,nil
}
func rankMarketTop(market string,members []string,caps map[string]float64,now time.Time) (marketTopResponse,error) {
 selected:=make([]rankedMarketCap,0,len(members))
 for _,symbol:=range members{
  if capValue:=caps[symbol];capValue>0{
   selected=append(selected,rankedMarketCap{Symbol:symbol,MarketCap:capValue})
  }
 }
 sort.Slice(selected,func(i,j int)bool{
  if selected[i].MarketCap==selected[j].MarketCap{return selected[i].Symbol<selected[j].Symbol}
  return selected[i].MarketCap>selected[j].MarketCap
 })
 if len(selected)<marketTopCount{return marketTopResponse{},fmt.Errorf("market %s has only %d constituents with valid market cap",market,len(selected))}
 response:=marketTopResponse{Market:market,Constituents:len(members),WithCaps:len(selected),RankedAt:now.UTC(),Source:"Nasdaq screener snapshot (zyhe16); index membership (zyhe16/yfiua)"}
 if market=="dow30"{
  // Dow has 30 constituents. Keep every symbol even when its capitalization
  // is unavailable, so none silently drops out of the index.
  selected=selected[:0]
  for _,symbol:=range members{
   if v:=caps[symbol];v>0{selected=append(selected,rankedMarketCap{Symbol:symbol,MarketCap:v})}
  }
  sort.Slice(selected,func(i,j int)bool{return selected[i].MarketCap>selected[j].MarketCap})
  for _,item:=range selected{response.Symbols=append(response.Symbols,item.Symbol)}
  // Preserve any member missing from the capitalizations list, at the end.
  for _,symbol:=range members{if caps[symbol]<=0{response.Symbols=append(response.Symbols,symbol)}}
 }else{
  selected=selected[:marketTopCount]
  for _,item:=range selected{response.Symbols=append(response.Symbols,item.Symbol)}
 }
 response.Items=selected
 return response,nil
}
func getTopMarket(ctx context.Context,market string) (marketTopResponse,error) {
 topMarketCache.Lock()
 defer topMarketCache.Unlock()
 if cached,ok:=topMarketCache.entries[market];ok&&time.Now().Before(topMarketCache.expires){return cached,nil}
 names:=[]string{"all","sp500","nasdaq100","dow30"}
 records:=make(map[string][][]string)
 for _,name:=range names{
  data,err:=fetchMarketCSV(ctx,marketDataURLs[name],2<<20)
  if err!=nil{
   // A transient upstream outage must not erase a previously good snapshot.
   if cached,ok:=topMarketCache.entries[market];ok{return cached,nil}
   return marketTopResponse{},fmt.Errorf("%s source: %w",name,err)
  }
  records[name]=data
 }
 caps,err:=parseMarketCaps(records["all"]);if err!=nil{return marketTopResponse{},err}
 spCaps,err:=parseMarketCaps(records["sp500"]);if err!=nil{return marketTopResponse{},err}
 // The S&P feed is the membership source, the Nasdaq snapshot is the
 // capitalization source; use its daily values for all three markets.
 for symbol,capValue:=range spCaps{if caps[symbol]==0{caps[symbol]=capValue}}
 memberships:=make(map[string][]string)
 for symbol:=range spCaps{memberships["sp500"]=append(memberships["sp500"],symbol)}
 for _,name:=range []string{"nasdaq100","dow30"}{
  members,err:=parseIndexSymbols(records[name]);if err!=nil{return marketTopResponse{},err}
  memberships[name]=members
 }
 now:=time.Now().UTC()
 next:=make(map[string]marketTopResponse)
 for _,name:=range []string{"sp500","nasdaq100","dow30"}{
  result,err:=rankMarketTop(name,memberships[name],caps,now)
  if err!=nil{return marketTopResponse{},err}
  next[name]=result
 }
 topMarketCache.entries=next
 topMarketCache.expires=now.Add(6*time.Hour)
 return next[market],nil
}
func topMarketHandler(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{http.Error(w,"GET only",http.StatusMethodNotAllowed);return}
 market:=r.URL.Query().Get("market")
 if market!="sp500"&&market!="nasdaq100"&&market!="dow30"{
  http.Error(w,"unknown market (sp500, nasdaq100, dow30)",http.StatusBadRequest);return
 }
 ctx,cancel:=context.WithTimeout(r.Context(),50*time.Second);defer cancel()
 response,err:=getTopMarket(ctx,market)
 if err!=nil{http.Error(w,err.Error(),http.StatusBadGateway);return}
 w.Header().Set("Content-Type","application/json; charset=utf-8")
 w.Header().Set("Cache-Control","no-store")
 _=json.NewEncoder(w).Encode(response)
}
