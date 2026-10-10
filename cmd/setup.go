package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/DaydreamerHW/octop-bridge/pkg/config"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactively configure Octop Bridge connection settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("=== Octop Bridge Setup Wizard ===")
		fmt.Println()

		fmt.Print("Octop Server URL [http://192.168.5.192:8089]: ")
		urlInput, _ := reader.ReadString('\n')
		urlInput = strings.TrimSpace(urlInput)
		if urlInput == "" {
			urlInput = "http://192.168.5.192:8089"
		}

		fmt.Print("Username [xqx]: ")
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)
		if userInput == "" {
			userInput = "xqx"
		}

		fmt.Print("Password [leave empty to keep default]: ")
		passInput, _ := reader.ReadString('\n')
		passInput = strings.TrimSpace(passInput)
		if passInput == "" {
			passInput = "Xqxq@1102@2026"
		}

		fmt.Print("Default Agent ID or Name [PQG34P]: ")
		agentInput, _ := reader.ReadString('\n')
		agentInput = strings.TrimSpace(agentInput)
		if agentInput == "" {
			agentInput = "PQG34P"
		}

		cfg := &config.Config{
			BaseURL:        urlInput,
			Username:       userInput,
			Password:       passInput,
			DefaultAgentID: agentInput,
			Timeout:        180,
		}

		if err := config.SaveFileConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		path, _ := config.ConfigFilePath()
		fmt.Printf("\n✓ Configuration saved to %s\n", path)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(setupCmd)
}
