package main

import (
 "os/exec"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
)

// The Stocks dashboard embeds JavaScript in a Go raw string. Go compilation
// cannot detect malformed JavaScript, which can otherwise blank the entire
// quote dashboard, identity header and local watchlists.
func TestEmbeddedDashboardJavaScriptParses(t *testing.T) {
 source:=page
 start:=strings.Index(source,"<script>")
 if start<0{t.Fatal("dashboard script tag missing")}
 start+=len("<script>")
 end:=strings.Index(source[start:],"</script>")
 if end<0{t.Fatal("dashboard script closing tag missing")}
 script:=source[start:start+end]
 if strings.TrimSpace(script)==""{t.Fatal("dashboard script empty")}
 node,err:=exec.LookPath("node")
 if err!=nil{
  t.Skip("node unavailable; manual JavaScript parsing needed on this machine")
 }
 cmd:=exec.Command(node,"--check")
 cmd.Stdin=strings.NewReader(script)
 output,err:=cmd.CombinedOutput()
 if err!=nil{t.Fatalf("embedded dashboard JavaScript syntax invalid: %v\n%s",err,output)}
}

func TestMarketCapEndpointRequiresServerKey(t *testing.T) {
 t.Setenv("FINNHUB_API_KEY","")
 req:=httptest.NewRequest(http.MethodGet,"/api/market-caps?symbols=NVDA",nil)
 rec:=httptest.NewRecorder()
 marketCapsHandler(rec,req)
 if rec.Code!=http.StatusServiceUnavailable {t.Fatalf("expected 503, got %d",rec.Code)}
}
func TestMarketCapEndpointRejectsOtherMethods(t *testing.T) {
 req:=httptest.NewRequest(http.MethodPost,"/api/market-caps?symbols=NVDA",nil)
 rec:=httptest.NewRecorder()
 marketCapsHandler(rec,req)
 if rec.Code!=http.StatusMethodNotAllowed {t.Fatalf("expected 405, got %d",rec.Code)}
}
func TestMarketCapEndpointRejectsBadSymbols(t *testing.T) {
 t.Setenv("FINNHUB_API_KEY","test-only")
 req:=httptest.NewRequest(http.MethodGet,"/api/market-caps?symbols=INVALID!",nil)
 rec:=httptest.NewRecorder()
 marketCapsHandler(rec,req)
 if rec.Code!=http.StatusBadRequest {t.Fatalf("expected 400, got %d",rec.Code)}
}

 // Regression: changing Symbol links must refresh existing heatmap anchor hrefs
 // immediately, not just table rows or after a subsequent price refresh.
func TestResearchSiteChangeUpdatesHeatmapLinks(t *testing.T) {
 scriptStart:=strings.Index(page,"<script>")
 if scriptStart<0 {t.Fatal("dashboard script missing")}
 script:=page[scriptStart:]
 changeHandler:="researchSelect.addEventListener('change',()=>{"
 start:=strings.Index(script,changeHandler)
 if start<0 {t.Fatal("Symbol links change listener missing")}
 end:=strings.Index(script[start:],"});")
 if end<0 {t.Fatal("Symbol links change listener malformed")}
 handler:=script[start:start+end]
 if !strings.Contains(handler,"renderQuotes()") || !strings.Contains(handler,"paintHeatmap()") {
  t.Fatal("Symbol links change must update both table and existing heatmap tile links")
 }
}

// Regression: an adjustable exponent must drive treemap geometry while
// preserving raw capitalization values for labels and research tooltips.
func TestHeatmapWeightingControlAndSizing(t *testing.T) {
 required:=[]string{
  `id="heat-weight" type="range" min="0.50" max="1.00" step="0.05" value="0.65"`,
  `id="heat-weight-value"`,
  `mauristocks_heat_weight_v1`,
  `heatWeightSlider.addEventListener('input'`,
  `localStorage.setItem(heatWeightKey,next)`,
  `Math.pow(caps.get(q.symbol),heatExponent)`,
  `v+e.weight`,
  `sum+=a[i].weight`,
  `heatExponent.toFixed(2)+'|'`,
  `cap:caps.get(q.symbol)`,
 }
 for _,item:=range required{
  if !strings.Contains(page,item){t.Errorf("missing heatmap weighting behavior: %s",item)}
 }
 if !strings.Contains(page,`heatExponent===1?'Tile area: exact market cap'`){
  t.Error("uncompressed tile area must be labeled as exact market cap")
 }
}

