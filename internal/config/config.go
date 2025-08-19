package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Address     string
	Port        string
	Protocol    string
	ApiToken    string
	BazarrUrl   string
	ApiUrl      string
	Schedule    ScheduleConfig
	Cache       CacheConfig
	SyncOptions SyncOptionsConfig
}

type ScheduleConfig struct {
	Enabled        bool
	SyncMovies     bool
	SyncShows      bool
	CronExpression string
	Timezone       string
}

type CacheConfig struct {
	Enabled     bool
	MoviesCache string
	ShowsCache  string
}

type SyncOptionsConfig struct {
	GoldenSection  bool
	NoFramerateFix bool
}

var cfg Config
var CfgFile string

func GetConfig() Config {
	return cfg
}

func InitConfig() {
	if CfgFile != "" {
		viper.SetConfigFile(CfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/config")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("Schedule.Enabled", false)
	viper.SetDefault("Schedule.SyncMovies", true)
	viper.SetDefault("Schedule.SyncShows", true)
	viper.SetDefault("Schedule.CronExpression", "0 1 * * 0")
	viper.SetDefault("Schedule.Timezone", "UTC")
	viper.SetDefault("Cache.Enabled", false)
	viper.SetDefault("Cache.MoviesCache", "movies-cache")
	viper.SetDefault("Cache.ShowsCache", "shows-cache")
	viper.SetDefault("SyncOptions.GoldenSection", false)
	viper.SetDefault("SyncOptions.NoFramerateFix", false)

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Fprintln(os.Stderr, "Configuration Error:", err)
		fmt.Fprintln(os.Stderr, "Please supply a config.yaml file by using the flag --config or by placing the file in the same directory as bazarr-sync")
		os.Exit(1)
	}

	viper.Unmarshal(&cfg)

	var (
		baseUrl string
		err     error
	)

	if strings.Contains(cfg.Address, "/") {
		baseUrl, err = url.JoinPath(cfg.Protocol + "://" + cfg.Address)
	} else {
		baseUrl, err = url.JoinPath(cfg.Protocol + "://" + cfg.Address + ":" + cfg.Port)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "URL Error:", err)
	}

	apiUrl, _ := url.JoinPath(baseUrl, "api/")
	cfg.BazarrUrl = baseUrl
	cfg.ApiUrl = apiUrl
}
