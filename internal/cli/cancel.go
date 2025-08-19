package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var cancelCmd = &cobra.Command{
	Use:     "cancel",
	Aliases: []string{"stop", "c"},
	Short:   "Cancel any running sync operations",
	Example: "  bazarr-sync cancel",
	Run: func(cmd *cobra.Command, args []string) {
		// Find all running sync processes
		out, err := exec.Command("sh", "-c", "ps aux | grep 'bazarr-sync' | grep -E '(sync|movies|shows)' | grep -v grep").Output()
		if err != nil {
			fmt.Println("❌ No sync operations currently running")
			return
		}

		lines := strings.Split(string(out), "\n")
		cancelled := false

		for _, line := range lines {
			if line == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) > 1 {
				pid := fields[1]

				// Send SIGTERM to process
				err := exec.Command("kill", "-TERM", pid).Run()
				if err == nil {
					fmt.Printf("🛑 Sent cancel signal to sync process (PID: %s)\n", pid)
					cancelled = true
				}
			}
		}

		if cancelled {
			fmt.Println("✅ Cancel signal sent. The sync will stop gracefully.")
		} else {
			fmt.Println("❌ No sync operations found to cancel")
		}
	},
}

func init() {
	rootCmd.AddCommand(cancelCmd)
}
