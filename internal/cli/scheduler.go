package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pterm/pterm"
	"github.com/regix1/bazarr-sync/internal/config"
	"github.com/robfig/cron/v3"
)

func RunScheduler(cfg config.Config) {
	if !cfg.Schedule.Enabled {
		// Run once and exit
		runSyncJobs(cfg)
		return
	}

	// Load timezone
	location, err := time.LoadLocation(cfg.Schedule.Timezone)
	if err != nil {
		pterm.Error.Printf("Invalid timezone '%s': %v. Using UTC instead.\n", cfg.Schedule.Timezone, err)
		location = time.UTC
	}

	// Create cron scheduler with timezone
	c := cron.New(cron.WithLocation(location))

	// Add scheduled job
	_, err = c.AddFunc(cfg.Schedule.CronExpression, func() {
		runSyncJobs(cfg)
	})

	if err != nil {
		pterm.Error.Printf("Invalid cron expression '%s': %v\n", cfg.Schedule.CronExpression, err)
		os.Exit(1)
	}

	// Start scheduler
	c.Start()

	// Calculate and display next run time
	entries := c.Entries()
	if len(entries) > 0 {
		nextRun := entries[0].Next
		pterm.Info.Printf("Scheduler started. Next sync scheduled for: %s\n",
			nextRun.Format("2006-01-02 15:04:05 MST"))
		pterm.Info.Printf("Schedule: %s (Timezone: %s)\n",
			cfg.Schedule.CronExpression, cfg.Schedule.Timezone)
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run initial sync if requested
	if runInitial {
		pterm.Info.Println("Running initial sync...")
		runSyncJobs(cfg)
	}

	// Wait for interrupt signal
	<-sigChan
	pterm.Warning.Println("\nReceived interrupt signal. Shutting down scheduler...")
	c.Stop()
	pterm.Success.Println("Scheduler stopped gracefully.")
}

func runSyncJobs(cfg config.Config) {
	startTime := time.Now()
	fmt.Printf("\n%s Starting scheduled sync job\n",
		startTime.Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 60))

	// Create channels for progress tracking
	progressChan := make(chan int, 1)
	doneChan := make(chan bool)

	// Track progress
	go func() {
		for {
			select {
			case <-progressChan:
				// Progress update handled in sync functions
			case <-doneChan:
				return
			}
		}
	}()

	// Load cache if enabled
	if cfg.Cache.Enabled {
		Load_cache(cfg)
	}

	// Run sync jobs based on configuration
	if cfg.Schedule.SyncShows {
		fmt.Println("\n📺 Syncing TV shows...")
		sync_shows(cfg, progressChan)
	}

	if cfg.Schedule.SyncMovies {
		fmt.Println("\n🎬 Syncing movies...")
		sync_movies(cfg, progressChan)
	}

	close(doneChan)

	duration := time.Since(startTime)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("✅ Sync job completed in %s\n", duration.Round(time.Second))

	// If scheduled, show next run time
	if cfg.Schedule.Enabled {
		nextRun := calculateNextRun(cfg.Schedule.CronExpression, cfg.Schedule.Timezone)
		if !nextRun.IsZero() {
			fmt.Printf("⏰ Next sync scheduled for: %s\n\n",
				nextRun.Format("2006-01-02 15:04:05 MST"))
		}
	}
}

func calculateNextRun(cronExpr string, timezone string) time.Time {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		location = time.UTC
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}
	}

	return schedule.Next(time.Now().In(location))
}
