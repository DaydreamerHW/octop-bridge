package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DaydreamerHW/octop-bridge/pkg/client"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{
	Use:     "team",
	Aliases: []string{"teams"},
	Short:   "Manage and inspect Octop expert teams",
}

var teamListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all expert teams on Octop",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		teams, err := octopClient.ListTeams()
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(teams, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("\nConnected to Octop: %s\n", cfg.HTTPURL())
		fmt.Printf("Found %d expert team(s):\n\n", len(teams))
		fmt.Printf("%-10s | %-22s | %-22s | %-8s | %-20s\n", "ID", "TEAM NAME", "MODEL", "STATUS", "MEMBERS")
		fmt.Println(strings.Repeat("-", 90))
		for _, t := range teams {
			status := t.State
			if status == "" {
				status = "Active"
			}
			model := t.DefaultModel
			if model == "" {
				model = "Default"
			}
			var memberNames []string
			for _, m := range t.Members {
				memberNames = append(memberNames, m.Name)
			}
			membersStr := fmt.Sprintf("%d (%s)", len(t.MemberIDs), strings.Join(memberNames, ", "))
			runes := []rune(membersStr)
			if len(runes) > 20 {
				membersStr = string(runes[:19]) + "…"
			}
			fmt.Printf("%-10s | %-22s | %-22s | %-8s | %-20s\n", t.TeamID, t.Name, model, status, membersStr)
		}
		fmt.Println()
		return nil
	},
}

var teamGetCmd = &cobra.Command{
	Use:   "get <team-id-or-name>",
	Short: "Get details and roster of a specific expert team",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		team, err := octopClient.ResolveTeam(args[0])
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(team, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("\nExpert Team: %s\n", team.Name)
		fmt.Printf("ID:          %s\n", team.TeamID)
		if team.Description != "" {
			fmt.Printf("Description: %s\n", team.Description)
		}
		if team.DefaultModel != "" {
			fmt.Printf("Model:       %s\n", team.DefaultModel)
		}
		fmt.Printf("State:       %s\n", team.State)
		if team.WelcomeMessage != "" {
			fmt.Printf("Welcome Msg: %s\n", team.WelcomeMessage)
		}

		fmt.Printf("\nTeam Roster (%d member(s)):\n", len(team.Members))
		fmt.Printf("%-10s | %-24s | %-10s\n", "AGENT ID", "NAME", "STATUS")
		fmt.Println(strings.Repeat("-", 50))
		for _, m := range team.Members {
			state := m.State
			if state == "" {
				state = "Active"
			}
			fmt.Printf("%-10s | %-24s | %-10s\n", m.AgentID, m.Name, state)
		}
		fmt.Println()
		return nil
	},
}

var (
	flagTeamCreateName        string
	flagTeamCreateDescription string
	flagTeamCreateModel       string
	flagTeamCreateMembers     string
	flagTeamCreateColor       string
	flagTeamCreateWelcome     string
)

var teamCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new expert team",
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(flagTeamCreateName) == "" {
			return fmt.Errorf("team name is required (--name)")
		}

		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		var memberIDs []string
		if flagTeamCreateMembers != "" {
			parts := strings.Split(flagTeamCreateMembers, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				// Resolve fuzzy member name/ID to exact agent ID
				if resolved, err := octopClient.ResolveAgent(p); err == nil {
					memberIDs = append(memberIDs, resolved.AgentID)
				} else {
					memberIDs = append(memberIDs, p)
				}
			}
		}

		req := client.CreateTeamRequest{
			Name:           flagTeamCreateName,
			Description:    flagTeamCreateDescription,
			DefaultModel:   flagTeamCreateModel,
			Color:          flagTeamCreateColor,
			WelcomeMessage: flagTeamCreateWelcome,
			MemberIDs:      memberIDs,
		}

		team, err := octopClient.CreateTeam(req)
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(team, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Created expert team: %s (%s)\n", team.Name, team.TeamID)
		if len(team.MemberIDs) > 0 {
			fmt.Printf("  Members: %s\n", strings.Join(team.MemberIDs, ", "))
		}
		return nil
	},
}

var (
	flagTeamUpdateName        string
	flagTeamUpdateDescription string
	flagTeamUpdateModel       string
	flagTeamUpdateMembers     string
	flagTeamUpdateWelcome     string
)

