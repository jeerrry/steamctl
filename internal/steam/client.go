// Package steam provides a client for the Steam Web API.
package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Client is a Steam Web API client.
type Client struct {
	apiKey       string
	steamID      string
	http         *http.Client
	baseURL      string // overridable for testing
	storeBaseURL string // overridable for testing
}

// NewClient creates a Steam Web API client with the given credentials.
func NewClient(apiKey, steamID string) *Client {
	return &Client{
		apiKey:       apiKey,
		steamID:      steamID,
		http:         &http.Client{},
		baseURL:      "https://api.steampowered.com",
		storeBaseURL: "https://store.steampowered.com",
	}
}

// OwnedGame represents a game in the user's Steam library.
type OwnedGame struct {
	AppID           int    `json:"appid"`
	Name            string `json:"name"`
	PlaytimeForever int    `json:"playtime_forever"`
}

// GetOwnedGames returns the list of games owned by the configured Steam user.
func (c *Client) GetOwnedGames(ctx context.Context) ([]OwnedGame, error) {
	params := url.Values{
		"key":                       {c.apiKey},
		"steamid":                   {c.steamID},
		"include_appinfo":           {"1"},
		"include_played_free_games": {"1"},
		"format":                    {"json"},
	}

	var resp struct {
		Response struct {
			Games []OwnedGame `json:"games"`
		} `json:"response"`
	}
	if err := c.get(ctx, "/IPlayerService/GetOwnedGames/v1/", params, &resp); err != nil {
		return nil, fmt.Errorf("get owned games: %w", err)
	}
	return resp.Response.Games, nil
}

// PlayerAchievement represents a single achievement's unlock status for a player.
type PlayerAchievement struct {
	APIName  string `json:"apiname"`
	Achieved int    `json:"achieved"`
}

// GetPlayerAchievements returns achievement unlock status for the configured user and given app.
func (c *Client) GetPlayerAchievements(ctx context.Context, appID int) ([]PlayerAchievement, error) {
	params := url.Values{
		"key":     {c.apiKey},
		"steamid": {c.steamID},
		"appid":   {strconv.Itoa(appID)},
		"format":  {"json"},
	}

	var resp struct {
		PlayerStats struct {
			Achievements []PlayerAchievement `json:"achievements"`
		} `json:"playerstats"`
	}
	if err := c.get(ctx, "/ISteamUserStats/GetPlayerAchievements/v1/", params, &resp); err != nil {
		return nil, fmt.Errorf("get player achievements for app %d: %w", appID, err)
	}
	return resp.PlayerStats.Achievements, nil
}

// GlobalAchievement represents an achievement's global unlock percentage.
type GlobalAchievement struct {
	Name    string  `json:"name"`
	Percent float64 `json:"percent"`
}

// GetGlobalAchievementPercentages returns global unlock percentages for the given app.
// This endpoint does not require an API key.
func (c *Client) GetGlobalAchievementPercentages(ctx context.Context, appID int) ([]GlobalAchievement, error) {
	params := url.Values{
		"gameid": {strconv.Itoa(appID)},
		"format": {"json"},
	}

	var resp struct {
		AchievementPercentages struct {
			Achievements []GlobalAchievement `json:"achievements"`
		} `json:"achievementpercentages"`
	}
	if err := c.get(ctx, "/ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2/", params, &resp); err != nil {
		return nil, fmt.Errorf("get global achievement percentages for app %d: %w", appID, err)
	}
	return resp.AchievementPercentages.Achievements, nil
}

// AppDetails holds basic information about a Steam app.
type AppDetails struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// GetAppDetails returns the name and type for a given app ID.
// Uses the store API which has a different base URL.
func (c *Client) GetAppDetails(ctx context.Context, appID int) (*AppDetails, error) {
	u := fmt.Sprintf("%s/api/appdetails?appids=%d", c.storeBaseURL, appID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get app details for %d: %w", appID, err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read app details response: %w", err)
	}

	// Response is keyed by appid string: {"12345": {"success": true, "data": {...}}}
	var raw map[string]struct {
		Success bool `json:"success"`
		Data    *struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode app details: %w", err)
	}

	entry, ok := raw[strconv.Itoa(appID)]
	if !ok || !entry.Success || entry.Data == nil {
		return nil, fmt.Errorf("app %d not found or request failed", appID)
	}

	return &AppDetails{
		Name: entry.Data.Name,
		Type: entry.Data.Type,
	}, nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, dst any) error {
	u := c.baseURL + path + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("steam API %s: status %d: %s", path, resp.StatusCode, body)
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
