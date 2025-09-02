package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/craftslab/cleansource-sca-cli/internal/app"
	"github.com/craftslab/cleansource-sca-cli/internal/config"
	"github.com/craftslab/cleansource-sca-cli/internal/logger"
)

var (
	// Global configuration
	cfg *config.ScanConfig

	// Root command
	rootCmd = &cobra.Command{
		Use:     "cleansource-sca-cli",
		Short:   "CleanSource SCA build scanner",
		Long:    `CleanSource SCA build scanner - Version 4.0.0`,
		Version: "4.0.0",
		Run:     runScan,
	}
)

func init() {
	cfg = config.NewScanConfig()
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&cfg.ServerURL, "server-url", "", "Server URL")
	rootCmd.PersistentFlags().StringVar(&cfg.Username, "username", "", "Username for authentication")
	rootCmd.PersistentFlags().StringVar(&cfg.Password, "password", "", "Password for authentication")
	rootCmd.PersistentFlags().StringVar(&cfg.Token, "token", "", "Authentication token")

	// Scan flags
	rootCmd.Flags().StringVar(&cfg.TaskDir, "task-dir", "", "Task directory to scan")
	rootCmd.Flags().StringVar(&cfg.ScanType, "scan-type", "source", "Scan type (source, docker, binary)")
	rootCmd.Flags().StringVar(&cfg.TaskType, "task-type", "scan", "Task type")
	rootCmd.Flags().StringVar(&cfg.ToPath, "to-path", "", "Output directory path")
	rootCmd.Flags().BoolVar(&cfg.BuildDepend, "build-depend", true, "Build dependency tree")
	rootCmd.Flags().StringVar(&cfg.CustomProject, "custom-project", "", "Custom project name")
	rootCmd.Flags().StringVar(&cfg.CustomProduct, "custom-product", "", "Custom product name")
	rootCmd.Flags().StringVar(&cfg.CustomVersion, "custom-version", "", "Custom version")
	rootCmd.Flags().StringVar(&cfg.LicenseName, "license-name", "", "License name")
	rootCmd.Flags().StringVar(&cfg.NotificationEmail, "notification-email", "", "Notification email")
	rootCmd.Flags().StringVar(&cfg.ThreadNum, "thread-num", "30", "Thread number (1-60)")

	// Build tool specific flags
	rootCmd.Flags().StringVar(&cfg.MavenPath, "maven-path", "", "Maven executable path")
	rootCmd.Flags().StringVar(&cfg.MavenBuildCommand, "maven-build-command", "", "Maven build command")
	rootCmd.Flags().StringVar(&cfg.PipPath, "pip-path", "", "Pip executable path")
	rootCmd.Flags().StringVar(&cfg.PipRequirementsPath, "pip-requirements-path", "", "Pip requirements file path")
}

func initConfig() {
	if cfg == nil {
		cfg = config.NewScanConfig()
	}
}

func runScan(cmd *cobra.Command, args []string) {
	// Initialize logger
	logger.InitLogger(cfg.LogLevel)
	log := logger.GetLogger()

	log.Info("-----        Detect Version CleanSource_SCA: 4.0.0        -----")
	log.Info("-------------START OF SCAN------------")

	// Print parameters
	printParamLog(cfg)

	// Create and run application
	application := app.NewBuildScanApplication(cfg)
	if err := application.Run(); err != nil {
		log.Errorf("Scan failed: %v", err)
		os.Exit(1)
	}

	log.Info("------------- END OF SCAN ------------")
}

func printParamLog(cfg *config.ScanConfig) {
	log := logger.GetLogger()
	log.Infof("Task Directory: %s", cfg.TaskDir)
	log.Infof("Scan Type: %s", cfg.ScanType)
	log.Infof("Build Depend: %t", cfg.BuildDepend)
	if cfg.CustomProject != "" {
		log.Infof("Custom Project: %s", cfg.CustomProject)
	}
	if cfg.CustomProduct != "" {
		log.Infof("Custom Product: %s", cfg.CustomProduct)
	}
	if cfg.CustomVersion != "" {
		log.Infof("Custom Version: %s", cfg.CustomVersion)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}
