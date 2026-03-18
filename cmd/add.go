package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/jeerrry/steamctl/internal/config"
	"github.com/jeerrry/steamctl/internal/steam"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <appid>",
	Short: "Add a game to the config",
	Long:  "Fetches the game name from Steam and adds it to the config file with default settings.",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(_ *cobra.Command, args []string) error {
	appID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid app ID %q: %w", args[0], err)
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return err
	}

	for _, g := range cfg.Games {
		if g.AppID == appID {
			fmt.Printf("Game %d (%s) is already in config\n", g.AppID, g.Name)
			return nil
		}
	}

	client := steam.NewClient(cfg.Steam.APIKey, cfg.Steam.SteamID)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	details, err := client.GetAppDetails(ctx, appID)
	if err != nil {
		slog.Warn("could not fetch app name, using placeholder", "err", err)
		details = &steam.AppDetails{Name: fmt.Sprintf("App %d", appID)}
	}

	game := config.Game{
		AppID:              appID,
		Name:               details.Name,
		Idle:               true,
		UnlockAchievements: true,
		TimeRange:          config.Duration{Duration: 30 * time.Hour},
		MinInterval:        config.Duration{Duration: 10 * time.Minute},
		MaxInterval:        config.Duration{Duration: 2 * time.Hour},
	}

	cfg.Games = append(cfg.Games, game)

	if err := cfg.Save(cfgFile); err != nil {
		return err
	}

	fmt.Printf("Added %s (%d) to config\n", details.Name, appID)
	return nil
}
