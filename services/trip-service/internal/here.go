package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
	"trip-service/internal/utils"
)

type RouteResponse struct {
	Routes []struct {
		Sections []struct {
			Summary struct {
				Length   float64 `json:"length"`
				Duration float64 `json:"duration"`
			} `json:"summary"`
		} `json:"sections"`
	} `json:"routes"`
}
type RouteSummary struct {
	Distance float64
	Duration float64
	Fare     float64
}

func GetRouteSummary(ctx context.Context, origin, destination string) (*RouteSummary, error) {
	token, err := utils.FetchHereToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}

	baseURL := "https://router.hereapi.com/v8/routes"
	params := url.Values{}
	params.Set("transportMode", "car")
	params.Set("origin", origin)
	params.Set("destination", destination)
	params.Set("return", "summary")

	fullURL := baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HERE API error (%d): %s", resp.StatusCode, string(body))
	}

	var data RouteResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if len(data.Routes) == 0 || len(data.Routes[0].Sections) == 0 {
		return &RouteSummary{
			Distance: 0,
			Duration: 0,
		}, nil
	}

	summary := data.Routes[0].Sections[0].Summary
	return &RouteSummary{
		Distance: summary.Length,
		Duration: summary.Duration,
		Fare:     5 * (summary.Length / 1000), // ví dụ: 5 đô 1km frfr
	}, nil
}
