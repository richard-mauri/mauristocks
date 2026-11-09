package main

import (
 "os/exec"
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
