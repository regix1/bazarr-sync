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
var verbose bool

var moviesCmd = &cobra.Command{
	Use:     "movies",
	Aliases: []string{"movie", "m"},
	Short:   "Sync subtitles to the audio track of movies",
	Example: `  bazarr-sync sync movies
  bazarr-sync sync movies --list
  bazarr-sync sync movies --radarr-id 123,456`,
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
		if cmd.Flags().Changed("verbose") {
			verbose = true
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
	moviesCmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed error messages")
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
	alreadySyncedCount := 0

movies:
	for i, movie := range movies.Data {
		if len(radarrid) > 0 {
			found := false
			for _, id := range radarrid {
				if id == movie.RadarrId {
					found = true
					break
				}
			}
			if !found {
				continue movies
			}
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

			if cfg.Cache.Enabled {
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
			ok, message := bazarr.Sync(cfg, params)

			if ok {
				fmt.Printf("✓ Success\n")
				Write_movies_cache(cfg, subtitle.Path)
				successCount++
			} else {
				// Check if it's already synced
				if strings.Contains(strings.ToLower(message), "already") ||
					strings.Contains(strings.ToLower(message), "sync") ||
					strings.Contains(message, "304") ||
					strings.Contains(message, "409") {
					fmt.Printf("✓ Already in sync\n")
					Write_movies_cache(cfg, subtitle.Path) // Cache it so we don't try again
					alreadySyncedCount++
				} else {
					// Retry once for real failures
					fmt.Printf("✗ Failed (%s), retrying...", message)
					time.Sleep(2 * time.Second)
					ok, message := bazarr.Sync(cfg, params)
					if ok {
						fmt.Printf(" ✓ Success\n")
						Write_movies_cache(cfg, subtitle.Path)
						successCount++
					} else if strings.Contains(strings.ToLower(message), "already") ||
						strings.Contains(strings.ToLower(message), "sync") {
						fmt.Printf(" ✓ Already in sync\n")
						Write_movies_cache(cfg, subtitle.Path)
						alreadySyncedCount++
					} else {
						if verbose {
							fmt.Printf(" ✗ Failed: %s\n", message)
						} else {
							fmt.Printf(" ✗ Failed\n")
						}
						failCount++
					}
				}
			}

			// Add delay between syncs to avoid overwhelming Bazarr
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Sync completed:\n")
	fmt.Printf("  ✅ %d newly synced\n", successCount)
	fmt.Printf("  ✓  %d already in sync\n", alreadySyncedCount)
	fmt.Printf("  ⏭️  %d skipped (cached/embedded)\n", skipCount)
	fmt.Printf("  ❌ %d failed\n", failCount)

	if failCount > 0 && !verbose {
		fmt.Println("\n💡 Tip: Run with --verbose to see detailed error messages")
	}

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
