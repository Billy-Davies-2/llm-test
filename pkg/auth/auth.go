package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	oidc "github.com/coreos/go-oidc"
	oauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// OIDCConfig holds Keycloak endpoints and client info
type OIDCConfig struct {
	IssuerURL string // e.g. "https://auth.example.com/auth/realms/llm"
	ClientID  string // e.g. "llm-client"
}

// DeviceFlowResult contains tokens from device auth
type DeviceFlowResult struct {
	AccessToken  string    // Bearer token
	RefreshToken string    // Refresh token
	Expiry       time.Time // Expiry time of AccessToken
}

// RunDeviceFlow starts the OAuth2 Device Code Flow with Keycloak.
func RunDeviceFlow(ctx context.Context, cfg OIDCConfig) (*DeviceFlowResult, error) {
	// Discover endpoints
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, err
	}

	// Request device/user codes
	deviceResp, err := requestDeviceCode(ctx, cfg.IssuerURL, cfg.ClientID)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\nVisit %s and enter code: %s\n", deviceResp.VerificationURI, deviceResp.UserCode)

	// Poll for token
	oauthCfg := clientcredentials.Config{
		ClientID:  cfg.ClientID,
		TokenURL:  deviceResp.TokenURL,
		AuthStyle: oauth2.AuthStyleInParams,
	}

	ticker := time.NewTicker(time.Duration(deviceResp.Interval) * time.Second)
	defer ticker.Stop()
	timeout := time.After(time.Duration(deviceResp.ExpiresIn) * time.Second)

	for {
		select {
		case <-ticker.C:
			tok, err := oauthCfg.Token(ctx)
			if err == nil {
				return &DeviceFlowResult{
					AccessToken:  tok.AccessToken,
					RefreshToken: tok.RefreshToken,
					Expiry:       tok.Expiry,
				}, nil
			}
		case <-timeout:
			return nil, fmt.Errorf("device flow timeout")
		}
	}
}

// deviceCodeResponse holds device auth response fields
type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	TokenURL        string // full token endpoint
}

// requestDeviceCode does the HTTP call to get device code
func requestDeviceCode(ctx context.Context, issuer, clientID string) (*deviceCodeResponse, error) {
	discoURL := issuer + "/protocol/openid-connect/auth/device"
	values := url.Values{}
	values.Set("client_id", clientID)

	req, err := http.NewRequestWithContext(ctx, "POST", discoURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var codeResp deviceCodeResponse
	if err := json.Unmarshal(body, &codeResp); err != nil {
		return nil, err
	}
	codeResp.TokenURL = issuer + "/protocol/openid-connect/token"
	return &codeResp, nil
}
