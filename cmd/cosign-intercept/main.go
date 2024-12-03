package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Start the proxy server in a goroutine.
	// proxyAddr := "localhost:8080"
	// go startProxyServer(proxyAddr)

	// // Set the environment variables for the proxy.
	// os.Setenv("HTTP_PROXY", "http://"+proxyAddr)
	// os.Setenv("HTTPS_PROXY", "http://"+proxyAddr)
	// os.Setenv("NO_PROXY", "localhost,127.0.0.1")
	os.Setenv("COSIGN_PASSWORD", "1234")

	go startReverseProxy()

	// Get Docker credentials
	username := "deep232"
	password := ""

	// Execute docker login command
	loginCmd := exec.Command("docker", "login", "-u", username, "-p", password)
	loginCmd.Stdout = os.Stdout

	loginCmd.Stderr = os.Stderr
	if err := loginCmd.Run(); err != nil {
		log.Fatalf("Docker login failed: %v", err)
	}

	// Build the cosign command arguments.
	cosignArgs := []string{
		"attest",
		"--y",
		"--key", "cosign.key",
		"--predicate", "package.json",
		"--rekor-url", "http://localhost:8080/v1/orgs/org1/projects/project1/",
		"deep232/plugins:latest",
	}

	// Create the cosign command.
	cmd := exec.Command("cosign", cosignArgs...)
	log.Printf("Running cosign command: cosign %s", strings.Join(cosignArgs, " "))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	// Run the cosign command.
	if err := cmd.Run(); err != nil {
		log.Fatalf("cosign command failed: %v", err)
	}
}

// startReverseProxy starts a reverse proxy server that adds the Authorization header.
func startReverseProxy() {
	// Define the actual Rekor server URL.
	rekorServerURL := "http://localhost:1234"

	target, err := url.Parse(rekorServerURL)
	if err != nil {
		log.Fatalf("Failed to parse Rekor server URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Inside startReverseProxy function
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
	// Modify the request to add the Authorization header.
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		// Remove the /v1/orgs/{org}/projects/{project} prefix from the URL path.
		// prefix := "/v1/orgs/org1/projects/project1"
		// if strings.HasPrefix(req.URL.Path, prefix) {
		// 	req.URL.Path = strings.TrimPrefix(req.URL.Path, prefix)
		// 	if req.URL.Path == "" {
		// 		req.URL.Path = "/"
		// 	}
		// }

		// Set the Authorization header.
		req.Header.Set("Authorization", "Bearer your_actual_token") // Replace with your actual token.
		// Log the modified request details
		log.Printf("Forwarding request to Rekor: %s %s", req.Method, req.URL.String())
		log.Printf("Request headers: %v", req.Header)

	}

	// Start the HTTP server.
	port := 8080
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxying request: %s %s", r.Method, r.URL)
		proxy.ServeHTTP(w, r)
	})
	log.Printf("Reverse proxy server is listening on port %d...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("Failed to start reverse proxy server: %v", err)
	}
}

// startProxyServer starts an HTTP proxy server that adds the Authorization header.
func startProxyServer(proxyAddr string) {
	server := &http.Server{
		Addr: proxyAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log.Printf("Received request: %s %s", req.Method, req.URL)
			if req.Method == http.MethodConnect {
				handleTunneling(w, req)
			} else {
				handleHTTP(w, req)
			}
		}),
	}

	log.Printf("Starting proxy server on %s", proxyAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Proxy server failed: %v", err)
	}
}

// handleTunneling handles HTTPS connections via the CONNECT method.
func handleTunneling(w http.ResponseWriter, req *http.Request) {
	destConn, err := net.Dial("tcp", req.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

// handleHTTP handles HTTP requests, adds the Authorization header, and forwards them.
func handleHTTP(w http.ResponseWriter, req *http.Request) {
	// Remove hop-by-hop headers.
	req.RequestURI = ""
	req.Header.Del("Proxy-Connection")
	req.Header.Del("Connection")
	req.Header.Del("Keep-Alive")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Te")
	req.Header.Del("Trailers")
	req.Header.Del("Transfer-Encoding")
	req.Header.Del("Upgrade")

	// Add the Authorization header.
	req.Header.Set("Authorization", "Bearer your_token_here") // Replace with your actual token.

	// Create a transport and forward the request.
	transport := http.DefaultTransport
	resp, err := transport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy the response headers and status code.
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	// Copy the response body.
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}

// transfer copies data between source and destination connections.
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	if _, err := io.Copy(destination, source); err != nil {
		log.Printf("Error during transfer: %v", err)
	}
}

// copyHeader copies headers from src to dst.
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
