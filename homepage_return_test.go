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
}
