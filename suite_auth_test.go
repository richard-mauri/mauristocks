package main

import (
 "context"
 "errors"
 "fmt"
 "io"
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"
 "time"

 suiteauth "github.com/richard-mauri/homepage/shared/auth"
)

type mockProfileTransport func(*http.Request)(*http.Response,error)
func (f mockProfileTransport)RoundTrip(r *http.Request)(*http.Response,error){return f(r)}

func TestStocksSharedAuthGuards(t *testing.T){
 original:=stocksProfileClient
 defer func(){stocksProfileClient=original}()
 token:=strings.Repeat("a",64)
 for _,tc:=range []struct{name string;token bool;status int;role,user string;outage bool;want int}{
  {"anonymous",false,0,"","",false,401},
  {"user",true,200,"user","reader",false,200},
  {"admin",true,200,"admin","operator",false,200},
  {"guest",true,200,"guest","guest-123",false,200},
  {"revoked",true,401,"","",false,401},
  {"forbidden",true,403,"","",false,401},
  {"outage",true,503,"","",false,503},
  {"network down",true,0,"","",true,503},
  {"unknown role",true,200,"root","root",false,401},
  {"missing username",true,200,"user","",false,401},
 }{
  t.Run(tc.name,func(t *testing.T){
   calls:=0
   stocksProfileClient=&http.Client{Transport:mockProfileTransport(func(r *http.Request)(*http.Response,error){
    calls++
    if r.URL.String()!=stocksProfileEndpoint || r.Header.Get("Origin")!=stocksOrigin ||
       r.Header.Get("Authorization")!="Bearer "+token{
      t.Fatal("server verifier must use only fixed Homepage profile endpoint and supplied bearer")
    }
    if tc.outage{return nil,errors.New("offline")}
    expires:=time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
    body:=fmt.Sprintf(`{"user":%q,"role":%q,"expiresAt":%q}`,tc.user,tc.role,expires)
    return &http.Response{StatusCode:tc.status,Body:io.NopCloser(strings.NewReader(body)),Header:make(http.Header)},nil
   })}
   hit:=false
   app:=guardStocksAPI(func(w http.ResponseWriter,r *http.Request){hit=true;w.WriteHeader(200)})
   request:=httptest.NewRequest("GET","http://stocks.themauris.org/api/quotes?symbols=MU",nil)
   request.Header.Set("X-Suite-Role","admin") // untrusted browser label
   if tc.token {request.Header.Set("Authorization","Bearer "+token)}
   rec:=httptest.NewRecorder()
   app(rec,request)
   if rec.Code!=tc.want{t.Fatalf("status %d want %d",rec.Code,tc.want)}
   if hit!=(tc.want==200){t.Fatal("guard allowed unexpected downstream access")}
   if !tc.token && calls!=0{t.Fatal("anonymous request contacted Homepage unnecessarily")}
   if tc.token && calls!=1{t.Fatalf("expected one verification, got %d",calls)}
  })
 }
}

func TestStocksMalformedAndExpiredBearer(t *testing.T){
 original:=stocksProfileClient
 defer func(){stocksProfileClient=original}()
 token:=strings.Repeat("c",64)
 for _,tc:=range []struct{name,auth,expires string;want int}{
  {"invalid token","Bearer garbage","",401},
  {"missing bearer","","",401},
  {"expired valid bearer","Bearer "+token,time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),401},
 }{
  t.Run(tc.name,func(t *testing.T){
   stocksProfileClient=&http.Client{Transport:mockProfileTransport(func(r *http.Request)(*http.Response,error){
    return &http.Response{StatusCode:200,Body:io.NopCloser(strings.NewReader(fmt.Sprintf(`{"user":"reader","role":"user","expiresAt":%q}`,tc.expires))),Header:make(http.Header)},nil
   })}
   req:=httptest.NewRequest("GET","http://stocks.themauris.org/api/quotes",nil)
   if tc.auth!=""{req.Header.Set("Authorization",tc.auth)}
   identity,err:=verifyStocksIdentity(context.Background(),req)
   if tc.auth!=""&&tc.auth!="Bearer garbage"&&err!=nil{t.Fatal(err)}
   if identity.Verified{t.Fatal("expired or malformed bearer accepted")}
   rec:=httptest.NewRecorder()
   guardStocksAPI(func(http.ResponseWriter,*http.Request){t.Fatal("unauthorized call reached handler")})(rec,req)
   if rec.Code!=tc.want{t.Fatalf("got %d want %d",rec.Code,tc.want)}
  })
 }
}

func TestStocksSharedPermissions(t *testing.T){
 personal:=suiteauth.Identity{User:"reader",Role:suiteauth.RoleUser,Verified:true}
 guest:=suiteauth.Identity{User:"guest",Role:suiteauth.RoleGuest,Verified:true}
 if !suiteauth.Allowed(personal,suiteauth.UsePersonalSettings){t.Fatal("personal settings disallowed")}
 if suiteauth.Allowed(guest,suiteauth.UsePersonalSettings){t.Fatal("guest gained account persistence")}
}
