package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type SimpleServer struct {
	address string
	proxy   *httputil.ReverseProxy
}

type Server interface {
	Address() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func newSimpleServer(address string) *SimpleServer {
	serverUrl, err := url.Parse(address)
	handleErr(err)

	return &SimpleServer{
		address: address,
		proxy:   httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		roundRobinCount: 0,
		port:            port,
		servers:         servers,
	}
}

func (s *SimpleServer) Address() string {
	return s.address
}

func (s *SimpleServer) IsAlive() bool {
	return true
}

func (s *SimpleServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]

	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}

	lb.roundRobinCount++

	return server
}

func (lb *LoadBalancer) serveProxy(w http.ResponseWriter, r *http.Request) {
	targetServer := lb.getNextAvailableServer()

	fmt.Printf("forwarding request to %q\n", targetServer.Address())

	targetServer.Serve(w, r)
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	servers := []Server{
		newSimpleServer("https://www.youtube.com"),
		newSimpleServer("https://www.github.com"),
		newSimpleServer("https://www.facebook.com"),
	}

	lb := NewLoadBalancer("8080", servers)

	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.serveProxy(w, r)
	}

	http.HandleFunc("/", handleRedirect)

	fmt.Printf("Serve on port %s\n", lb.port)

	http.ListenAndServe(":"+lb.port, nil)
}
