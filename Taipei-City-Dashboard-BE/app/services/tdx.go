package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const tdxTokenURL = "https://tdx.transportdata.tw/auth/realms/TDXConnect/protocol/openid-connect/token"

type tdxTokenCache struct {
	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

var tdxCache = &tdxTokenCache{}

// getTDXToken returns a valid TDX bearer token, refreshing if expired.
func getTDXToken() (string, error) {
	tdxCache.mu.Lock()
	defer tdxCache.mu.Unlock()

	if time.Now().Before(tdxCache.expiresAt.Add(-60 * time.Second)) {
		return tdxCache.accessToken, nil
	}

	clientID := os.Getenv("TDX_CLIENT_ID")
	clientSecret := os.Getenv("TDX_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("TDX_CLIENT_ID or TDX_CLIENT_SECRET not set")
	}

	resp, err := http.PostForm(tdxTokenURL, map[string][]string{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	})
	if err != nil {
		return "", fmt.Errorf("TDX token request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("TDX token decode failed: %w", err)
	}

	tdxCache.accessToken = result.AccessToken
	tdxCache.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return tdxCache.accessToken, nil
}

// FetchTDXLiveBoard calls the TDX LiveBoard API for the given operator
// and returns the raw JSON bytes.
// operator is one of: "NTMETRO", "TYMC"
func FetchTDXLiveBoard(operator string) ([]byte, error) {
	token, err := getTDXToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://tdx.transportdata.tw/api/basic/v2/Rail/Metro/LiveBoard/%s?$format=JSON",
		operator,
	)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("TDX LiveBoard request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TDX LiveBoard returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
