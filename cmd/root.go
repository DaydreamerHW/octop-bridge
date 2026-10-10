package cmd

import (
	"fmt"
	"os"

	"github.com/DaydreamerHW/octop-bridge/pkg/config"
	"github.com/spf13/cobra"
)

var (
	flagURL      string
	flagUsername string
	flagPassword string
	flagToken    string
	flagAgent    string
	flagInsecure bool
	flagJSON     bool
	flagTimeout  int
)

var RootCmd = &cobra.Command{
	Use:   "octop-bridge [query]",
	Short: "Octop Bridge: Unified CLI client to chat and delegate tasks to Octop AI agents",
	Long: `Octop Bridge is a universal CLI and AgentSkill bridge to chat, collaborate,
and delegate tasks to self-hosted Octop AI agents (like "小通") running on NAS or Linux servers.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		// If query given directly at root level, forward to chat command logic
		return runChat(args, false, "", "craft", false)
	},
}

func GetConfig() *config.Config {
	return config.Resolve(
		flagURL,
		flagUsername,
		flagPassword,
		flagToken,
		flagAgent,
		flagInsecure,
		flagTimeout,
	)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&flagURL, "url", "", "Octop server base URL (e.g. http://192.168.5.192:8089)")
	RootCmd.PersistentFlags().StringVarP(&flagUsername, "username", "u", "", "Octop login username")
	RootCmd.PersistentFlags().StringVarP(&flagPassword, "password", "p", "", "Octop login password")
	RootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Direct JWT Access Token (skips login)")
	RootCmd.PersistentFlags().StringVarP(&flagAgent, "agent", "a", "", "Target Agent ID or fuzzy name")
	RootCmd.PersistentFlags().BoolVarP(&flagInsecure, "insecure", "k", false, "Allow insecure TLS/SSL connections")
	RootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output result as JSON")
	RootCmd.PersistentFlags().IntVar(&flagTimeout, "timeout", 0, "Execution timeout in seconds")
}
