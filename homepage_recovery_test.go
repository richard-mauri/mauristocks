package main

import (
 "os"
 "strings"
 "testing"
)

// The persistent session is owned by Homepage. Stocks may cache a temporary
// handoff in this tab but must recover it on the next direct visit.
func TestHomepageSessionRecoveryFromDirectStocksVisit(t *testing.T) {
 raw,err:=os.ReadFile("main.go")
 if err!=nil{t.Fatal(err)}
 js:=string(raw)
 for _,part:=range []string{
  "const suiteBase='https://home.themauris.org'",
  "const suiteHandoffKey='mauristocks.homepage.handoff.v1'",
  "fragment.get('suite_mode')==='local'",
  "explicitBrowserOnly=true",
  "if(!suiteHandoff && !explicitBrowserOnly)",
  "location.replace(suiteBase+'/launch/stocks')",
  "sessionStorage.setItem(suiteHandoffKey,delivered)",
  "sessionStorage.removeItem(suiteHandoffKey)",
  "await suiteRequest('/api/session/profile')",
 }{
  if !strings.Contains(js,part){t.Errorf("missing Homepage-only recovery guard: %s",part)}
 }
 for _,forbidden:=range []string{
  "/api/handoff/extend-session", "/api/handoff/renew",
  "/api/session/renew", "suiteRenewButton",
 }{
  if strings.Contains(js,forbidden){t.Errorf("Stocks must not manage parent session lifetime: %s",forbidden)}
 }
}

func TestCachedExpiryRevalidatesHomepageBeforeDiscardingHandoff(t *testing.T) {
 raw,err:=os.ReadFile("main.go")
 if err!=nil{t.Fatal(err)}
 js:=string(raw)
 a:=strings.Index(js,"function updateSuiteAccountClock(){")
 b:=strings.Index(js,"setInterval(updateSuiteAccountClock,1000)",a)
 if a<0||b<0{t.Fatal("session display code not found")}
 clock:=js[a:b]
 if strings.Contains(clock,"sessionStorage.removeItem(suiteHandoffKey)")||
    strings.Contains(clock,"suiteHandoff=''"){
  t.Fatal("cached expiry must not revoke a potentially extended live bearer")
 }
 if !strings.Contains(clock,"refreshSuiteSession().finally")||
    !strings.Contains(clock,"expiryRetryAfter=Date.now()+30000"){
  t.Fatal("expired cached handoff must be revalidated with bounded retries")
 }
}
