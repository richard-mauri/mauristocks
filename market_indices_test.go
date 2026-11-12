package main

import (
 "fmt"
 "net/http"
 "net/http/httptest"
 "testing"
 "time"
)

func TestMarketTopRejectsInvalidRequests(t *testing.T) {
 for _,tc:=range []struct{method,path string;want int}{
  {"POST","/api/market-top?market=sp500",http.StatusMethodNotAllowed},
  {"GET","/api/market-top?market=unknown",http.StatusBadRequest},
  {"GET","/api/market-top",http.StatusBadRequest},
 } {
  w:=httptest.NewRecorder()
  topMarketHandler(w,httptest.NewRequest(tc.method,tc.path,nil))
  if w.Code!=tc.want{t.Errorf("%s %s: got %d, want %d",tc.method,tc.path,w.Code,tc.want)}
 }
}

func TestMarketCapsSnapshotRejectsInvalidAndDuplicateRows(t *testing.T){
 rows:=[][]string{{"symbol","name","price","marketCap","volume","industry"}}
 for i:=0;i<100;i++{
  rows=append(rows,[]string{fmt.Sprintf("S%d",i),"Stock","","10000","",""})
 }
 rows=append(rows,[]string{"NVDA","","","4210000000000","",""},
  []string{"NVDA","","","42","",""},
  []string{"BAD SYMBOL","","","2000","",""},
  []string{"NOPE","","","NaN","",""},
  []string{"ZERO","","","0","",""})
 caps,err:=parseMarketCaps(rows)
 if err!=nil{t.Fatal(err)}
 if caps["NVDA"]!=4210000000000{t.Fatalf("unexpected NVDA market cap %v",caps["NVDA"])}
 if _,ok:=caps["BAD SYMBOL"];ok{t.Fatal("invalid symbol accepted")}
 if _,ok:=caps["NOPE"];ok{t.Fatal("NaN market cap accepted")}
 if _,ok:=caps["ZERO"];ok{t.Fatal("zero market cap accepted")}
}
func TestIndexMembershipValidation(t *testing.T) {
 rows:=[][]string{{"Symbol","Name"}}
 for i:=0;i<30;i++{rows=append(rows,[]string{fmt.Sprintf("X%d",i),"Test"})}
 rows=append(rows,[]string{"X0","Duplicated"})
 members,err:=parseIndexSymbols(rows)
 if err!=nil||len(members)!=30{t.Fatalf("got %d members and %v",len(members),err)}
}
func TestTop30RankingAndDowPreservesUnweighted(t *testing.T){
 members:=make([]string,35)
 caps:=make(map[string]float64)
 for i:=range members{
  members[i]=fmt.Sprintf("SYM%d",i)
  caps[members[i]]=float64(i+1)*1e9
 }
 top,err:=rankMarketTop("sp500",members,caps,time.Now())
 if err!=nil{t.Fatal(err)}
 if len(top.Symbols)!=30||top.Symbols[0]!="SYM34"||top.Symbols[29]!="SYM5"{
  t.Fatalf("incorrect top-30 ranking: %v",top.Symbols)
 }
 dow:=members[:30]
 delete(caps,"SYM0")
 result,err:=rankMarketTop("dow30",dow,caps,time.Now())
 if err!=nil{t.Fatal(err)}
 if len(result.Symbols)!=30||len(result.Items)!=29||result.Symbols[29]!="SYM0"{
  t.Fatalf("Dow must preserve missing-cap constituent: %+v",result)
 }
}
