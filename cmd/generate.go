package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/khulnasoft/packages-registry/config"
	"github.com/khulnasoft/packages-registry/khulnasoft"
	"github.com/khulnasoft/packages-registry/logger"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates the pipeline config path.",
	Long: `Generates the pipeline config path.

	Use the pipeline_config flag to specify a file, otherwise "child_pipeline.yml" is used`,
	Run: func(cmd *cobra.Command, args []string) {
		if errorReadingConfig {
			return
		}

		configuration, err := config.Load()
		if err != nil {
			logger.LogError("Error while loading the config:", err)
			return
		}

		outputFile, err := outputFile()
		if err != nil {
			logger.LogError("Error while opening the output file:", err)
			return
		}

		generator := khulnasoft.NewGenerator(configuration)
		if err = generator.Generate(outputFile); err != nil {
			logger.LogError("Error while generating the engine config:", err)
			return
		}

		logger.LogSuccess("Pipeline config generated!")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func outputFile() (*os.File, error) {
	f, err := os.Create(pipelineConfigFilePath)
	if err != nil {
		return nil, err
	}
	logger.LogInfo(fmt.Sprintf("Writing pipeline config file %q", f.Name()))
	return f, nil
}
