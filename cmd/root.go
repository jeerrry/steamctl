package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "steamctl",
	Short: "Steam game idler and achievement unlocker",
	Long:  "A headless CLI tool that idles Steam games and unlocks achievements automatically, ordered from most common to rarest with randomized timing.",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default %s)", defaultConfigPath()))
}

func defaultConfigPath() string {
	home, err := homeDir()
	if err != nil {
		return "~/.steamctl/config.toml"
	}
	return home + "/.steamctl/config.toml"
}
