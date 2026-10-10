package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DaydreamerHW/octop-bridge/pkg/client"
	"github.com/spf13/cobra"
)

var (
	skillFlagOverwrite bool
	skillFlagKind      string
)

var skillCmd = &cobra.Command{
	Use:     "skill",
	Aliases: []string{"skills"},
	Short:   "Manage and deploy skills on Octop agents",
}

var skillListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List installed and built-in skills for an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		agent, err := octopClient.ResolveAgent(cfg.DefaultAgentID)
		if err != nil {
			return err
		}

		skills, err := octopClient.ListSkills(agent.AgentID)
		if err != nil {
			return err
		}

		var filtered []client.SkillSummary
		for _, s := range skills {
			if skillFlagKind != "" && s.Kind != skillFlagKind {
				continue
			}
			filtered = append(filtered, s)
		}

		if flagJSON {
			data, _ := json.MarshalIndent(filtered, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("\nSkills for Agent: %s (%s)\n", agent.Name, agent.AgentID)
		fmt.Printf("Total: %d skill(s):\n\n", len(filtered))
		fmt.Printf("%-24s | %-10s | %-8s | %s\n", "NAME", "KIND", "STATUS", "DESCRIPTION")
		fmt.Println(strings.Repeat("-", 80))
		for _, s := range filtered {
			status := "Enabled"
			if !s.Enabled {
				status = "Disabled"
			}
			kind := s.Kind
			if kind == "" {
				kind = "builtin"
			}
			desc := strings.ReplaceAll(s.Description, "\n", " ")
			if len(desc) > 36 {
				desc = desc[:33] + "..."
			}
			fmt.Printf("%-24s | %-10s | %-8s | %s\n", s.Name, kind, status, desc)
		}
		fmt.Println()
		return nil
	},
}

var skillEnableCmd = &cobra.Command{
	Use:   "enable <skill-name>",
	Short: "Enable a skill for the agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)
		skillName := args[0]

		agent, err := octopClient.ResolveAgent(cfg.DefaultAgentID)
		if err != nil {
			return err
		}

		if err := octopClient.EnableSkill(agent.AgentID, skillName); err != nil {
			return err
		}

		if flagJSON {
			res := map[string]string{
				"status":     "enabled",
				"agent_id":   agent.AgentID,
				"skill_name": skillName,
			}
			data, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Enabled skill %q for agent %s (%s)\n", skillName, agent.Name, agent.AgentID)
		return nil
	},
}

var skillDisableCmd = &cobra.Command{
	Use:   "disable <skill-name>",
	Short: "Disable a skill for the agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)
		skillName := args[0]

		agent, err := octopClient.ResolveAgent(cfg.DefaultAgentID)
		if err != nil {
			return err
		}

		if err := octopClient.DisableSkill(agent.AgentID, skillName); err != nil {
			return err
		}

		if flagJSON {
			res := map[string]string{
				"status":     "disabled",
				"agent_id":   agent.AgentID,
				"skill_name": skillName,
			}
			data, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Disabled skill %q for agent %s (%s)\n", skillName, agent.Name, agent.AgentID)
		return nil
	},
}

var skillRemoveCmd = &cobra.Command{
	Use:     "remove <skill-name>",
	Aliases: []string{"rm", "delete"},
	Short:   "Delete a workspace skill from the agent",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)
		skillName := args[0]

		agent, err := octopClient.ResolveAgent(cfg.DefaultAgentID)
		if err != nil {
			return err
		}

		if err := octopClient.DeleteSkill(agent.AgentID, skillName); err != nil {
			return err
		}

		if flagJSON {
			res := map[string]string{
				"status":     "deleted",
				"agent_id":   agent.AgentID,
				"skill_name": skillName,
			}
			data, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Removed skill %q from agent %s (%s)\n", skillName, agent.Name, agent.AgentID)
		return nil
	},
}

var skillInstallCmd = &cobra.Command{
	Use:     "install <path-to-skill-directory-or-zip>",
	Aliases: []string{"add"},
	Short:   "Remotely deploy and install a skill folder or zip file to Octop",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)
		localPath := args[0]

		agent, err := octopClient.ResolveAgent(cfg.DefaultAgentID)
		if err != nil {
			return err
		}

		installed, err := octopClient.InstallSkill(agent.AgentID, localPath, skillFlagOverwrite)
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(installed, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Successfully installed & enabled skill %q for agent %s (%s)\n", installed.Name, agent.Name, agent.AgentID)
		return nil
	},
}

// Top-level shortcut alias for list-skills
var listSkillsShortcutCmd = &cobra.Command{
	Use:     "list-skills",
	Aliases: []string{"skills-list"},
	Short:   "Shortcut to list all installed skills",
	RunE:    skillListCmd.RunE,
}

func init() {
	skillListCmd.Flags().StringVar(&skillFlagKind, "kind", "", "Filter by kind (builtin, workspace)")
	skillInstallCmd.Flags().BoolVar(&skillFlagOverwrite, "overwrite", true, "Overwrite existing skill with same name")

	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillEnableCmd)
	skillCmd.AddCommand(skillDisableCmd)
	skillCmd.AddCommand(skillRemoveCmd)
	skillCmd.AddCommand(skillInstallCmd)

	RootCmd.AddCommand(skillCmd)
	RootCmd.AddCommand(listSkillsShortcutCmd)
}
