package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jeerrry/steamctl/internal/config"
	"github.com/jeerrry/steamctl/internal/state"
	"github.com/jeerrry/steamctl/internal/steam"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show progress for all configured games",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(*cobra.Command, []string) error {
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

	client := steam.NewClient(cfg.Steam.APIKey, cfg.Steam.SteamID)

	for _, g := range cfg.Games {
		fmt.Printf("\n%s (%d)\n", g.Name, g.AppID)

		gs := st.GetGame(g.AppID)
		if gs == nil {
			fmt.Println("  Status: not started")
			continue
		}

		if gs.Completed {
			fmt.Printf("  Status: completed at %s\n", gs.CompletedAt.Format(time.RFC822))
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		global, err := client.GetGlobalAchievementPercentages(ctx, g.AppID)
		cancel()

		total := len(global)
		if err != nil {
			total = gs.AchievementIndex // best guess
		}

		fmt.Printf("  Status: in progress\n")
		fmt.Printf("  Progress: %d/%d achievements unlocked\n", gs.AchievementIndex, total)
		if !gs.NextUnlockAt.IsZero() {
			remaining := time.Until(gs.NextUnlockAt)
			if remaining > 0 {
				fmt.Printf("  Next unlock: %s (in %s)\n", gs.NextUnlockAt.Format(time.RFC822), remaining.Truncate(time.Second))
			} else {
				fmt.Printf("  Next unlock: pending\n")
			}
		}
	}

	fmt.Println()
	return nil
}
