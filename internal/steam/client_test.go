package steam

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetOwnedGames(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/IPlayerService/GetOwnedGames/v1/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"response": {
				"games": [
					{"appid": 440, "name": "Team Fortress 2", "playtime_forever": 100},
					{"appid": 730, "name": "Counter-Strike 2", "playtime_forever": 200}
				]
			}
		}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	games, err := c.GetOwnedGames(context.Background())
	if err != nil {
		t.Fatalf("GetOwnedGames() error: %v", err)
	}
	if len(games) != 2 {
		t.Fatalf("len(games) = %d, want 2", len(games))
	}
	if games[0].AppID != 440 {
		t.Errorf("games[0].AppID = %d, want 440", games[0].AppID)
	}
	if games[1].Name != "Counter-Strike 2" {
		t.Errorf("games[1].Name = %q, want Counter-Strike 2", games[1].Name)
	}
}

func TestGetPlayerAchievements(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"playerstats": {
				"achievements": [
					{"apiname": "ACH_1", "achieved": 1},
					{"apiname": "ACH_2", "achieved": 0}
				]
			}
		}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	achs, err := c.GetPlayerAchievements(context.Background(), 440)
	if err != nil {
		t.Fatalf("GetPlayerAchievements() error: %v", err)
	}
	if len(achs) != 2 {
		t.Fatalf("len(achievements) = %d, want 2", len(achs))
	}
	if achs[0].Achieved != 1 {
		t.Errorf("achievements[0].Achieved = %d, want 1", achs[0].Achieved)
	}
}

func TestGetGlobalAchievementPercentages(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"achievementpercentages": {
				"achievements": [
					{"name": "ACH_1", "percent": 85.5},
					{"name": "ACH_2", "percent": 12.3}
				]
			}
		}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	achs, err := c.GetGlobalAchievementPercentages(context.Background(), 440)
	if err != nil {
		t.Fatalf("GetGlobalAchievementPercentages() error: %v", err)
	}
	if len(achs) != 2 {
		t.Fatalf("len(achievements) = %d, want 2", len(achs))
	}
	if achs[0].Percent != 85.5 {
		t.Errorf("achievements[0].Percent = %f, want 85.5", achs[0].Percent)
	}
}

func TestGetAppDetails(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"440": {
				"success": true,
				"data": {
					"name": "Team Fortress 2",
					"type": "game"
				}
			}
		}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	details, err := c.GetAppDetails(context.Background(), 440)
	if err != nil {
		t.Fatalf("GetAppDetails() error: %v", err)
	}
	if details.Name != "Team Fortress 2" {
		t.Errorf("Name = %q, want Team Fortress 2", details.Name)
	}
	if details.Type != "game" {
		t.Errorf("Type = %q, want game", details.Type)
	}
}

func TestGetOwnedGames_APIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("access denied"))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.GetOwnedGames(context.Background())
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
}

func newTestClient(serverURL string) *Client {
	c := NewClient("testkey", "testid")
	c.baseURL = serverURL
	c.storeBaseURL = serverURL
	return c
}
