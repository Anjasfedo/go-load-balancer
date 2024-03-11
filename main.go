package main

import (
	"net/http/httputil"
)

type simpleServer struct {
	address string
	proxy   *httputil.ReverseProxy
}

func main() {

}
