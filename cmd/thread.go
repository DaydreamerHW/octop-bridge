package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DaydreamerHW/octop-bridge/pkg/client"
	"github.com/spf13/cobra"
)

var threadCmd = &cobra.Command{
	Use:     "thread",
	Aliases: []string{"threads"},
	Short:   "Manage and inspect Octop conversation threads",
}

var threadListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List conversation threads for an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		agent, err := octopClient.ResolveAgent(cfg.DefaultAgentID)
		if err != nil {
			return err
		}

		threads, err := octopClient.ListThreads(agent.AgentID)
		if err != nil {
			return err
		}

		if flagJSON {
			data, _ := json.MarshalIndent(threads, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("\nThreads for Agent: %s (%s)\n", agent.Name, agent.AgentID)
		fmt.Printf("Total: %d thread(s):\n\n", len(threads))
		fmt.Printf("%-32s | %s\n", "THREAD ID", "PREVIEW")
		fmt.Println(strings.Repeat("-", 75))
		for _, t := range threads {
			preview := t.Preview
			if preview == "" {
				preview = t.Title
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			if len(preview) > 38 {
				preview = preview[:35] + "..."
			}
			fmt.Printf("%-32s | %s\n", t.ThreadID, preview)
		}
		fmt.Println()
		return nil
	},
}

var threadNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a fresh, clean conversation thread",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfig()
		octopClient := client.NewClient(cfg)

		agent, err := octopClient.ResolveAgent(cfg.DefaultAgentID)
		if err != nil {
			return err
		}

		tid, err := octopClient.CreateThread(agent.AgentID)
		if err != nil {
			return err
		}

		if flagJSON {
			res := map[string]string{
				"status":    "created",
				"agent_id":  agent.AgentID,
				"thread_id": tid,
			}
			data, _ := json.MarshalIndent(res, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("✓ Created new thread: %s for agent %s (%s)\n", tid, agent.Name, agent.AgentID)
		return nil
	},
}

// Top-level shortcut alias for list-threads
var listThreadsShortcutCmd = &cobra.Command{
	Use:     "list-threads",
	Aliases: []string{"threads"},
	Short:   "Shortcut to list conversation threads",
	RunE:    threadListCmd.RunE,
}

func init() {
	threadCmd.AddCommand(threadListCmd)
	threadCmd.AddCommand(threadNewCmd)
	RootCmd.AddCommand(threadCmd)
	RootCmd.AddCommand(listThreadsShortcutCmd)
}
