package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DaydreamerHW/octop-bridge/pkg/client"
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:     "agent",
	Aliases: []string{"agents"},
	Short:   "Manage and inspect Octop agents",
}

var agentListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all available agents on Octop",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		agents, err := octopClient.ListAgents()
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(agents, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("\nConnected to Octop: %s\n", cfg.HTTPURL())
		fmt.Printf("Found %d agent(s):\n\n", len(agents))
		fmt.Printf("%-10s | %-20s | %-28s | %-8s\n", "ID", "NAME", "MODEL", "STATUS")
		fmt.Println(strings.Repeat("-", 75))
		for _, a := range agents {
			status := "Active"
			if !a.Enabled {
				status = "Disabled"
			}
			model := a.DefaultModel
			if model == "" {
				model = "Default"
			}
			fmt.Printf("%-10s | %-20s | %-28s | %-8s\n", a.AgentID, a.Name, model, status)
		}
		fmt.Println()
		return nil
	},
}

var agentGetCmd = &cobra.Command{
	Use:   "get <agent-id-or-name>",
	Short: "Get details of a specific agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		agent, err := octopClient.ResolveAgent(args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(agent, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("\nAgent Details: %s\n", agent.Name)
		fmt.Printf("ID:          %s\n", agent.AgentID)
		fmt.Printf("Description: %s\n", agent.Description)
		fmt.Printf("Model:       %s\n", agent.DefaultModel)
		fmt.Printf("Enabled:     %v\n", agent.Enabled)
		fmt.Printf("Shared:      %v\n", agent.IsShared)
		fmt.Println()
		return nil
	},
}

// Top-level shortcut alias for list-agents
var listAgentsShortcutCmd = &cobra.Command{
	Use:     "list-agents",
	Aliases: []string{"ls"},
	Short:   "Shortcut to list all available agents",
	RunE:    agentListCmd.RunE,
}

func init() {
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentGetCmd)
	RootCmd.AddCommand(agentCmd)
	RootCmd.AddCommand(listAgentsShortcutCmd)
}
