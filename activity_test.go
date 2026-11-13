package main

import (
 "context"
 "encoding/json"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"
)

func TestStocksActivityReporterConfig(t *testing.T) {
 if got,err:=NewActivityReporter("","");err!=nil||got!=nil{t.Fatalf("disabled reporter: %v",err)}
 for _,endpoint:=range []string{"http://example.org/api/admin/activity/ingest","http://localhost:8788/other","https://u:p@example.org/api/admin/activity/ingest"}{
  if _,err:=NewActivityReporter(endpoint,strings.Repeat("k",40));err==nil{t.Errorf("accepted unsafe endpoint: %s",endpoint)}
 }
 if _,err:=NewActivityReporter("http://127.0.0.1:8788/api/admin/activity/ingest","short");err==nil{t.Error("accepted short key")}
 if got,err:=NewActivityReporter("http://127.0.0.1:8788/api/admin/activity/ingest",strings.Repeat("k",40));err!=nil||got==nil{t.Fatalf("valid loopback setup failed: %v",err)}
}

func TestStocksActivityPresenceAndInFlight(t *testing.T){
 a,_:=NewActivityReporter("http://127.0.0.1:8788/api/admin/activity/ingest",strings.Repeat("k",40))
 now:=time.Now().UTC().Truncate(time.Second)
 a.now=func()time.Time{return now}
 send:=func(origin,id string)int {
  req:=httptest.NewRequest("POST","http://127.0.0.1:8086/api/activity/ping",strings.NewReader(`{"id":"`+id+`","visible":true,"interacted":true}`))
  req.Header.Set("Origin",origin)
  req.Header.Set("Content-Type","application/json")
  result:=httptest.NewRecorder();a.ping(result,req);return result.Code
 }
 id:=strings.Repeat("b",32)
 if got:=send("https://evil.example",id);got!=403{t.Errorf("cross origin: %d",got)}
 if got:=send("http://127.0.0.1:8086",id);got!=204{t.Fatalf("ping: %d",got)}
 p:=a.snapshot()
 if p.Active2!=1||p.Active60!=1||p.Foreground!=1||p.Anonymous!=1||p.Authenticated!=0{t.Fatalf("presence: %+v",p)}
 finish:=a.working()
 if p=a.snapshot();p.InFlight!=1{t.Fatalf("in-flight request absent: %+v",p)}
 finish()
 if p=a.snapshot();p.InFlight!=0{t.Fatalf("work counter stuck: %+v",p)}
 now=now.Add(3*time.Minute)
 if p=a.snapshot();p.Active2!=0||p.Active5!=1{t.Fatalf("window rollover: %+v",p)}
}

func TestStocksReporterAggregateWireContract(t *testing.T){
 key:=strings.Repeat("x",40)
 var report activityPayload
 endpoint:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.Header.Get("Authorization")!="Bearer "+key||r.Header.Get("X-Activity-Service")!="stocks"{t.Error("missing server credentials or service")}
  if r.Header.Get("X-Activity-Timestamp")==""||len(r.Header.Get("X-Activity-Nonce"))<16{t.Error("timestamp/nonce missing")}
  if err:=json.NewDecoder(r.Body).Decode(&report);err!=nil{t.Error(err)}
  w.WriteHeader(204)
 }))
 defer endpoint.Close()
 a,err:=NewActivityReporter(endpoint.URL+"/api/admin/activity/ingest",key)
 if err!=nil{t.Fatal(err)}
 a.mu.Lock()
 a.presence[strings.Repeat("c",32)]=activityPresence{seen:time.Now(),visible:true}
 a.mu.Unlock()
 if err:=a.report(context.Background());err!=nil{t.Fatal(err)}
 if report.Active2!=1||report.Anonymous!=1||report.Active60!=1{t.Fatalf("unexpected report: %+v",report)}
}

func TestStockActivityScriptPrivateAndOptional(t *testing.T){
 if !strings.Contains(activityJS,"/api/activity/ping")||!strings.Contains(activityJS,"sessionStorage"){t.Fatal("missing browser heartbeat")}
 for _,secret:=range []string{"HOMEPAGE_ACTIVITY_KEY_STOCKS","STOCKS_ACTIVITY_URL"}{
  if strings.Contains(activityJS,secret){t.Errorf("browser script contains private config %s",secret)}
 }
 if !strings.Contains(page,"</body>"){t.Fatal("page template anchor missing")}
}

func TestStocksVerifiedSessionsAndRevocation(t *testing.T) {
 valid:=true
 homepage:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/api/session/profile" || r.Header.Get("Origin")!="https://stocks.themauris.org" {t.Errorf("unexpected profile request %s",r.URL.Path);http.Error(w,"bad origin",403);return}
  if !valid || r.Header.Get("Authorization")!="Bearer "+strings.Repeat("a",64) {http.Error(w,"revoked",401);return}
  w.Header().Set("Content-Type","application/json")
  _=json.NewEncoder(w).Encode(map[string]string{"user":"guest-example","activitySession":strings.Repeat("b",64),"activityUser":strings.Repeat("c",64)})
 }))
 defer homepage.Close()
 a,err:=NewActivityReporter(homepage.URL+"/api/admin/activity/ingest",strings.Repeat("k",40))
 if err!=nil{t.Fatal(err)}
 for _,id:=range []string{strings.Repeat("d",32),strings.Repeat("e",32)} {
  req:=httptest.NewRequest(http.MethodPost,"http://localhost:8082/api/activity/ping",strings.NewReader(`{"id":"`+id+`","visible":true}`))
  req.Header.Set("Origin","http://localhost:8082")
  req.Header.Set("Content-Type","application/json")
  req.Header.Set("Authorization","Bearer "+strings.Repeat("a",64))
  result:=httptest.NewRecorder();a.ping(result,req)
  if result.Code!=204 {t.Fatalf("verified ping status %d",result.Code)}
 }
 aggregate:=a.snapshot()
 if aggregate.Active2!=1 || aggregate.Authenticated!=1 || aggregate.UniqueUsers!=1 || aggregate.Anonymous!=0 || len(aggregate.Users)!=1 || aggregate.Users[0].Sessions!=1 {t.Fatalf("session deduplication: %+v",aggregate)}
 valid=false
 req:=httptest.NewRequest(http.MethodPost,"http://localhost:8082/api/activity/ping",strings.NewReader(`{"id":"`+strings.Repeat("d",32)+`","visible":true}`))
 req.Header.Set("Origin","http://localhost:8082");req.Header.Set("Content-Type","application/json")
 req.Header.Set("Authorization","Bearer "+strings.Repeat("a",64))
 result:=httptest.NewRecorder();a.ping(result,req)
 if result.Code!=401 {t.Fatalf("revoked session status %d",result.Code)}
 if strings.Contains(activityJS,"HOMEPAGE_ACTIVITY_KEY_STOCKS") {t.Fatal("collector credential exposed")}
 for _,part:=range []string{"Suite sign-in required","Go to Homepage","event.preventDefault()","response.status!==401"} {
  if !strings.Contains(activityJS,part) {t.Errorf("missing browser session contract %q",part)}
 }
}
