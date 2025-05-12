package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// UserInfo holds Keycloak user profile fields
// retrieved from the /userinfo endpoint
type UserInfo struct {
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
}

// FetchUserInfo retrieves the user info from Keycloak's userinfo endpoint
func FetchUserInfo(ctx context.Context, issuerURL, accessToken string) (UserInfo, error) {
	userinfoURL := issuerURL + "/protocol/openid-connect/userinfo"
	req, err := http.NewRequestWithContext(ctx, "GET", userinfoURL, nil)
	if err != nil {
		return UserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UserInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UserInfo{}, fmt.Errorf("userinfo returned status %d", resp.StatusCode)
	}

	var ui UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&ui); err != nil {
		return UserInfo{}, err
	}
	return ui, nil
}
