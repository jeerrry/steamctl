package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jeerrry/steamctl/internal/config"
	"github.com/jeerrry/steamctl/internal/scheduler"
	"github.com/jeerrry/steamctl/internal/steam"
	"github.com/spf13/cobra"
)

var dryRunCmd = &cobra.Command{
	Use:   "dry-run",
	Short: "Preview the unlock schedule without executing",
	RunE:  runDryRun,
}

func init() {
	rootCmd.AddCommand(dryRunCmd)
}

func runDryRun(*cobra.Command, []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	if len(cfg.Games) == 0 {
		fmt.Println("No games configured. Use 'steamctl add <appid>' to add one.")
		return nil
	}

	client := steam.NewClient(cfg.Steam.APIKey, cfg.Steam.SteamID)

	for _, g := range cfg.Games {
		if !g.UnlockAchievements {
			fmt.Printf("\n%s (%d): achievement unlocking disabled\n", g.Name, g.AppID)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		global, err := client.GetGlobalAchievementPercentages(ctx, g.AppID)
		if err != nil {
			cancel()
			return fmt.Errorf("fetch global achievements for %s: %w", g.Name, err)
		}

		playerAchs, err := client.GetPlayerAchievements(ctx, g.AppID)
		cancel()
		if err != nil {
			return fmt.Errorf("fetch player achievements for %s: %w", g.Name, err)
		}

		unlocked := make(map[string]bool, len(playerAchs))
		for _, a := range playerAchs {
			if a.Achieved == 1 {
				unlocked[a.APIName] = true
			}
		}

		schedule := scheduler.BuildSchedule(
			global,
			unlocked,
			time.Now(),
			g.TimeRange.Duration,
			g.MinInterval.Duration,
			g.MaxInterval.Duration,
		)

		fmt.Printf("\n%s (%d) — %d achievements to unlock over %s\n",
			g.Name, g.AppID, len(schedule), g.TimeRange.Duration)
		fmt.Printf("  Already unlocked: %d/%d\n\n", len(unlocked), len(global))

		if len(schedule) == 0 {
			fmt.Println("  All achievements unlocked!")
			continue
		}

		for i, s := range schedule {
			offset := s.UnlockAt.Sub(schedule[0].UnlockAt)
			fmt.Printf("  %3d. [%6.2f%%] %-40s +%s\n",
				i+1, s.Percent, s.Name, offset.Truncate(time.Second))
		}
	}

	fmt.Println()
	return nil
}
