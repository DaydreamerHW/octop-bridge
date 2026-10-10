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

var agentStatusCmd = &cobra.Command{
	Use:   "status [agent-id-or-name]",
	Short: "Inspect the real-time runtime status of an agent",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)
		target := cfg.DefaultAgentID
		if len(args) > 0 {
			target = args[0]
		}
		agent, err := octopClient.ResolveAgent(target)
		if err != nil {
			return err
		}

		status, err := octopClient.GetAgentStatus(agent.AgentID)
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(status, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("\nAgent Runtime Status: %s (%s)\n", agent.Name, agent.AgentID)
		fmt.Printf("State:       %s\n", status.State)
		if status.LastError != nil && *status.LastError != "" {
			fmt.Printf("Last Error:  %s\n", *status.LastError)
		} else {
			fmt.Printf("Last Error:  None\n")
		}
		if len(status.Channels) > 0 {
			var chNames []string
			for _, ch := range status.Channels {
				chNames = append(chNames, fmt.Sprintf("%s (%s)", ch.Name, ch.Kind))
			}
			fmt.Printf("Channels:    %s\n", strings.Join(chNames, ", "))
		} else {
			fmt.Printf("Channels:    None\n")
		}
		fmt.Println()
		return nil
	},
}

var agentStartCmd = &cobra.Command{
	Use:   "start [agent-id-or-name]",
	Short: "Start the agent runtime",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)
		target := cfg.DefaultAgentID
		if len(args) > 0 {
			target = args[0]
		}
		agent, err := octopClient.ResolveAgent(target)
		if err != nil {
			return err
		}

		if err := octopClient.StartAgent(agent.AgentID); err != nil {
			return err
		}
		fmt.Printf("✓ Started agent runtime: %s (%s)\n", agent.Name, agent.AgentID)
		return nil
	},
}

var agentStopCmd = &cobra.Command{
	Use:   "stop [agent-id-or-name]",
	Short: "Stop the agent runtime",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)
		target := cfg.DefaultAgentID
		if len(args) > 0 {
			target = args[0]
		}
		agent, err := octopClient.ResolveAgent(target)
		if err != nil {
			return err
		}

		if err := octopClient.StopAgent(agent.AgentID); err != nil {
			return err
		}
		fmt.Printf("✓ Stopped agent runtime: %s (%s)\n", agent.Name, agent.AgentID)
		return nil
	},
}

var agentReloadCmd = &cobra.Command{
	Use:     "reload [agent-id-or-name]",
	Aliases: []string{"restart"},
	Short:   "Reload the agent runtime, skills, and configuration",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)
		target := cfg.DefaultAgentID
		if len(args) > 0 {
			target = args[0]
		}
		agent, err := octopClient.ResolveAgent(target)
		if err != nil {
			return err
		}

		if err := octopClient.ReloadAgent(agent.AgentID); err != nil {
			return err
		}
		fmt.Printf("✓ Reloaded agent runtime: %s (%s)\n", agent.Name, agent.AgentID)
		return nil
	},
}

var flagAgentDeleteYes bool

var agentDeleteCmd = &cobra.Command{
	Use:     "delete <agent-id-or-name>",
	Aliases: []string{"rm"},
	Short:   "Permanently delete an agent and its workspace",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)
		agent, err := octopClient.ResolveAgent(args[0])
		if err != nil {
			return err
		}

		if !flagAgentDeleteYes {
			return fmt.Errorf("refusing to delete agent %q without confirmation; pass --yes / -y to confirm", agent.Name)
		}

		if err := octopClient.DeleteAgent(agent.AgentID); err != nil {
			return err
		}
		fmt.Printf("✓ Deleted agent %s (%s)\n", agent.Name, agent.AgentID)
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
	agentDeleteCmd.Flags().BoolVarP(&flagAgentDeleteYes, "yes", "y", false, "Confirm deletion without prompt")

	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentGetCmd)
	agentCmd.AddCommand(agentStatusCmd)
	agentCmd.AddCommand(agentStartCmd)
	agentCmd.AddCommand(agentStopCmd)
	agentCmd.AddCommand(agentReloadCmd)
	agentCmd.AddCommand(agentDeleteCmd)
	RootCmd.AddCommand(agentCmd)
	RootCmd.AddCommand(listAgentsShortcutCmd)
}
