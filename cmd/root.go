// Package cmd host all the commands available and follows the [cobra](https://github.com/spf13/cobra) skeleton.
// Only one command is available: generate. The root command hosts the pieces that can be shared between available commands.
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/khulnasoft/packages-registry/logger"
)

var (
	configFilePath         string
	pipelineConfigFilePath string
	version                string
	buildTime              string
	commit                 string
	errorReadingConfig     bool
)

var rootCmd = &cobra.Command{
	Use:   "pkgs_importer",
	Short: "A tool to generate a configuration for engines that will import packages between two registries",
	Long: `This tool will use a configuration file that will describe imports. An import is a one way road between
two packages regisitries. A import configuration will contain the credentials to be used for each packages
registry and a description of the packages to copy.

From this, this tool will configure a KhulnaSoft dynamic child pipeline that will carry out the copy.

Supported package types:
- npm
- maven
- nuget
- pypi`,
	Version: fmt.Sprintf("%q, build time %q, commit %q", version, buildTime, commit),
}

const defaultPipelineConfigFilePath = "child_pipeline.yml"

// Execute will execute the command. Depending on the arguments, the generate command is executed or the help message is displayed.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "config.yml", "Configuration file path")
	rootCmd.PersistentFlags().StringVarP(&pipelineConfigFilePath, "pipeline_config", "p", defaultPipelineConfigFilePath, "Pipeline configuration file path")
}

func initConfig() {
	logger.LogInfo("Loading Config")
	viper.SetCaseSensitive() // To avoid https://github.com/spf13/viper#does-viper-support-case-sensitive-keys
	viper.SetConfigFile(configFilePath)
	if err := viper.ReadInConfig(); err != nil {
		logger.LogError("Error while reading the config:", err)
		errorReadingConfig = true
	}
}