var teamUpdateCmd = &cobra.Command{
	Use:   "update <team-id-or-name>",
	Short: "Update an expert team's metadata or roster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		existing, err := octopClient.ResolveTeam(args[0])
		if err != nil {
			return err
		}

		req := client.UpdateTeamRequest{}
		if cmd.Flags().Changed("name") {
			req.Name = &flagTeamUpdateName
		}
		if cmd.Flags().Changed("description") {
			req.Description = &flagTeamUpdateDescription
		}
		if cmd.Flags().Changed("model") {
			req.DefaultModel = &flagTeamUpdateModel
		}
		if cmd.Flags().Changed("welcome") {
			req.WelcomeMessage = &flagTeamUpdateWelcome
		}
		if cmd.Flags().Changed("members") {
			var memberIDs []string
			if flagTeamUpdateMembers != "" {
				parts := strings.Split(flagTeamUpdateMembers, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					if resolved, err := octopClient.ResolveAgent(p); err == nil {
						memberIDs = append(memberIDs, resolved.AgentID)
					} else {
						memberIDs = append(memberIDs, p)
					}
				}
			}
			req.MemberIDs = &memberIDs
		}

		updated, err := octopClient.UpdateTeam(existing.TeamID, req)
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(updated, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Updated expert team: %s (%s)\n", updated.Name, updated.TeamID)
		if len(updated.MemberIDs) > 0 {
			fmt.Printf("  Members: %s\n", strings.Join(updated.MemberIDs, ", "))
		}
		return nil
	},
}

var flagTeamDeleteYes bool

var teamDeleteCmd = &cobra.Command{
	Use:     "delete <team-id-or-name>",
	Aliases: []string{"rm"},
	Short:   "Permanently delete an expert team",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		team, err := octopClient.ResolveTeam(args[0])
		if err != nil {
			return err
		}

		if !flagTeamDeleteYes {
			return fmt.Errorf("refusing to delete expert team %q without confirmation; pass --yes / -y to confirm", team.Name)
		}

		if err := octopClient.DeleteTeam(team.TeamID); err != nil {
			return err
		}

		fmt.Printf("✓ Deleted expert team %s (%s)\n", team.Name, team.TeamID)
		return nil
	},
}

var listTeamsShortcutCmd = &cobra.Command{
	Use:     "list-teams",
	Aliases: []string{"teams"},
	Short:   "Shortcut to list all expert teams",
	RunE:    teamListCmd.RunE,
}

func init() {
	teamCreateCmd.Flags().StringVar(&flagTeamCreateName, "name", "", "Team name (required)")
	teamCreateCmd.Flags().StringVarP(&flagTeamCreateDescription, "description", "d", "", "Team description")
	teamCreateCmd.Flags().StringVarP(&flagTeamCreateModel, "model", "m", "", "Default model for the team host")
	teamCreateCmd.Flags().StringVar(&flagTeamCreateMembers, "members", "", "Comma-separated agent IDs or fuzzy names for team members")
	teamCreateCmd.Flags().StringVar(&flagTeamCreateColor, "color", "", "Hex color theme for team")
	teamCreateCmd.Flags().StringVar(&flagTeamCreateWelcome, "welcome", "", "Custom welcome message")

	teamUpdateCmd.Flags().StringVar(&flagTeamUpdateName, "name", "", "New team name")
	teamUpdateCmd.Flags().StringVarP(&flagTeamUpdateDescription, "description", "d", "", "New description")
	teamUpdateCmd.Flags().StringVarP(&flagTeamUpdateModel, "model", "m", "", "New default model")
	teamUpdateCmd.Flags().StringVar(&flagTeamUpdateMembers, "members", "", "Comma-separated updated agent IDs or fuzzy names")
	teamUpdateCmd.Flags().StringVar(&flagTeamUpdateWelcome, "welcome", "", "New welcome message")

	teamDeleteCmd.Flags().BoolVarP(&flagTeamDeleteYes, "yes", "y", false, "Confirm deletion without prompt")

	teamCmd.AddCommand(teamListCmd)
	teamCmd.AddCommand(teamGetCmd)
	teamCmd.AddCommand(teamCreateCmd)
	teamCmd.AddCommand(teamUpdateCmd)
	teamCmd.AddCommand(teamDeleteCmd)

	RootCmd.AddCommand(teamCmd)
	RootCmd.AddCommand(listTeamsShortcutCmd)
}
