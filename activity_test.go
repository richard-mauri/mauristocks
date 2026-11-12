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
