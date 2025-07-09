package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"pseudonymous/config"
	"pseudonymous/fhir"
	"regexp"
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
			p, err := fhir.NewProcessor(cfg, projectName)
			if err != nil {
				return err
			}
			_, err = p.Run()
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

	parseMapEnvs()

	configMap := make(map[string]map[string]string)
	re := regexp.MustCompile(`(.*)\[[0-9]+]$`)
	for _, e := range os.Environ() {
		split := strings.Split(e, "=")
		k := split[0]
		v := strings.Split(split[1], ":")

		result := re.FindStringSubmatch(k)

		if result != nil {
			slog.Info(k, "result", result)
			key := result[1]

			var val map[string]string
			val, exists := configMap[key]
			if !exists {
				val = make(map[string]string)
				configMap[key] = val
			}
			val[v[0]] = v[1]
		}
	}

	// reverse (doesn't work for '-' though)
	replacer := strings.NewReplacer(`_`, `.`)
	for k, v := range configMap {
		viper.Set(replacer.Replace(k), v)
	}

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

func parseMapEnvs() {
	// reverse (doesn't work for '-' though)
	replacer := strings.NewReplacer(`_`, `.`)

	configMap := make(map[string]map[string]string)
	re := regexp.MustCompile(`(.*)\[[0-9]+]$`)
	for _, e := range os.Environ() {
		split := strings.Split(e, "=")
		k := split[0]
		v := strings.Split(split[1], ":")

		result := re.FindStringSubmatch(k)

		if result != nil {
			key := replacer.Replace(result[1])

			var val map[string]string
			val, exists := configMap[key]
			if !exists {
				val = make(map[string]string)
				configMap[key] = val
			}
			val[v[0]] = v[1]
		}
	}

	for k, v := range configMap {
		viper.Set(k, v)
	}
}

func validateCmd() error {
	if projectName == "" {
		return errors.New("project name is empty")
	}
	return nil
}
