package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type EndpointLoad struct {
	Name     string
	Method   string
	URL      string
	QPS      uint64
	Duration time.Duration
	Body     []byte
}

// loginResponse represents the response body of /api/v1/login
type loginResponse struct {
	Message string `json:"message"` // Add "message" field
	Token   string `json:"token"`   // Keep existing "token" field
}

// loginAndGetTokens is a function that logs in for the specified user range and obtains JWT tokens
func loginAndGetTokens(baseURL string, password string, numUsers int) ([]string, error) {
	tokens := make([]string, 0, numUsers)
	var wg sync.WaitGroup
	var mu sync.Mutex                                 // Protect addition to tokens slice
	client := &http.Client{Timeout: 10 * time.Second} // Set timeout

	fmt.Printf("Starting token acquisition for %d users...\n", numUsers)

	for i := 1; i <= numUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			username := fmt.Sprintf("sample%d@example.com", userID)
			loginURL := baseURL + "/auth/login"
			// Create request body as byte array
			loginBodyBytes, _ := json.Marshal(map[string]string{
				"email":    username,
				"password": password,
			})

			req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(loginBodyBytes))
			if err != nil {
				log.Printf("Error creating login request for user %s: %v", username, err)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				// In case of network errors etc.
				log.Printf("Error sending login request for user %s: %v", username, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Login failed for user %s: Status code %d", username, resp.StatusCode)
				return
			}

			var res loginResponse
			if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
				log.Printf("Error parsing login response for user %s: %v", username, err)
				return
			}

			if res.Token == "" {
				log.Printf("Token is empty for user %s", username)
				return
			}

			mu.Lock()
			tokens = append(tokens, res.Token)
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	fmt.Println()

	if len(tokens) == 0 {
		return nil, fmt.Errorf("Could not obtain any tokens. Please check login endpoint and authentication credentials")
	}
	if len(tokens) < numUsers {
		log.Printf("Warning: Only %d tokens were obtained out of %d users.", len(tokens), numUsers)
	} else {
		fmt.Printf("Successfully obtained %d tokens.\n", len(tokens))
	}
	return tokens, nil
}

