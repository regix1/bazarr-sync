package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/regix1/bazarr-sync/internal/config"
	"github.com/spf13/cobra"
)

var gss bool
var no_framerate_fix bool
var to_list bool
var use_cache bool
var runInitial bool
var schedule bool
var movies_cache = make(map[string]bool)
var shows_cache = make(map[string]bool)

var rootCmd = &cobra.Command{
	Use:     "bazarr-sync",
	Aliases: []string{"bs"},
	Short:   "Bulk-sync subtitles downloaded via Bazarr",
	Long: `Bulk-sync subtitles downloaded via Bazarr.
	
Bazarr lets you download subs for your titles automatically. 
But if for some reason you needed to sync old subtitles, you will be forced 
to do it one by one as there is no option to bulk sync them.
This cli tool helps you achieve that by utilizing bazarr's api.`,
	Example: `  bazarr-sync movies
  bazarr-sync shows
  bazarr-sync cancel
  bazarr-sync --schedule`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()

		// Override config with command line flags if provided
		if cmd.Flags().Changed("golden-section") {
			cfg.SyncOptions.GoldenSection = gss
		}
		if cmd.Flags().Changed("no-framerate-fix") {
			cfg.SyncOptions.NoFramerateFix = no_framerate_fix
		}
		if cmd.Flags().Changed("use-cache") {
			cfg.Cache.Enabled = use_cache
		}

		// Load cache if enabled
		if cfg.Cache.Enabled {
			Load_cache(cfg)
		}

		// Run scheduler if enabled in config or via flag
		if schedule || cfg.Schedule.Enabled {
			RunScheduler(cfg)
		} else {
			// Show help if no subcommand
			cmd.Help()
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		config.InitConfig()
	})

	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&gss, "golden-section", false, "Use Golden-Section Search")
	rootCmd.PersistentFlags().BoolVar(&no_framerate_fix, "no-framerate-fix", false, "Don't try to fix framerate")
	rootCmd.PersistentFlags().BoolVar(&to_list, "list", false, "List your media with their respective Radarr/Sonarr id")
	rootCmd.PersistentFlags().BoolVar(&use_cache, "use-cache", false, "Use cache to skip already synced subtitles")
	rootCmd.PersistentFlags().BoolVar(&schedule, "schedule", false, "Run on schedule defined in config file")
	rootCmd.PersistentFlags().BoolVar(&runInitial, "run-initial", false, "Run initial sync when starting scheduler")
}

func Load_cache(cfg config.Config) {
	if !cfg.Cache.Enabled {
		return
	}

	movies_cache_file, err := os.Open(cfg.Cache.MoviesCache)
	if err != nil {
		// File doesn't exist yet, that's okay
		if !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Error opening movies cache file:", err)
		}
		return
	}
	defer movies_cache_file.Close()

	scanner := bufio.NewScanner(movies_cache_file)
	for scanner.Scan() {
		movies_cache[scanner.Text()] = true
	}

	shows_cache_file, err := os.Open(cfg.Cache.ShowsCache)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Error opening shows cache file:", err)
		}
		return
	}
	defer shows_cache_file.Close()

	scanner = bufio.NewScanner(shows_cache_file)
	for scanner.Scan() {
		shows_cache[scanner.Text()] = true
	}
}

func Write_movies_cache(cfg config.Config, key string) {
	if !cfg.Cache.Enabled {
		return
	}
	movies_cache[key] = true
	file, err := os.Create(cfg.Cache.MoviesCache)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing movies cache file:", err)
		return
	}
	defer file.Close()
	for key := range movies_cache {
		fmt.Fprintln(file, key)
	}
}

func Write_shows_cache(cfg config.Config, key string) {
	if !cfg.Cache.Enabled {
		return
	}
	shows_cache[key] = true
	file, err := os.Create(cfg.Cache.ShowsCache)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing shows cache file:", err)
		return
	}
	defer file.Close()
	for key := range shows_cache {
		fmt.Fprintln(file, key)
	}
}

func runWithSignalHandler(syncFunc func(chan int)) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	progressChan := make(chan int, 1)
	go syncFunc(progressChan)

	lastSubtitleId := -1
	for {
		select {
		case subtitle, more := <-progressChan:
			if !more {
				return
			}
			lastSubtitleId = subtitle
		case <-sigChan:
			if lastSubtitleId != -1 {
				showContinueMessage(lastSubtitleId)
			} else {
				fmt.Println("Stopping current sync. No subtitles have been processed yet.")
			}
			return
		}
	}
}

func showContinueMessage(lastSubtitleId int) {
	fmt.Println("\nStopping current sync. To continue from this point the next time, run:")
	commandName := os.Args[0]
	var args []string
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--continue-from" {
			i++
			continue
		}
		args = append(args, arg)
	}
	arguments := strings.Join(args, " ")
	fmt.Printf("  %s %s --continue-from %d\n", commandName, arguments, lastSubtitleId)
}
