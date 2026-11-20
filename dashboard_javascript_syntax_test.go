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
  `stageSetting(heatWeightKey,next)`,
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
// User preferences remain pending until an explicit button action. No event
// handler may directly write account settings or start an automatic PUT.
func TestDisplayPreferenceEventsAreStagedUntilSaveChanges(t *testing.T) {
 required:=[]string{
  "const pendingSettings=new Map()",
  "function stageSetting(key,value)",
  "stageSetting('mauri_stock_refresh_interval',String(seconds))",
  "stageSetting('mauristocks_columns_v1',JSON.stringify(colPrefs))",
  "stageSetting('mauri_research_site',researchSelect.value)",
  "stageSetting('mauristocks_market_v1',marketMode)",
  "stageSetting('mauri_stock_sort',JSON.stringify({key:sortKey,direction:sortDirection}))",
  "stageSetting('mauristocks_view_v1',v)",
  "stageSetting(heatWeightKey,next)",
  "stageSetting(flagStorageKey,JSON.stringify([...flaggedSymbols].sort()))",
  "wlSave.addEventListener('click',()=>{void commitSettings()})",
  "async function commitSettings()",
  "async function saveWatchlistToAccount()",
  "accountSyncNeeded=!saved",
  "if(response.status===409)",
  "window.addEventListener('beforeunload'",
 }
 for _,fragment:=range required{
  if !strings.Contains(page,fragment){t.Errorf("explicit save workflow missing: %s",fragment)}
 }
 for _,forbidden:=range []string{"function queueSuiteSave()","function queueWatchlistAccountSave()","function flushSuiteSave()","setTimeout(()=>{suiteSaveTimer","suiteSaveQueued=true"}{
  if strings.Contains(page,forbidden){t.Errorf("automatic account save path remains: %s",forbidden)}
 }
}
 
// Symbol edits still need explicit Save Changes; all other preferences now
// participate in the same transaction and localStorage commit.
func TestWatchlistExplicitSaveAndRestoreControls(t *testing.T) {
 for _,id:=range []string{`id="watchlist-save"`,`id="watchlist-new"`,`id="watchlist-restore"`,`id="watchlist-delete"`}{
  if !strings.Contains(page,id){t.Errorf("missing watchlist control: %s",id)}
 }
 for _,removed:=range []string{`id="watchlist-select"`,`id="suite-sync-load"`,`<option value="watchlist">My Watchlists</option>`,"saveActiveWatchlist();","watchlistSaveTimer"}{
  if strings.Contains(page,removed){t.Errorf("obsolete watchlist control or autosave remains: %s",removed)}
 }
 required:=[]string{
  "function watchlistChanged()",
  "normalizeSymbols(input.value)!==normalizeSymbols(savedSymbols())",
  "wlSave.disabled=settingsSaveBusy||!(committableSymbols||pendingSettings.size>0||accountSyncNeeded)",
  "wlNew.disabled=settingsSaveBusy||indexed||!changed",
  "wlRestore.disabled=indexed||!changed",
  "input.readOnly=indexed||activeWatchlist===dynamicName",
  "marketSelect.replaceChildren()",
  "op.value='list:'+n",
  "if(watchlistChanged()&&!(await suiteAsk('Discard unsaved symbol changes and change Market?')))",
  "wlSave.addEventListener('click',",
  "wlRestore.addEventListener('click',",
  "wlNew.addEventListener('click',",
  "persistLists();",
  "wlSave.addEventListener('click',()=>{void commitSettings()})",
  "reconcileAccountOnSignIn(identity)",
  "suiteSettingsDiff(browser,account)",
 }
 for _,check:=range required {
  if !strings.Contains(page,check){t.Errorf("missing explicit watchlist behavior: %s",check)}
 }
}

func TestAccountDiffReviewIsReadOnlyAndPreservesSaveGuards(t *testing.T) {
 required:=[]string{
  `id="suite-dialog-diff"`,
  "suiteDialogDiff.replaceChildren()",
  "detail.textContent=value===null?'Not present':value",
  "suite-diff-pair",
  "grid-template-columns:minmax(0,1fr)",
  "if(key==='mauri_stock_symbols')continue;",
  "if(name==='Default'||name===dynamicName)continue;",
  "Added here",
  "Missing here",
  "suiteListPresence(local)",
  "suiteListPresence(remote)",
  "#suite-dialog-diff{flex:0 1 auto;min-height:0;overflow-y:auto;",
  "suiteDialogCancel.textContent=diffRows?'Keep Browser Settings':'Cancel'",
  "suiteDialogOK.textContent=diffRows?'Restore Account Settings':'OK'",
  "function suiteSettingsDiff(browser,account)",
  "function suiteSymbolDelta(browser,account)",
  "const backup={};for(const key of suiteKeys)backup[key]=localStorage.getItem(key)",
  "if(response.status===409)",
  "async function commitSettings()",
  "const saved=await saveWatchlistToAccount()",
  "function suiteFeedback(text){wlStatus.textContent=text;wlStatus.hidden=!text}",
  "Settings saved to browser and account.",
 }
 for _,check:=range required {
  if !strings.Contains(page,check){t.Errorf("missing guarded account comparison behavior: %s",check)}
 }
}

func TestQuoteRefreshDoesNotCommitPendingWatchlistEdits(t *testing.T) {
 start:=strings.Index(page,"async function refreshQuotes(){")
 if start<0{t.Fatal("quote refresh function not found")}
 end:=strings.Index(page[start:],"button.addEventListener('click',refreshQuotes)")
 if end<0{t.Fatal("quote refresh listener boundary not found")}
 body:=page[start:start+end]
 if strings.Contains(body,"saveActiveWatchlist()")||strings.Contains(body,"persistLists()")||strings.Contains(body,"localStorage.setItem('mauri_stock_symbols'"){
  t.Fatal("quote refresh must not persist pending symbol edits")
 }
 if !strings.Contains(body,"input.value=symbols.join(',');updateWatchlistButtons()"){
  t.Fatal("quote refresh must preserve changed-state feedback")
 }
}
