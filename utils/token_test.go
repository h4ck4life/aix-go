package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/h4ck4life/aix-go/constants"
)

// TestTokenResult holds the result of a token test
type TestTokenResult struct {
	Success bool
	Status  int
	Latency time.Duration
	Error   error
}

// ValidateToken makes an API request to validate a token
func ValidateToken(baseURL, token, tokenVar string) *TestTokenResult {
	start := time.Now()

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return &TestTokenResult{
			Success: false,
			Latency: time.Since(start),
			Error:   err,
		}
	}

	if tokenVar == constants.TokenTypeAPIKey {
		req.Header.Set("x-api-key", token)
		req.Header.Set("anthropic-api-key", token)
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := client.Do(req)
	if err != nil {
		return &TestTokenResult{
			Success: false,
			Latency: time.Since(start),
			Error:   err,
		}
	}
	defer resp.Body.Close()

	latency := time.Since(start)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return &TestTokenResult{
		Success: success,
		Status:  resp.StatusCode,
		Latency: latency,
		Error:   nil,
	}
}

// FormatTestResult formats a test result for display
func FormatTestResult(result *TestTokenResult) string {
	if result.Error != nil {
		return fmt.Sprintf("Failed: %v (%s)", result.Error, result.Latency)
	}
	if result.Success {
		return fmt.Sprintf("OK (%s)", result.Latency)
	}
	return fmt.Sprintf("HTTP %d (%s)", result.Status, result.Latency)
}
