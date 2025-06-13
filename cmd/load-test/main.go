// File: j-ticketing/cmd/load-test/main.go
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// CreateFreeOrderRequest matches the structure from your project
type CreateFreeOrderRequest struct {
	TicketGroupId    uint            `json:"ticketGroupId"`
	IdentificationNo string          `json:"identificationNo"`
	FullName         string          `json:"fullName"`
	Email            string          `json:"email"`
	ContactNo        string          `json:"contactNo"`
	Date             string          `json:"date"`
	LangChosen       string          `json:"langChosen"`
	Tickets          []TicketRequest `json:"tickets"`
	AllowBypass      bool            `json:"AllowBypass"`
}

type TicketRequest struct {
	TicketId string `json:"ticketId"`
	Qty      int    `json:"qty"`
}

// TicketVariantResponse structures based on your project
type TicketVariantResponse struct {
	RespCode int                 `json:"respCode"`
	RespDesc string              `json:"respDesc"`
	Result   TicketVariantResult `json:"result"`
}

type TicketVariantResult struct {
	TicketVariants []TicketVariantDTO `json:"ticketVariants"`
}

type TicketVariantDTO struct {
	TicketVariantId *uint   `json:"ticketVariantId"`
	TicketGroupId   *uint   `json:"ticketGroupId"`
	TicketId        *string `json:"ticketId"`
	NameBm          string  `json:"nameBm"`
	NameEn          string  `json:"nameEn"`
	NameCn          string  `json:"nameCn"`
	DescBm          string  `json:"descBm"`
	DescEn          string  `json:"descEn"`
	DescCn          string  `json:"descCn"`
	UnitPrice       float64 `json:"unitPrice"`
	PrintType       string  `json:"printType"`
}

// Response structure to capture results
type TestResult struct {
	UserID       int
	StatusCode   int
	ResponseBody string
	Duration     time.Duration
	Error        error
}

// Configuration for the load test
type LoadTestConfig struct {
	BaseURL           string
	NumUsers          int
	TicketGroupId     uint
	RequestsPerSecond int
	Emails            []string
}

// Global variable to store fetched ticket variants
var availableTickets []TicketVariantDTO
var ticketsFetched bool
var ticketsMutex sync.RWMutex

func main() {
	// Define command-line flags
	var (
		baseURL           = flag.String("url", "http://localhost:8080", "Base URL of the server")
		numUsers          = flag.Int("users", 5, "Number of concurrent users")
		ticketGroupId     = flag.Uint("ticketGroupId", 1, "Ticket group ID")
		requestsPerSecond = flag.Int("rps", 10, "Requests per second rate limit")
		emailsFlag        = flag.String("emails", "test@gmail.com", "Comma-separated list of emails")
		showHelp          = flag.Bool("help", false, "Show help message")
	)

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Load Test Tool for Ticket Ordering System\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -url=http://localhost:8080 -users=10 -group=2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -users=50 -rps=20 -emails=test1@example.com,test2@example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -help\n", os.Args[0])
	}

	flag.Parse()

	// Show help if requested
	if *showHelp {
		flag.Usage()
		return
	}

	// Parse emails from comma-separated string
	emails := strings.Split(strings.TrimSpace(*emailsFlag), ",")
	for i, email := range emails {
		emails[i] = strings.TrimSpace(email)
	}

	// Validate inputs
	if *numUsers <= 0 {
		fmt.Fprintf(os.Stderr, "Error: Number of users must be greater than 0\n")
		os.Exit(1)
	}

	if *requestsPerSecond <= 0 {
		fmt.Fprintf(os.Stderr, "Error: Requests per second must be greater than 0\n")
		os.Exit(1)
	}

	if len(emails) == 0 || (len(emails) == 1 && emails[0] == "") {
		fmt.Fprintf(os.Stderr, "Error: At least one email must be provided\n")
		os.Exit(1)
	}

	// Create configuration from command line arguments
	config := LoadTestConfig{
		BaseURL:           *baseURL,
		NumUsers:          *numUsers,
		TicketGroupId:     *ticketGroupId,
		RequestsPerSecond: *requestsPerSecond,
		Emails:            emails,
	}

	// Display configuration
	fmt.Println("=================================================")
	fmt.Println("LOAD TEST CONFIGURATION")
	fmt.Println("=================================================")
	fmt.Printf("Base URL: %s\n", config.BaseURL)
	fmt.Printf("Concurrent Users: %d\n", config.NumUsers)
	fmt.Printf("Ticket Group ID: %d\n", config.TicketGroupId)
	fmt.Printf("Rate Limit: %d requests/second\n", config.RequestsPerSecond)
	fmt.Printf("Test Emails: %v\n", config.Emails)
	fmt.Printf("Target Endpoint: %s/api/orderTicketGroup/free\n", config.BaseURL)
	fmt.Println("=================================================")

	// First, fetch available tickets for the ticket group
	fmt.Println("Fetching available tickets for ticket group...")
	err := fetchAvailableTickets(config)
	if err != nil {
		fmt.Printf("Error fetching ticket variants: %v\n", err)
		fmt.Println("Proceeding with fallback ticket configuration...")
		// Continue with default ticket if API call fails
	} else {
		fmt.Printf("Successfully fetched %d ticket variants\n", len(availableTickets))
		for i, ticket := range availableTickets {
			ticketId := "N/A"
			if ticket.TicketId != nil {
				ticketId = *ticket.TicketId
			}
			fmt.Printf("  %d. %s (ID: %s, Price: %.2f)\n", i+1, ticket.NameEn, ticketId, ticket.UnitPrice)
		}
	}

	fmt.Println("=================================================")
	fmt.Println("STARTING LOAD TEST...")
	fmt.Println("=================================================")

	results := runLoadTest(config)

	// Analyze and display results
	analyzeResults(results)
}

