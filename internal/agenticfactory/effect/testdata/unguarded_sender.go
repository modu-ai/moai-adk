// Package sample is a positive control for the write-path scanner: it names a write method and never
// calls the guard. It lives in testdata, so it is never built.
package sample

func send() string { return "POST" }
