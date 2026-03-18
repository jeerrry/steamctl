package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeerrry/steamctl/internal/config"
	"github.com/jeerrry/steamctl/internal/scheduler"
	"github.com/jeerrry/steamctl/internal/state"
	"github.com/jeerrry/steamctl/internal/steam"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon loop to idle games and unlock achievements",
	RunE:  runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(*cobra.Command, []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	st, err := state.Load("")
	if err != nil {
		return err
	}

	if len(cfg.Games) == 0 {
		fmt.Println("No games configured. Use 'steamctl add <appid>' to add one.")
		return nil
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	client := steam.NewClient(cfg.Steam.APIKey, cfg.Steam.SteamID)

	for _, g := range cfg.Games {
		if err := ctx.Err(); err != nil {
			slog.Info("shutting down")
			return nil
		}

		if err := processGame(ctx, client, st, g); err != nil {
			slog.Error("processing game", "game", g.Name, "app_id", g.AppID, "err", err)
			continue
		}
	}

	slog.Info("all games processed")
	return nil
}

func processGame(ctx context.Context, client *steam.Client, st *state.Store, g config.Game) error {
	slog.Info("processing game", "name", g.Name, "app_id", g.AppID)

	if !g.UnlockAchievements {
		slog.Info("achievement unlocking disabled, skipping", "name", g.Name)
		return nil
	}

	gs := st.GetGame(g.AppID)
	if gs != nil && gs.Completed {
		slog.Info("already completed", "name", g.Name)
		return nil
	}

	fetchCtx, fetchCancel := context.WithTimeout(ctx, 15*time.Second)
	defer fetchCancel()

	global, err := client.GetGlobalAchievementPercentages(fetchCtx, g.AppID)
	if err != nil {
		return fmt.Errorf("fetch global achievements: %w", err)
	}

	playerAchs, err := client.GetPlayerAchievements(fetchCtx, g.AppID)
	if err != nil {
		return fmt.Errorf("fetch player achievements: %w", err)
	}

	unlocked := make(map[string]bool, len(playerAchs))
	for _, a := range playerAchs {
		if a.Achieved == 1 {
			unlocked[a.APIName] = true
		}
	}

	startTime := time.Now()
	if gs != nil && !gs.NextUnlockAt.IsZero() && gs.NextUnlockAt.After(startTime) {
		startTime = gs.NextUnlockAt
	}

	schedule := scheduler.BuildSchedule(
		global,
		unlocked,
		startTime,
		g.TimeRange.Duration,
		g.MinInterval.Duration,
		g.MaxInterval.Duration,
	)

	if len(schedule) == 0 {
		slog.Info("all achievements unlocked", "name", g.Name)
		now := time.Now()
		st.SetGame(&state.GameState{
			AppID:            g.AppID,
			Name:             g.Name,
			AchievementIndex: len(global),
			Completed:        true,
			StartedAt:        now,
			CompletedAt:      now,
		})
		return st.Save("")
	}

	if gs == nil {
		gs = &state.GameState{
			AppID:     g.AppID,
			Name:      g.Name,
			StartedAt: time.Now(),
		}
	}

	slog.Info("starting unlock schedule",
		"name", g.Name,
		"remaining", len(schedule),
		"total_time", g.TimeRange.Duration,
	)

	for i, s := range schedule {
		waitDuration := time.Until(s.UnlockAt)
		if waitDuration > 0 {
			slog.Info("waiting for next unlock",
				"achievement", s.Name,
				"percent", fmt.Sprintf("%.2f%%", s.Percent),
				"wait", waitDuration.Truncate(time.Second),
			)

			select {
			case <-ctx.Done():
				slog.Info("interrupted, saving state")
				gs.AchievementIndex = len(unlocked) + i
				gs.NextUnlockAt = s.UnlockAt
				st.SetGame(gs)
				return st.Save("")
			case <-time.After(waitDuration):
			}
		}

		// SDK stub: log the unlock instead of calling Steamworks
		slog.Info("UNLOCK (stub)",
			"achievement", s.Name,
			"percent", fmt.Sprintf("%.2f%%", s.Percent),
			"index", fmt.Sprintf("%d/%d", i+1, len(schedule)),
		)

		gs.AchievementIndex = len(unlocked) + i + 1
		if i+1 < len(schedule) {
			gs.NextUnlockAt = schedule[i+1].UnlockAt
		}
		st.SetGame(gs)

		if err := st.Save(""); err != nil {
			slog.Error("saving state", "err", err)
		}
	}

	gs.Completed = true
	gs.CompletedAt = time.Now()
	st.SetGame(gs)

	if err := st.Save(""); err != nil {
		return fmt.Errorf("saving final state: %w", err)
	}

	slog.Info("completed all achievements", "name", g.Name)
	return nil
}
