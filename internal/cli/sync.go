package cli

import (
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:     "sync",
	Aliases: []string{"s"},
	Short:   "Sync subtitles to media files",
	Long: `Sync subtitles to the audio track of media files.
	
Use 'movies' or 'shows' subcommands to specify what to sync.`,
	Example: `  bazarr-sync sync movies
  bazarr-sync sync shows
  bazarr-sync sync movies --list`,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
