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
 if !strings.Contains(source,`class="suite-home-link" href="https://home.themauris.org/"`) {
  t.Fatal("dashboard missing same-tab Back to Homepage anchor")
 }
 if strings.Contains(source,`class="suite-home-link" target="_blank"`) {
  t.Fatal("Homepage return should not open another tab")
 }
 if strings.Contains(source,"suiteRenewButton") || strings.Contains(source,"renewSuiteAccount(") {t.Fatal("no separate app-token renewal UI is permitted")}
 if !strings.Contains(source,"fragment.get('suite_mode')==='local'") || !strings.Contains(source,"sessionStorage.removeItem(suiteHandoffKey)") {t.Fatal("browser-only launch must discard stale handoff")}
 if strings.Contains(source,"suiteRenewButton.textContent=suiteSignedIn&&left>0?'Renew':'Open Homepage'") {t.Fatal("redundant Open Homepage action remains")}
 if !strings.Contains(source,"Not signed in · Browser-only settings") {t.Fatal("signed-out state feedback missing")}
 if !strings.Contains(source,".suite-home-link:hover{background:rgba(96,151,172,.16);color:var(--accent,#69c6d3);text-decoration:underline;opacity:1}") {t.Fatal("standardized hover border style missing")}

}
