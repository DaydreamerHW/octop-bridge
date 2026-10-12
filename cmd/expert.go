package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DaydreamerHW/octop-bridge/pkg/client"
	"github.com/spf13/cobra"
)

var expertCmd = &cobra.Command{
	Use:     "expert",
	Aliases: []string{"experts"},
	Short:   "Browse expert templates and instantiate expert agents",
}

var expertListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all available expert templates on Octop",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		experts, err := octopClient.ListExperts()
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(experts, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("\nConnected to Octop: %s\n", cfg.HTTPURL())
		fmt.Printf("Found %d expert template(s):\n\n", len(experts))
		fmt.Printf("%-28s | %-22s | %-36s\n", "TEMPLATE ID", "NAME", "DESCRIPTION")
		fmt.Println(strings.Repeat("-", 90))
		for _, e := range experts {
			name := e.Label.ZH
			if name == "" {
				name = e.Label.EN
			}
			desc := e.Description.ZH
			if desc == "" {
				desc = e.Description.EN
			}
			// Truncate desc if too long for clean table
			runes := []rune(desc)
			if len(runes) > 34 {
				desc = string(runes[:33]) + "…"
			}
			fmt.Printf("%-28s | %-22s | %-36s\n", e.ID, name, desc)
		}
		fmt.Println()
		return nil
	},
}

var expertGetCmd = &cobra.Command{
	Use:   "get <expert-id>",
	Short: "Get details of a specific expert template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		expert, err := octopClient.GetExpert(args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(expert, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		name := expert.Label.ZH
		if name == "" {
			name = expert.Label.EN
		}
		desc := expert.Description.ZH
		if desc == "" {
			desc = expert.Description.EN
		}
		welcome := expert.WelcomeMessage.ZH
		if welcome == "" {
			welcome = expert.WelcomeMessage.EN
		}

		fmt.Printf("\nExpert Template Details: %s (%s)\n", name, expert.ID)
		fmt.Printf("Description: %s\n", desc)
		if welcome != "" {
			fmt.Printf("Welcome Msg: %s\n", welcome)
		}
		if expert.IconName != "" {
			fmt.Printf("Icon:        %s\n", expert.IconName)
		}
		if expert.Color != "" {
			fmt.Printf("Color:       %s\n", expert.Color)
		}
		fmt.Println()
		return nil
	},
}

var (
	flagExpertCreateName  string
	flagExpertCreateModel string
	flagExpertCreateColor string
)

var expertCreateCmd = &cobra.Command{
	Use:   "create <expert-id>",
	Short: "Create an active agent from an expert template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		req := client.CreateAgentFromExpertRequest{
			Name:         flagExpertCreateName,
			DefaultModel: flagExpertCreateModel,
			Color:        flagExpertCreateColor,
		}

		agent, err := octopClient.CreateAgentFromExpert(args[0], req)
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(agent, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Created expert agent: %s (%s)\n", agent.Name, agent.AgentID)
		if agent.Description != "" {
			fmt.Printf("  Description: %s\n", agent.Description)
		}
		if agent.DefaultModel != "" {
			fmt.Printf("  Model:       %s\n", agent.DefaultModel)
		}
		return nil
	},
}

var listExpertsShortcutCmd = &cobra.Command{
	Use:     "list-experts",
	Aliases: []string{"experts"},
	Short:   "Shortcut to list all available expert templates",
	RunE:    expertListCmd.RunE,
}

func init() {
	expertCreateCmd.Flags().StringVar(&flagExpertCreateName, "name", "", "Custom name for the instantiated agent")
	expertCreateCmd.Flags().StringVarP(&flagExpertCreateModel, "model", "m", "", "Override default model for the agent")
	expertCreateCmd.Flags().StringVar(&flagExpertCreateColor, "color", "", "Hex color for the agent")

	expertCmd.AddCommand(expertListCmd)
	expertCmd.AddCommand(expertGetCmd)
	expertCmd.AddCommand(expertCreateCmd)

	RootCmd.AddCommand(expertCmd)
	RootCmd.AddCommand(listExpertsShortcutCmd)
}