// A cap refresh or weight-slider adjustment must never blank/rebuild all tiles.
// Existing DOM anchors, colors, keyboard focus and links should survive a relayout.
func TestHeatmapRelayoutRetainsExistingTiles(t *testing.T) {
 start:=strings.Index(page,"function paintHeatmap(){")
 end:=strings.Index(page[start:],"// Cache capitalization by symbol")
 if start<0||end<0{t.Fatal("heatmap paint function not found")}
 paint:=page[start:start+end]
 if strings.Contains(paint,"ha.replaceChildren()")||strings.Contains(paint,"tiles.clear()"){
  t.Fatal("heatmap relayout clears the whole panel, causing refresh flash")
 }
 for _,required:=range []string{
  "let el=tiles.get(symbol)",
  "if(!el){",
  "el.remove();tiles.delete(symbol)",
  "el.style.left=r.x+'%'",
  "el.style.width=r.w+'%'",
 }{
  if !strings.Contains(paint,required){t.Errorf("missing tile reconciliation behavior %s",required)}
 }
}

// All user-editable browser preferences must use the same per-account,
// revision-checked Homepage settings record. Provider metadata caches do not
// belong in that record. Preserve compatibility with old schema:1 settings.
func TestAccountSettingsIncludeEveryDisplayPreference(t *testing.T) {
 start:=strings.Index(page,"const suiteKeys=[")
 if start<0{t.Fatal("suite account preference list missing")}
 end:=strings.Index(page[start:],"];")
 if end<0{t.Fatal("suite account preference list not terminated")}
 keys:=page[start:start+end]
 for _,key:=range []string{
  "wlKey","activeKey","flagStorageKey",
  "'mauri_stock_symbols'","'mauri_stock_sort'",
  "'mauri_stock_refresh_interval'","'mauri_research_site'",
  "'mauristocks_columns_v1'","'mauristocks_market_v1'",
  "'mauristocks_view_v1'","heatWeightKey",
 }{
  if !strings.Contains(keys,key){t.Errorf("user preference not included in account record: %s",key)}
 }
 if strings.Contains(keys,"capRecordKey")||strings.Contains(keys,"mauristocks_cap_cache_v1"){
  t.Fatal("stock metadata caches must remain browser-local")
 }
 if !strings.Contains(page,"return {schema:1,settings};") {
  t.Fatal("legacy account settings schema compatibility lost")
 }
}
func TestDisplayPreferenceEventsQueueAccountSave(t *testing.T) {
 checks:=[]struct{start,end string}{
  {"marketSelect.addEventListener('change',()=>{","wlSelect.addEventListener('change'"},
  {"heatWeightSlider.addEventListener('input',()=>{","let heatView="},
  {"heatSelector.addEventListener('change',()=>{","new ResizeObserver"},
 }
 for _,check:=range checks{
  start:=strings.Index(page,check.start)
  if start<0{t.Errorf("missing preference listener: %s",check.start);continue}
  end:=strings.Index(page[start:],check.end)
  if end<0{t.Errorf("missing listener boundary: %s",check.end);continue}
  body:=page[start:start+end]
  if !strings.Contains(body,"queueSuiteSave()"){
   t.Errorf("preference listener does not request account synchronization: %s",check.start)
  }
 }
 if !strings.Contains(page,"if(!suiteSignedIn)return;"){
  t.Fatal("browser-only users must not send account writes")
 }
 if !strings.Contains(page,"if(response.status===409)"){
  t.Fatal("account revision-conflict protection missing")
 }
}

// The save/load buttons were removed in favor of browser autosave and a guarded
// account check. Assert that future UI edits cannot regress to manual saving.
func TestWatchlistBrowserAutosaveAndNoManualAccountButtons(t *testing.T) {
 if strings.Contains(page,`id="watchlist-save"`) || strings.Contains(page,`id="suite-sync-load"`) {
  t.Fatal("manual Save Changes / Load from account controls must not be present")
 }
 checks:=[]string{
  "input.addEventListener('input',",
  "watchlistSaveTimer=setTimeout(saveActiveWatchlist,600)",
  "input.addEventListener('change',saveActiveWatchlist)",
  "wlSelect.addEventListener('change',()=>",
  "saveActiveWatchlist();",
  "persistLists();",
  "queueSuiteSave();",
  "reconcileAccountOnSignIn(identity)",
  "suiteAsk('Saved account settings differ from this browser.",
 }
 for _,check:=range checks {
  if !strings.Contains(page,check){t.Errorf("missing automatic persistence behavior: %s",check)}
 }
}
