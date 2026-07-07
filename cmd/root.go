package cmd

import (
	"fmt"
	"os"

	"github.com/lmorchard/bsh-spy-go/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	log     = logrus.New()
	cfg     *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bsh-spy-go",
	Short: "Scrape a radio station's now-playing feed into a Spotify playlist",
	Long: `bsh-spy-go polls a radio station's now-playing feed (e.g. a Streemlion
JSON endpoint), matches tracks against Spotify, and appends newly played
songs to a target Spotify playlist. It can run as a one-shot check or as a
long-running polling daemon.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initConfig()
		setupLogging()
	},
}

// Execute adds all child commands to the root command and sets appropriate flags.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Configuration file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./bsh-spy-go.yaml)")

	// Logging flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "debug output")
	rootCmd.PersistentFlags().Bool("log-json", false, "output logs in JSON format")

	// Database flag
	rootCmd.PersistentFlags().String("database", "bsh-spy-go.db", "database file path")

	// Bind flags to viper
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindPFlag("log_json", rootCmd.PersistentFlags().Lookup("log-json"))
	_ = viper.BindPFlag("database", rootCmd.PersistentFlags().Lookup("database"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("bsh-spy-go")
	}

	// Set defaults
	config.SetDefaults()

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		if cfgFile != "" {
			// Only error if config was explicitly specified
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
			os.Exit(1)
		}
	}
}

// setupLogging configures the logger based on configuration
func setupLogging() {
	if viper.GetBool("log_json") {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	log.SetLevel(resolveLogLevel(viper.GetBool("debug"), viper.GetBool("verbose"), viper.GetString("log_level")))
}

// resolveLogLevel picks the logrus level: --debug and --verbose take
// precedence (debug > verbose), otherwise the configured log_level string
// (default "info" via config.SetDefaults), falling back to Info if unparseable.
func resolveLogLevel(debug, verbose bool, logLevel string) logrus.Level {
	if debug {
		return logrus.DebugLevel
	}
	if verbose {
		return logrus.InfoLevel
	}
	if lvl, err := logrus.ParseLevel(logLevel); err == nil {
		return lvl
	}
	return logrus.InfoLevel
}

// GetConfig returns the application configuration, loading it if necessary
func GetConfig() *config.Config {
	if cfg == nil {
		cfg = config.Load()
	}
	return cfg
}

// GetLogger returns the configured logger
func GetLogger() *logrus.Logger {
	return log
}
