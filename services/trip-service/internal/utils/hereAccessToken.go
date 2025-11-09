package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Cái thư viện Here Go siêu lỏ nên t vibe code lấy token
const OAUTH_URL = "https://account.api.here.com/oauth2/token"

var (
	keyID        = os.Getenv("HERE_ID")
	keySecret    = os.Getenv("HERE_SECRET")
	cachedToken  string
	cachedExpiry time.Time
	mu           sync.Mutex
)

type HereTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func percentEncode(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(url.QueryEscape(s), "+", "%20"), "%7E", "~")
}

func signRequest(method, baseURL string, params url.Values, consumerSecret string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, percentEncode(k)+"="+percentEncode(params.Get(k)))
	}

	paramString := strings.Join(pairs, "&")
	baseString := method + "&" + percentEncode(baseURL) + "&" + percentEncode(paramString)

	key := consumerSecret + "&"
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(baseString))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func FetchHereToken() (string, error) {
	mu.Lock()
	defer mu.Unlock()

	if cachedToken != "" && time.Now().Before(cachedExpiry) {
		return cachedToken, nil
	}

	oauthParams := url.Values{}
	oauthParams.Set("oauth_consumer_key", keyID)
	oauthParams.Set("oauth_nonce", strconv.FormatInt(rand.Int63(), 10))
	oauthParams.Set("oauth_signature_method", "HMAC-SHA256")
	oauthParams.Set("oauth_timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	oauthParams.Set("oauth_version", "1.0")
	oauthParams.Set("grant_type", "client_credentials")

	signature := signRequest("POST", OAUTH_URL, oauthParams, keySecret)
	oauthParams.Set("oauth_signature", signature)

	authHeaderParts := []string{
		fmt.Sprintf(`oauth_consumer_key="%s"`, percentEncode(keyID)),
		fmt.Sprintf(`oauth_nonce="%s"`, percentEncode(oauthParams.Get("oauth_nonce"))),
		fmt.Sprintf(`oauth_signature="%s"`, percentEncode(signature)),
		`oauth_signature_method="HMAC-SHA256"`,
		fmt.Sprintf(`oauth_timestamp="%s"`, oauthParams.Get("oauth_timestamp")),
		`oauth_version="1.0"`,
	}
	authHeader := "OAuth " + strings.Join(authHeaderParts, ", ")

	client := &http.Client{}
	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", OAUTH_URL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HERE API error: %s", string(body))
	}

	var result HereTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	cachedToken = result.AccessToken
	cachedExpiry = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)
	return cachedToken, nil
}
