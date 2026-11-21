package main

import (
 "strings"
 "testing"
 "os"
)

func TestStocksAllDataRoutesRequireSharedGuard(t *testing.T){
 for _,name:=range []string{
  "quotesHandler","marketCapsHandler","topMarketHandler",
  "companyHandler","companyRedirectHandler","mostActiveHandler",
 }{
  if !strings.Contains(mainSourceForAuthTest(),"track(guardStocksAPI("+name+"))"){
   t.Errorf("%s is missing shared authentication gate",name)
  }
 }
}

// The dashboard is a public handoff shell: it cannot show stock content or
// send API requests until a verified Homepage personal/guest profile is seen.
func TestStocksBrowserRequiresVerifiedHomepage(t *testing.T){
 page:=pageTemplate
 for _,want:=range []string{
  "html:not(.suite-auth-ready) .stocks-content{visibility:hidden}",
  "function stockFetch(path,options={})",
  "if(!suiteAccessVerified||!suiteHandoff)",
  "Authorization:'Bearer '+suiteHandoff",
  "activateVerifiedStocks();",
  "if(suiteBlockedLocalMode)return;",
  "location.replace(suiteBase+'/#suite_signin=required')",
  "if(!identity||(data.role!=='user'&&data.role!=='admin'))",
  "stockFetch('/api/quotes?",
  "stockFetch('/api/most-active'",
  "stockFetch('/api/market-top?",
  "stockFetch('/api/market-caps?",
 }{
  if !strings.Contains(page,want){t.Errorf("missing fail-closed client contract %q",want)}
 }
 if strings.Contains(page,"explicitBrowserOnly=true") ||
    strings.Contains(page,"Not signed in · Browser-only settings"){
  t.Fatal("anonymous browser-only access or label must not return")
 }
}

func mainSourceForAuthTest() string {
 // main.go source is embedded in the browser page only for JavaScript.
 // Route registrations are asserted from the checked-in source file, not
 // approximated by probing remote production.
 contents,err:=os.ReadFile("main.go")
 if err!=nil{return ""}
 return string(contents)
}
