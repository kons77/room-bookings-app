package main

import (
	"net/http"
	"os"
	"testing"
)

/*
it also gives me a place to store variables
that I might need outside of the test main function.
*/
func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

type myHandler struct{}

func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}