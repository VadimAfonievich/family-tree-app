package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

var ErrNoUserData = errors.New("no user data in initData")

// ValidateInitData validates Telegram WebApp initData string.
// Returns telegram user data as map if valid.
func ValidateInitData(initData, botToken string) (map[string]string, error) {
	if botToken == "" {
		return parseInitDataValues(initData)
	}

	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("parse initData: %w", err)
	}

	hash := values.Get("hash")
	values.Del("hash")

	// Collect and sort data parts
	var parts []string
	for k, v := range values {
		parts = append(parts, k+"="+v[0])
	}
	sort.Strings(parts)

	// Build data check string
	dataCheckString := strings.Join(parts, "\n")

	// Calculate secret key
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(botToken))
	secretKey := h.Sum(nil)

	// Calculate hash
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))
	computedHash := hex.EncodeToString(mac.Sum(nil))

	if computedHash != hash {
		return nil, fmt.Errorf("invalid initData hash")
	}

	return parseInitDataValues(values.Encode())
}

// ParseTelegramUser parses the user JSON field from initData.
func ParseTelegramUser(userJSON string) (telegramID int64, username string, err error) {
	var user struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return 0, "", fmt.Errorf("parse user json: %w", err)
	}
	if user.ID == 0 {
		return 0, "", fmt.Errorf("no id in user data")
	}
	return user.ID, user.Username, nil
}

func parseInitDataValues(initData string) (map[string]string, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("parse initData: %w", err)
	}
	if values.Get("user") == "" {
		return nil, ErrNoUserData
	}
	result := make(map[string]string, len(values))
	for k, v := range values {
		result[k] = v[0]
	}
	return result, nil
}