func fetchAvailableTickets(config LoadTestConfig) error {
	// Get today's date in Malaysian timezone
	malaysiaLoc, err := time.LoadLocation("Asia/Kuala_Lumpur")
	if err != nil {
		return fmt.Errorf("failed to load Malaysia timezone: %w", err)
	}

	todayDate := time.Now().In(malaysiaLoc).Format("2006-01-02")

	// Create HTTP request with query parameters - GET method
	url := fmt.Sprintf("%s/api/ticketGroups/ticketVariants?ticketGroupId=%d&date=%s",
		config.BaseURL, config.TicketGroupId, todayDate)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, responseBody.String())
	}

	// Parse the response
	var variantResponse TicketVariantResponse
	err = json.Unmarshal(responseBody.Bytes(), &variantResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check if the response indicates success
	if variantResponse.RespCode != 200 && variantResponse.RespCode != 2000 {
		return fmt.Errorf("API returned error code %d: %s", variantResponse.RespCode, variantResponse.RespDesc)
	}

	// Store the ticket variants
	ticketsMutex.Lock()
	availableTickets = variantResponse.Result.TicketVariants
	ticketsFetched = true
	ticketsMutex.Unlock()

	return nil
}

func runLoadTest(config LoadTestConfig) []TestResult {
	var wg sync.WaitGroup
	results := make([]TestResult, config.NumUsers)

	// Channel for rate limiting
	rateLimiter := time.NewTicker(time.Second / time.Duration(config.RequestsPerSecond))
	defer rateLimiter.Stop()

	startTime := time.Now()

	for i := 0; i < config.NumUsers; i++ {
		wg.Add(1)

		// Rate limiting - wait for ticker
		<-rateLimiter.C

		go func(userID int) {
			defer wg.Done()

			result := TestResult{UserID: userID}
			requestStart := time.Now()

			// Generate test data for this user
			request := generateTestRequest(config, userID)

			// Make the HTTP request
			statusCode, responseBody, err := makeRequest(config, request)

			result.StatusCode = statusCode
			result.ResponseBody = responseBody
			result.Duration = time.Since(requestStart)
			result.Error = err

			results[userID] = result

			// Log individual results
			if err != nil {
				fmt.Printf("User %d: ERROR - %v (Duration: %v)\n", userID, err, result.Duration)
			} else {
				fmt.Printf("User %d: Status %d (Duration: %v)\n", userID, statusCode, result.Duration)
			}
		}(i)
	}

	wg.Wait()

	totalDuration := time.Since(startTime)
	fmt.Printf("\nAll requests completed in %v\n", totalDuration)

	return results
}

func generateTestRequest(config LoadTestConfig, userID int) CreateFreeOrderRequest {
	// Generate unique test data for each user
	languages := []string{"bm", "en", "cn"}

	// Generate tickets based on available variants or fallback
	tickets := generateRandomTickets()

	// Use today's date in Malaysian timezone for the order
	malaysiaLoc, _ := time.LoadLocation("Asia/Kuala_Lumpur")
	todayDate := time.Now().In(malaysiaLoc).Format("2006-01-02")

	return CreateFreeOrderRequest{
		TicketGroupId:    config.TicketGroupId,
		IdentificationNo: fmt.Sprintf("ID%06d", 100000+userID),
		FullName:         fmt.Sprintf("Test User %d", userID),
		Email:            config.Emails[rand.Intn(len(config.Emails))], // Randomly select email
		ContactNo:        fmt.Sprintf("60123456%03d", userID%1000),
		Date:             todayDate,
		LangChosen:       languages[userID%len(languages)],
		AllowBypass:      true,
		Tickets:          tickets,
	}
}

