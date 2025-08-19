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

var radarrid []int
var moviesContinueFrom int

var moviesCmd = &cobra.Command{
	Use:     "movies",
	Short:   "Sync subtitles to the audio track of the movie",
	Example: "bazarr-sync --config config.yaml sync movies --no-framerate-fix",
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
			list_movies(cfg)
			return
		}

		runWithSignalHandler(func(c chan int) {
			sync_movies(cfg, c)
		})
	},
}

func init() {
	syncCmd.AddCommand(moviesCmd)
	moviesCmd.Flags().IntSliceVar(&radarrid, "radarr-id", []int{}, "Specify a list of radarr Ids to sync. Use --list to view your movies with respective radarr id.")
	moviesCmd.Flags().IntVar(&moviesContinueFrom, "continue-from", -1, "Continue with the given Radarr movie ID.")
}

func sync_movies(cfg config.Config, c chan int) {
	movies, err := bazarr.QueryMovies(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Query Error: Could not query movies")
		return
	}

	totalMovies := len(movies.Data)
	fmt.Printf("Found %d movies in your Bazarr library.\n", totalMovies)
	fmt.Println("Starting sync process...")
	fmt.Println(strings.Repeat("-", 60))

	skipForward := moviesContinueFrom != -1
	successCount := 0
	skipCount := 0
	failCount := 0

movies:
	for i, movie := range movies.Data {
		specified_id := false
		if len(radarrid) > 0 {
			for _, id := range radarrid {
				if id == movie.RadarrId {
					specified_id = true
					goto subtitle
				}
			}
			continue movies
		}

		if skipForward {
			if movie.RadarrId == moviesContinueFrom {
				skipForward = false
			} else {
				fmt.Printf("[%d/%d] SKIPPING: %s (continue mode)\n", i+1, totalMovies, movie.Title)
				skipCount++
				continue
			}
		}

	subtitle:
		c <- movie.RadarrId

		if len(movie.Subtitles) == 0 {
			fmt.Printf("[%d/%d] NO SUBS: %s\n", i+1, totalMovies, movie.Title)
			continue
		}

		fmt.Printf("[%d/%d] PROCESSING: %s (%d subtitles)\n", i+1, totalMovies, movie.Title, len(movie.Subtitles))

		for _, subtitle := range movie.Subtitles {
			if subtitle.Path == "" || subtitle.File_size == 0 {
				fmt.Printf("  └─ SKIP [%s]: Embedded or missing subtitle\n", subtitle.Code2)
				skipCount++
				continue
			}

			if !specified_id && cfg.Cache.Enabled {
				_, exists := movies_cache[subtitle.Path]
				if exists {
					fmt.Printf("  └─ CACHED [%s]: Already synced\n", subtitle.Code2)
					skipCount++
					continue
				}
			}

			params := bazarr.GetSyncParams("movie", movie.RadarrId, subtitle)
			if cfg.SyncOptions.GoldenSection {
				params.Gss = "True"
			}
			if cfg.SyncOptions.NoFramerateFix {
				params.No_framerate_fix = "True"
			}

			fmt.Printf("  └─ SYNCING [%s]: ", subtitle.Code2)
			ok := bazarr.Sync(cfg, params)

			if ok {
				fmt.Printf("✓ Success\n")
				Write_movies_cache(cfg, subtitle.Path)
				successCount++
			} else {
				fmt.Printf("✗ Failed, retrying...")
				time.Sleep(2 * time.Second)
				ok := bazarr.Sync(cfg, params)
				if ok {
					fmt.Printf(" ✓ Success\n")
					Write_movies_cache(cfg, subtitle.Path)
					successCount++
				} else {
					fmt.Printf(" ✗ Failed\n")
					failCount++
				}
			}
		}
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Sync completed: %d successful, %d skipped, %d failed\n",
		successCount, skipCount, failCount)
	close(c)
}

func list_movies(cfg config.Config) {
	movies, err := bazarr.QueryMovies(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Query Error: Could not query movies")
		return
	}

	fmt.Printf("%-60s %s\n", "Title", "RadarrId")
	fmt.Println(strings.Repeat("-", 70))

	for _, movie := range movies.Data {
		fmt.Printf("%-60s %d\n", movie.Title, movie.RadarrId)
	}

	fmt.Printf("\nTotal: %d movies\n", len(movies.Data))
}
