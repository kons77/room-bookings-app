package main

import (
	"net/http"
	"testing"
)

/*
I need to create a handler to pass to no surf so it can hand me back a handler.
*/
func TestNoSurf(t *testing.T) {
	// create a variable that satisfies the interface for Http handler
	var myH myHandler

	h := NoSurf(&myH)

	switch v := h.(type) {
	case http.Handler:
		/// do nothing
	default:
		t.Errorf("type is not http.Handler, but is %T", v)
	}
}

func TestSessionLoad(t *testing.T) {
	// create a variable that satisfies the interface for Http handler
	var myH myHandler

	h := SessionLoad(&myH)

	switch v := h.(type) {
	case http.Handler:
		/// do nothing
	default:
		t.Errorf("type is not http.Handler, but is %T", v)
	}
}
