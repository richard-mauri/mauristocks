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
