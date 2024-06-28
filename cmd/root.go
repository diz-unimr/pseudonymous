package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"pseudonymous/config"
	"pseudonymous/fhir"
	"strings"
)

var (
	projectName string
	cfgFile     string
	cfg         *config.AppConfig
	rootCmd     = NewRootCmd()
)

func NewRootCmd() *cobra.Command {

	return &cobra.Command{
		Use:   "pseudonymous",
		Short: "Pseudonymization of FHIR resources via the FHIR Pseudonymizer service ",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := validateCmd(); err != nil {
				slog.Error("Failed to validate command flags", "error", err.Error())
				return err
			}

			config.ConfigureLogger(*cfg)
			p := fhir.NewProcessor(cfg, projectName)
			_, err := p.Run()
			if err != nil {
				slog.Error("Processor run exited", "error", err.Error())
			}
			return err
		},
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Execution failed", "error", err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&projectName, "project", "p", "", "project name (required)")
	if err := rootCmd.MarkPersistentFlagRequired("project"); err != nil {
		slog.Error("Please provide a project name with the -p flag")
		os.Exit(1)
	}

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./app.yaml)")
}

func initConfig() {

	if cfgFile != "" {
		// use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("./")
		viper.SetConfigName("app")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`, `-`, `_`))

	if err := viper.ReadInConfig(); err == nil {
		slog.Info("Using config file", "file", viper.ConfigFileUsed())
	} else {
		slog.Error("Error reading config", "error", err.Error())
		os.Exit(1)
	}

	err := viper.Unmarshal(&cfg)
	if err != nil {
		slog.Error("Error unmarshalling app config", "error", err.Error())
		os.Exit(1)
	}
}

func validateCmd() error {
	if projectName == "" {
		return errors.New("project name is empty")
	}
	return nil
}
