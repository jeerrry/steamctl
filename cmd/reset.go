package cmd

import (
	"fmt"
	"strconv"

	"github.com/jeerrry/steamctl/internal/state"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset <appid>",
	Short: "Reset progress state for a game",
	Long:  "Clears the stored unlock progress for a game. Does not undo achievements already unlocked on Steam.",
	Args:  cobra.ExactArgs(1),
	RunE:  runReset,
}

func init() {
	rootCmd.AddCommand(resetCmd)
}

func runReset(_ *cobra.Command, args []string) error {
	appID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid app ID %q: %w", args[0], err)
	}

	st, err := state.Load("")
	if err != nil {
		return err
	}

	gs := st.GetGame(appID)
	if gs == nil {
		fmt.Printf("No state found for app %d\n", appID)
		return nil
	}

	name := gs.Name
	st.ResetGame(appID)

	if err := st.Save(""); err != nil {
		return err
	}

	fmt.Printf("Reset state for %s (%d)\n", name, appID)
	return nil
}
