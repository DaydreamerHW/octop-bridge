package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DaydreamerHW/octop-bridge/pkg/client"
	"github.com/spf13/cobra"
)

var (
	chatFlagThreadID  string
	chatFlagNewThread bool
	chatFlagMode      string
	chatFlagNoStream  bool
)

var chatCmd = &cobra.Command{
	Use:   "chat <query>",
	Short: "Send a message or delegated task to an Octop agent",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChat(args, chatFlagNewThread, chatFlagThreadID, chatFlagMode, chatFlagNoStream)
	},
}

func runChat(args []string, newThread bool, threadID string, mode string, noStream bool) error {
	query := strings.Join(args, " ")
	cfg := GetConfig()
	octopClient := client.NewClient(cfg)

	stream := !noStream && !flagJSON

	onToken := func(content string) {
		if stream {
			fmt.Print(content)
			_ = os.Stdout.Sync()
		}
	}

	onToolStart := func(toolName string) {
		if stream {
			fmt.Printf("\n[小通正在调用工具: %s]...\n", toolName)
			_ = os.Stdout.Sync()
		}
	}

	onToolEnd := func(toolName string) {
		if stream {
			fmt.Printf("[工具调用完成: %s]\n", toolName)
			_ = os.Stdout.Sync()
		}
	}

	timeoutSec := cfg.Timeout
	if timeoutSec <= 0 {
		timeoutSec = 180
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	res, err := octopClient.ChatStream(
		ctx,
		query,
		cfg.DefaultAgentID,
		threadID,
		newThread,
		mode,
		onToken,
		onToolStart,
		onToolEnd,
	)
	if err != nil {
		if stream {
			fmt.Println()
		}
		return err
	}

	if stream {
		fmt.Println()
	}

	if flagJSON {
		data, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(data))
	} else if !stream {
		fmt.Println(res.Response)
	}

	return nil
}

func init() {
	chatCmd.Flags().StringVarP(&chatFlagThreadID, "thread-id", "t", "", "Conversation thread ID to continue")
	chatCmd.Flags().BoolVarP(&chatFlagNewThread, "new", "n", false, "Start a new isolated conversation thread")
	chatCmd.Flags().StringVarP(&chatFlagMode, "mode", "m", "craft", "Turn mode (craft, ask, plan)")
	chatCmd.Flags().BoolVar(&chatFlagNoStream, "no-stream", false, "Disable live streaming output")
	RootCmd.AddCommand(chatCmd)
}
