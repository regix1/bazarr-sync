package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/regix1/bazarr-sync/internal/bazarr"
	"github.com/regix1/bazarr-sync/internal/config"
	"github.com/spf13/cobra"
)

var sonarrid []int
var showsContinueFrom int

var showsCmd = &cobra.Command{
	Use:     "shows",
	Aliases: []string{"show", "tv", "series"},
	Short:   "Sync subtitles to the audio track of TV shows",
	Example: `  bazarr-sync sync shows
  bazarr-sync sync shows --list
  bazarr-sync sync shows --sonarr-id 123,456`,
	Long: `By default, Bazarr will try to sync the sub to the audio track:0 of the media. 
This can fail due to many reasons mainly due to failure of bazarr to extract audio info. This is unfortunately out of my hands.
The script by default will try to not use the golden section search method and will try to fix framerate issues. This can be changed using the flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()

		// Override config with command line flags
		if cmd.Flags().Changed("golden-section") {
			cfg.SyncOptions.GoldenSection = gss
		}
		if cmd.Flags().Changed("no-framerate-fix") {
			cfg.SyncOptions.NoFramerateFix = no_framerate_fix
		}
		if cmd.Flags().Changed("use-cache") {
			cfg.Cache.Enabled = use_cache
		}

		if cfg.Cache.Enabled {
			Load_cache(cfg)
		}

		bazarr.HealthCheck(cfg)

		if to_list {
			list_shows(cfg)
			return
		}

		runWithSignalHandler(func(c chan int) {
			sync_shows(cfg, c)
		})
	},
}

func init() {
	syncCmd.AddCommand(showsCmd)
	showsCmd.Flags().IntSliceVar(&sonarrid, "sonarr-id", []int{}, "Specify a list of sonarr Ids to sync. Use --list to view your shows with respective sonarr id.")
	showsCmd.Flags().IntVar(&showsContinueFrom, "continue-from", -1, "Continue with the given Sonarr episode ID.")
}

func sync_shows(cfg config.Config, c chan int) {
	shows, err := bazarr.QuerySeries(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Query Error: Could not query series")
		return
	}

	totalShows := len(shows.Data)
	fmt.Printf("Found %d shows in your Bazarr library.\n", totalShows)
	fmt.Println("Starting sync process...")
	fmt.Println(strings.Repeat("-", 60))

	skipForward := showsContinueFrom != -1
	successCount := 0
	skipCount := 0
	failCount := 0

shows:
	for i, show := range shows.Data {
		if len(sonarrid) > 0 {
			found := false
			for _, id := range sonarrid {
				if id == show.SonarrSeriesId {
					found = true
					break
				}
			}
			if !found {
				continue shows
			}
		}

		episodes, err := bazarr.QueryEpisodes(cfg, show.SonarrSeriesId)
		if err != nil {
			fmt.Printf("[%d/%d] ERROR: %s - Could not query episodes\n", i+1, totalShows, show.Title)
			continue
		}

		if len(episodes.Data) == 0 {
			fmt.Printf("[%d/%d] NO EPISODES: %s\n", i+1, totalShows, show.Title)
			continue
		}

		fmt.Printf("[%d/%d] PROCESSING: %s (%d episodes)\n", i+1, totalShows, show.Title, len(episodes.Data))

		for _, episode := range episodes.Data {
			for _, subtitle := range episode.Subtitles {
				if skipForward {
					if episode.SonarrEpisodeId == showsContinueFrom {
						skipForward = false
					} else {
						skipCount++
						continue
					}
				}

				c <- episode.SonarrEpisodeId

				if subtitle.Path == "" || subtitle.File_size == 0 {
					fmt.Printf("  └─ SKIP [%s - %s]: Embedded or missing\n", episode.Title, subtitle.Code2)
					skipCount++
					continue
				}

				if cfg.Cache.Enabled {
					_, exists := shows_cache[subtitle.Path]
					if exists {
						fmt.Printf("  └─ CACHED [%s - %s]: Already synced\n", episode.Title, subtitle.Code2)
						skipCount++
						continue
					}
				}

				params := bazarr.GetSyncParams("episode", episode.SonarrEpisodeId, subtitle)
				if cfg.SyncOptions.GoldenSection {
					params.Gss = "True"
				}
				if cfg.SyncOptions.NoFramerateFix {
					params.No_framerate_fix = "True"
				}

				fmt.Printf("  └─ SYNCING [%s - %s]: ", episode.Title, subtitle.Code2)
				ok := bazarr.Sync(cfg, params)

				if ok {
					fmt.Printf("✓ Success\n")
					Write_shows_cache(cfg, subtitle.Path)
					successCount++
				} else {
					fmt.Printf("✗ Failed, retrying...")
					time.Sleep(2 * time.Second)
					ok := bazarr.Sync(cfg, params)
					if ok {
						fmt.Printf(" ✓ Success\n")
						Write_shows_cache(cfg, subtitle.Path)
						successCount++
					} else {
						fmt.Printf(" ✗ Failed\n")
						failCount++
					}
				}

				// Add delay between syncs to avoid overwhelming Bazarr
				time.Sleep(1 * time.Second)
			}
		}
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Sync completed: %d successful, %d skipped, %d failed\n",
		successCount, skipCount, failCount)
	close(c)
}

func list_shows(cfg config.Config) {
	shows, err := bazarr.QuerySeries(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Query Error: Could not query shows")
		return
	}

	fmt.Printf("%-60s %s\n", "Title", "SonarrSeriesId")
	fmt.Println(strings.Repeat("-", 70))

	for _, show := range shows.Data {
		fmt.Printf("%-60s %d\n", show.Title, show.SonarrSeriesId)
	}

	fmt.Printf("\nTotal: %d shows\n", len(shows.Data))
}