func main() {
	totalQPS := flag.Uint64("qps", 300, "Total QPS (default: 300)")
	flag.Parse()

	// --- Settings ---
	baseURL := "http://localhost:8081" // API base URL
	loginPassword := "sample_password" // Login password
	numLoginUsers := 20                // Number of users to attempt login
	// --- Settings end ---

	// Initialize random number generator
	rand.Seed(time.Now().UnixNano())

	// --- Token acquisition process ---
	tokens, err := loginAndGetTokens(baseURL, loginPassword, numLoginUsers)
	if err != nil {
		log.Fatalf("Failed to obtain tokens: %v", err)
	}
	// --- Token acquisition process end ---

	// Endpoint ratios
	ratios := map[string]float64{
		"Purchase":   0.7,
		"GetAllItem": 0.19,
		"GetItem":    0.1,
		"Login":      0.08,
		"Register":   0.02,
	}

	// QPS allocation
	qpsMap := map[string]uint64{}
	for name, ratio := range ratios {
		qpsMap[name] = uint64(math.Round(float64(*totalQPS) * ratio))
	}

	endpoints := []EndpointLoad{
		{
			Name:     "Purchase",
			Method:   "POST",
			URL:      baseURL + "/purchases",
			QPS:      qpsMap["Purchase"],
			Duration: 30 * time.Second,
			Body:     nil,
		},
		{
			Name:     "GetAllItem",
			Method:   "GET",
			URL:      baseURL + "/items",
			QPS:      qpsMap["GetAllItem"],
			Duration: 30 * time.Second,
		},
		{
			Name:     "Register",
			Method:   "POST",
			URL:      baseURL + "/register",
			QPS:      qpsMap["Register"],
			Duration: 30 * time.Second,
			Body:     []byte(`{"username":"testuser","password":"testpass"}`),
		},
		{
			Name:     "Login",
			Method:   "POST",
			URL:      baseURL + "/login",
			QPS:      qpsMap["Login"],
			Duration: 30 * time.Second,
			Body:     []byte(`{"username":"testuser","password":"testpass"}`),
		},
		{
			Name:     "GetItem",
			Method:   "GET",
			URL:      baseURL + "/items",
			QPS:      qpsMap["GetItem"],
			Duration: 30 * time.Second,
			Body:     nil,
		},
		{
			Name:     "GetItems",
			Method:   "GET",
			URL:      baseURL + "/items",
			QPS:      0, // If not used, set to 0
			Duration: 30 * time.Second,
		},
	}

	var wg sync.WaitGroup
	var purchaseTokenIndex uint64 // Purchase request token index (for atomic operations)

	// Overall aggregation
	var totalRequests uint64
	var totalDuration time.Duration
	var muTotal sync.Mutex

	for _, ep := range endpoints {
		wg.Add(1)
		// Copy token list to pass to goroutine (thread-safe)
		currentTokens := make([]string, len(tokens))
		copy(currentTokens, tokens)

		go func(ep EndpointLoad, tokensForGoroutine []string) {
			defer wg.Done()
			rate := vegeta.Rate{Freq: int(ep.QPS), Per: time.Second}

			var targeter vegeta.Targeter // Declare Targeter

			// Set dynamic Targeter for Purchase endpoint
			if ep.Name == "Purchase" && len(tokensForGoroutine) > 0 {
				targeter = func(tgt *vegeta.Target) error {
					if tgt == nil {
						return vegeta.ErrNilTarget
					}
					// Increment atomically and cycle through token list
					idx := atomic.AddUint64(&purchaseTokenIndex, 1)
					token := tokensForGoroutine[idx%uint64(len(tokensForGoroutine))] // Round-robin

					// Generate random item ID from 1 to 30
					itemID := rand.Intn(30) + 1

					// Generate dynamic body
					bodyMap := map[string]interface{}{
						"item_id":  itemID, // Use random itemID
						"Quantity": 1,      // Send "Quantity" in uppercase
					}
					bodyBytes, err := json.Marshal(bodyMap)
					if err != nil {
						// JSON generation error is usually rare, but log for safety
						log.Printf("Error marshaling purchase body: %v", err)
						return fmt.Errorf("failed to marshal request body: %w", err)
					}

					// Set original target information
					tgt.Method = ep.Method
					tgt.URL = ep.URL     // /purchase remains
					tgt.Body = bodyBytes // Set dynamic body
					// Generate headers each time
					tgt.Header = http.Header{}
					tgt.Header.Set("Content-Type", "application/json")
					tgt.Header.Set("Authorization", "Bearer "+token) // Set obtained token
					return nil
				}
			} else if ep.Name == "GetItem" { // Set dynamic Targeter for GetItem endpoint
				targeter = func(tgt *vegeta.Target) error {
					if tgt == nil {
						return vegeta.ErrNilTarget
					}
					// Generate random item ID from 1 to 30
					itemID := rand.Intn(30) + 1
					// Generate dynamic URL
					targetURL := fmt.Sprintf("%s/%d", ep.URL, itemID)

					// Set original target information
					tgt.Method = ep.Method
					tgt.URL = targetURL // Set dynamic URL
					tgt.Body = nil      // GET request, so body is nil
					tgt.Header = nil    // GET request, so headers are nil (if needed, set)
					return nil
				}
			} else { // Use static Targeter for other endpoints
				target := vegeta.Target{
					Method: ep.Method,
					URL:    ep.URL,
					Body:   ep.Body,
				}
				header := http.Header{}
				// Set Content-Type for POST/PUT requests
				if ep.Method == "POST" || ep.Method == "PUT" {
					header.Set("Content-Type", "application/json")
				}
				target.Header = header
				targeter = vegeta.NewStaticTargeter(target) // Static Targeter
			}

			attacker := vegeta.NewAttacker()
			var metrics vegeta.Metrics
			// Attack execution
			for res := range attacker.Attack(targeter, rate, ep.Duration, ep.Name) {
				metrics.Add(res)
			}
			metrics.Close()

			// Display results
			fmt.Printf("【%s】\n", ep.Name)
			fmt.Printf("  Requests: %d\n", metrics.Requests)
			fmt.Printf("  Success rate: %.2f%%\n", metrics.Success*100)
			fmt.Printf("  Average latency: %s\n", metrics.Latencies.Mean)
			fmt.Printf("  P95 latency: %s\n", metrics.Latencies.P95)
			fmt.Printf("  Error count: %d\n", len(metrics.Errors))
			fmt.Printf("  Actual average QPS: %.2f\n", float64(metrics.Requests)/metrics.Duration.Seconds())
			if len(metrics.Errors) > 0 {
				fmt.Println("  Error details:")
				for _, errStr := range metrics.Errors {
					fmt.Printf("    - %s\n", errStr)
				}
			}
			fmt.Println("------")

			// Overall aggregation
			muTotal.Lock()
			totalRequests += metrics.Requests
			if metrics.Duration > totalDuration {
				totalDuration = metrics.Duration // Use the longest one as the overall Duration
			}
			muTotal.Unlock()
		}(ep, currentTokens) // Pass token list
	}
	wg.Wait()
	// Output total QPS
	fmt.Println("====== Overall Aggregation ======")
	fmt.Printf("Total requests across all endpoints: %d\n", totalRequests)
	fmt.Printf("Overall test duration (longest): %s\n", totalDuration)
	if totalDuration > 0 {
		fmt.Printf("Overall QPS: %.2f\n", float64(totalRequests)/totalDuration.Seconds())
	}
	fmt.Println("=====================")
	fmt.Println("All load tests have completed.")
}
