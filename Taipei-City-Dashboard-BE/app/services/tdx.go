package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type tdxTokenCache struct {
	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

var cache = &tdxTokenCache{}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func getTDXToken() (string, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Use a 60s buffer to avoid token expiring mid-request
	if cache.accessToken != "" && time.Now().Add(60*time.Second).Before(cache.expiresAt) {
		return cache.accessToken, nil
	}

	clientID := os.Getenv("TDX_CLIENT_ID")
	clientSecret := os.Getenv("TDX_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("TDX_CLIENT_ID or TDX_CLIENT_SECRET environment variables not set")
	}

	tokenURL := "https://tdx.transportdata.tw/auth/realms/TDXConnect/protocol/openid-connect/token"
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenRes tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return "", err
	}

	cache.accessToken = tokenRes.AccessToken
	cache.expiresAt = time.Now().Add(time.Duration(tokenRes.ExpiresIn) * time.Second)

	return cache.accessToken, nil
}

// FetchTDXLiveBoard fetches the LiveBoard data for the given operator.
// operator should be "NTMETRO" or "TYMC".
func FetchTDXLiveBoard(operator string) ([]byte, error) {
	token, err := getTDXToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get TDX token: %w", err)
	}

	apiURL := fmt.Sprintf("https://tdx.transportdata.tw/api/basic/v2/Rail/Metro/LiveBoard/%s?$format=JSON", operator)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TDX API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
