package main

import (
 "context"
 "encoding/json"
 "errors"
 "io"
 "net/http"
 "regexp"
 "time"

 suiteauth "github.com/richard-mauri/homepage/shared/auth"
)

const stocksProfileEndpoint = "https://home.themauris.org/api/session/profile"
const stocksOrigin = "https://stocks.themauris.org"

var stocksBearerPattern = regexp.MustCompile(`^Bearer [a-f0-9]{64}$`)
var stocksProfileClient = &http.Client{Timeout:4*time.Second}

var errMissingStocksBearer = errors.New("Homepage bearer missing")

// Only Homepage validates bearer identity. A user-supplied role/name/header
// never substitutes for this server-to-server check.
func verifyStocksIdentity(ctx context.Context,r *http.Request)(suiteauth.Identity,error){
 authorization:=r.Header.Get("Authorization")
 if !stocksBearerPattern.MatchString(authorization){
  return suiteauth.Identity{},errMissingStocksBearer
 }
 request,err:=http.NewRequestWithContext(ctx,http.MethodGet,stocksProfileEndpoint,nil)
 if err!=nil{return suiteauth.Identity{},err}
 request.Header.Set("Authorization",authorization)
 request.Header.Set("Origin",stocksOrigin)
 request.Header.Set("Accept","application/json")
 response,err:=stocksProfileClient.Do(request)
 if err!=nil{return suiteauth.Identity{},err}
 defer response.Body.Close()
 if response.StatusCode==http.StatusUnauthorized || response.StatusCode==http.StatusForbidden{
  return suiteauth.Identity{},nil
 }
 if response.StatusCode!=http.StatusOK{return suiteauth.Identity{},suiteauth.ErrVerificationUnavailable}
 var profile struct{
  User string `json:"user"`
  Role suiteauth.Role `json:"role"`
  ExpiresAt string `json:"expiresAt"`
 }
 if err:=json.NewDecoder(io.LimitReader(response.Body,8192)).Decode(&profile);err!=nil{return suiteauth.Identity{},err}
 expiration,err:=time.Parse(time.RFC3339,profile.ExpiresAt)
 if err!=nil || !time.Now().Before(expiration){return suiteauth.Identity{},nil}
 identity:=suiteauth.Identity{User:profile.User,Role:profile.Role,Verified:true}
 if !suiteauth.Allowed(identity,suiteauth.UseApplication){return suiteauth.Identity{},nil}
 return identity,nil
}

// Guard application data APIs on every request, not only on initial page load.
// The HTML shell can bootstrap a fragment bearer; API responses require proof.
func guardStocksAPI(next http.HandlerFunc) http.HandlerFunc {
 wrapped:=suiteauth.Guard(func(ctx context.Context,r *http.Request)(suiteauth.Identity,error){
  identity,err:=verifyStocksIdentity(ctx,r)
  if errors.Is(err,errMissingStocksBearer){return suiteauth.Identity{},nil}
  return identity,err
 },suiteauth.UseApplication,next)
 return wrapped.ServeHTTP
}