func generateRandomTickets() []TicketRequest {
	ticketsMutex.RLock()
	defer ticketsMutex.RUnlock()

	// If we have fetched tickets, use them randomly
	if ticketsFetched && len(availableTickets) > 0 {
		// Randomly select 1-3 different ticket types
		numTicketTypes := rand.Intn(min(3, len(availableTickets))) + 1
		selectedTickets := make([]TicketRequest, 0, numTicketTypes)

		// Create a copy of available tickets for random selection
		ticketPool := make([]TicketVariantDTO, len(availableTickets))
		copy(ticketPool, availableTickets)

		// Randomly shuffle the ticket pool
		for i := range ticketPool {
			j := rand.Intn(i + 1)
			ticketPool[i], ticketPool[j] = ticketPool[j], ticketPool[i]
		}

		// Select random tickets
		for i := 0; i < numTicketTypes; i++ {
			ticket := ticketPool[i]
			ticketId := "TIC-UNKNOWN"
			if ticket.TicketId != nil {
				ticketId = *ticket.TicketId
			}

			selectedTickets = append(selectedTickets, TicketRequest{
				TicketId: ticketId,
				Qty:      rand.Intn(3) + 1, // Random quantity between 1-3
			})
		}

		return selectedTickets
	}

	// Fallback to original hardcoded ticket if API fetch failed
	return []TicketRequest{
		{
			TicketId: "TIC-O-0020",
			Qty:      rand.Intn(3) + 1, // Random quantity between 1-3
		},
	}
}

func makeRequest(config LoadTestConfig, request CreateFreeOrderRequest) (int, string, error) {
	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return 0, "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/orderTicketGroup/free", config.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return resp.StatusCode, "", fmt.Errorf("failed to read response: %w", err)
	}

	responseBodyStr := responseBody.String()

	return resp.StatusCode, responseBodyStr, nil
}

func analyzeResults(results []TestResult) {
	fmt.Println("\n=================================================")
	fmt.Println("LOAD TEST RESULTS ANALYSIS")
	fmt.Println("=================================================")

	var (
		successCount  int
		errorCount    int
		totalDuration time.Duration
		minDuration   = time.Hour
		maxDuration   time.Duration
		statusCounts  = make(map[int]int)
	)

	for _, result := range results {
		// Count successes and errors
		if result.Error != nil {
			errorCount++
		} else if result.StatusCode >= 200 && result.StatusCode < 300 {
			successCount++
		} else if result.StatusCode >= 400 {
			// Count HTTP error status codes (4xx, 5xx) as failures
			errorCount++
		}

		// Track status codes
		statusCounts[result.StatusCode]++

		// Calculate duration statistics
		totalDuration += result.Duration
		if result.Duration < minDuration {
			minDuration = result.Duration
		}
		if result.Duration > maxDuration {
			maxDuration = result.Duration
		}
	}

	// Calculate statistics
	avgDuration := totalDuration / time.Duration(len(results))
	successRate := float64(successCount) / float64(len(results)) * 100

	// Display results
	fmt.Printf("Total Requests: %d\n", len(results))
	fmt.Printf("Successful Requests: %d\n", successCount)
	fmt.Printf("Failed Requests: %d\n", errorCount)
	fmt.Printf("Success Rate: %.2f%%\n", successRate)
	fmt.Println()

	fmt.Println("Response Time Statistics:")
	fmt.Printf("  Min Duration: %v\n", minDuration)
	fmt.Printf("  Max Duration: %v\n", maxDuration)
	fmt.Printf("  Avg Duration: %v\n", avgDuration)
	fmt.Println()

	fmt.Println("Status Code Distribution:")
	for statusCode, count := range statusCounts {
		fmt.Printf("  %d: %d requests\n", statusCode, count)
	}

	// Show sample error responses if any
	if errorCount > 0 {
		fmt.Println("\nSample Error Responses:")
		errorSamples := 0
		for _, result := range results {
			if result.Error != nil && errorSamples < 3 {
				fmt.Printf("  User %d: %v\n", result.UserID, result.Error)
				errorSamples++
			} else if result.StatusCode >= 400 && errorSamples < 3 {
				//fmt.Printf("  User %d (HTTP %d): %s\n", result.UserID, result.StatusCode,
				//	truncateString(result.ResponseBody, 100))
				fmt.Printf("  User %d (HTTP %d): %s\n", result.UserID, result.StatusCode, result.ResponseBody)
				errorSamples++
			}
		}
	}

	fmt.Println("\n=================================================")
}

//func truncateString(s string, maxLen int) string {
//	if len(s) <= maxLen {
//		return s
//	}
//	return s[:maxLen] + "..."
//}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
