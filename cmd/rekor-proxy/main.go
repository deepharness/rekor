package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// possible rekorInstanceMap maps organization/project combinations to Rekor instance URLs.
var rekorInstanceMap = map[string]string{
	"org1/project1": "http://localhost:3001",
	"org2/project2": "http://localhost:3002",
}

// ProxyHandler handles incoming requests and proxies them to the appropriate Rekor server.
func ProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received headers: %v", r.Header)
		// Extract the auth header.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Parse the URL path to extract org and project.
		pathSegments := strings.Split(r.URL.Path, "/")
		if len(pathSegments) < 6 || pathSegments[1] != "v1" || pathSegments[2] != "orgs" || pathSegments[4] != "projects" {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		org := pathSegments[3]
		project := pathSegments[5]
		// key := fmt.Sprintf("%s/%s", org, project)

		// // Retrieve the corresponding Rekor instance URL.
		// rekorURLStr, exists := rekorInstanceMap[key]
		// if !exists {
		// 	http.Error(w, "Organization or project not found", http.StatusNotFound)
		// 	return
		// }

		rekorURLStr := "http://localhost:3000"

		// Parse the Rekor server URL.
		rekorURL, err := url.Parse(rekorURLStr)
		if err != nil {
			log.Printf("Failed to parse Rekor URL: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Remove the /v1/orgs/{org}/projects/{project} prefix from the URL path.
		newPath := "/" + strings.Join(pathSegments[6:], "/")
		r.URL.Path = newPath

		// Create a reverse proxy to the Rekor server.
		proxy := httputil.NewSingleHostReverseProxy(rekorURL)

		// Modify the request before sending it to the Rekor server.
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			// Call the original director to set up the request.
			originalDirector(req)

			// Update the request URL to point to the Rekor instance.
			req.URL.Scheme = rekorURL.Scheme
			req.URL.Host = rekorURL.Host
			req.URL.Path = newPath

			// Optionally, forward the auth header.
			// req.Header.Set("Authorization", authHeader)
		}

		// Log the request for debugging purposes.
		log.Printf("Proxying request for org: %s, project: %s, path: %s to %s", org, project, r.URL.Path, rekorURLStr)

		// Proxy the request to the Rekor server.
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	// Set up the HTTP server.
	http.HandleFunc("/v1/orgs/", ProxyHandler())

	// Start the HTTP server.
	port := 1234
	log.Printf("Proxy server is listening on port %d...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}
}
