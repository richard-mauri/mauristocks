package main

import (
 "os"
 "strings"
 "testing"
)

func TestDashboardHasSameTabHomepageReturn(t *testing.T) {
 code,err:=os.ReadFile("main.go")
 if err!=nil{t.Fatal(err)}
 source:=string(code)
 if !strings.Contains(page,`class="suite-banner__home suite-type-control" href="https://home.themauris.org/"`) {
  t.Fatal("dashboard missing same-tab Back to Homepage anchor")
 }
 if strings.Contains(page,`class="suite-banner__home suite-type-control" target="_blank"`) {
  t.Fatal("Homepage return should not open another tab")
 }
 if strings.Contains(source,"suiteRenewButton") || strings.Contains(source,"renewSuiteAccount(") {t.Fatal("no separate app-token renewal UI is permitted")}
 if !strings.Contains(source,"fragment.get('suite_mode')==='local'") || !strings.Contains(source,"sessionStorage.removeItem(suiteHandoffKey)") {t.Fatal("browser-only launch must discard stale handoff")}
 if strings.Contains(source,"suiteRenewButton.textContent=suiteSignedIn&&left>0?'Renew':'Open Homepage'") {t.Fatal("redundant Open Homepage action remains")}
 if !strings.Contains(source,"Not signed in · Browser-only settings") {t.Fatal("signed-out state feedback missing")}
 if !strings.Contains(page,".suite-banner__home:hover{text-decoration:underline}") {t.Fatal("shared hover style missing")}

}

func TestHomepageAccountHeaderOmitsMinuteCountdown(t *testing.T) {
 code,err:=os.ReadFile("main.go")
 if err!=nil{t.Fatal(err)}
 source:=string(code)
 if !strings.Contains(source,"suiteIdentityEl.textContent='Signed in as '+suiteAccountIdentity;") {
  t.Fatal("signed-in header must show identity without a minute countdown")
 }
 if !strings.Contains(source,"if(left===0 && suiteSignedIn)") {
  t.Fatal("existing session expiry enforcement must remain")
 }
 if strings.Contains(source,"remaining';") {
  t.Fatal("signed-in account header still displays a countdown")
 }
}
