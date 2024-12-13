package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"golang.org/x/exp/rand"
)

func main() {
	var wg sync.WaitGroup
	os.Setenv("COSIGN_PASSWORD", "1234")

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
	// Start the reverse proxy in a separate goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		startReverseProxy()
	}()

	wg.Wait()
	// Build the cosign command arguments.
	// Create the cosign command.
	// Run the cosign command.
	//testCosignWithReverseProxy()
}

func testCosignWithReverseProxy() {
	cosignArgs := []string{
		"attest",
		"--y",
		"--key", "cosign.key",
		"--predicate", "package.json",
		"--rekor-url", "http://localhost:8080/v1/orgs/org1/projects/project1/",
		"deep232/plugins:latest",
	}

	cmd := exec.Command("cosign", cosignArgs...)
	log.Printf("Running cosign command: cosign %s", strings.Join(cosignArgs, " "))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

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

		// Array of possible harness-account values
		harnessAccounts := []string{
			"wFHXHD0RRQWoO8tIZT5YVw",
			"XJDKA76aDhGqLZB2N4PxYQ",
			"ZPQLXM9RtJFDWqBYKDHE5A",
			"MTGQ8NABPWoLJZKCWD6FYX",
		}

		// Seed the random number generator
		rand.Seed(rand.Uint64())
		randomAccount := harnessAccounts[rand.Intn(len(harnessAccounts))]

		// Set the Authorization header.
		req.Header.Set("Authorization", "Bearer your_actual_token") // Replace with your actual token.
		//TODO no need to set this. this is for testing
		req.Header.Set("harness-account", randomAccount)
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
