package squadstats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Client handles interaction with the mysquadstats API.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new SquadStats API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetServerData fetches the server status including the current matchID.
func (c *Client) GetServerData(ctx context.Context) ([]ServerPlayer, error) {
	url := "https://mysquadstats.com/serversPlayers/480081.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Origin", "https://mysquadstats.com")
	req.Header.Set("Referer", "https://mysquadstats.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data []ServerPlayer
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return data, nil
}

// GetMatchDetails fetches detailed information about a match by its ID.
func (c *Client) GetMatchDetails(ctx context.Context, matchID int) (*MatchResponse, error) {
	url := fmt.Sprintf("https://api.mysquadstats.com/matches?matchID=%d", matchID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Origin", "https://mysquadstats.com")
	req.Header.Set("Referer", "https://mysquadstats.com/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	slog.Info("Retorno cru da API (matches)", "raw_body", string(bodyBytes))

	var matchData MatchResponse
	if err := json.Unmarshal(bodyBytes, &matchData); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if matchData.Status == "Error" || (matchData.Match.MapClassname == "" && matchData.Match.LayerClassname == "") {
		return nil, fmt.Errorf("invalid match data (status: %s, message: %s)", matchData.Status, matchData.Message)
	}

	if matchData.Match.Map == "" {
		matchData.Match.Map = matchData.Match.MapClassname
	}
	if matchData.Match.Layer == "" {
		matchData.Match.Layer = matchData.Match.LayerClassname
	}

	return &matchData, nil
}
