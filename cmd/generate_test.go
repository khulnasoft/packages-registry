package cmd

import (
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const testdataPipelineConfigPath = "../testdata/output/pipeline_config.yml"

func TestGenerate(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedOutputs []string
	}{
		{
			name:            "without arguments",
			args:            []string{"generate"},
			expectedOutputs: []string{"Error while reading the config", "open config.yml: no such file or directory"},
		},
		{
			name:            "non existing config file",
			args:            []string{"generate", "-c", "does_not_exist.yml"},
			expectedOutputs: []string{"Error while reading the config:", "open does_not_exist.yml: no such file or directory"},
		},
		{
			name:            "incoherent config file",
			args:            []string{"generate", "-c", "../testdata/incoherent.yml"},
			expectedOutputs: []string{"Error while loading the config", "expected a map, got 'string'"},
		},
		{
			name:            "single import",
			args:            []string{"generate", "-c", "../testdata/single_import.yml"},
			expectedOutputs: []string{"Config loaded", "Pipeline config generated!"},
		},
		{
			name:            "multiple imports",
			args:            []string{"generate", "-c", "../testdata/multiple_imports.yml"},
			expectedOutputs: []string{"Config loaded", `Writing pipeline config file "child_pipeline.yml"`, "Pipeline config generated!"},
		},
		{
			name:            "pipeline config path",
			args:            []string{"generate", "-c", "../testdata/single_import.yml", "-p", testdataPipelineConfigPath},
			expectedOutputs: []string{"Config loaded", `Writing pipeline config file "../testdata/output/pipeline_config.yml"`, "Pipeline config generated!"},
		},
		{
			name:            "unknown type",
			args:            []string{"generate", "-c", "../testdata/unknown_type.yml"},
			expectedOutputs: []string{"'Configuration.Imports[import1].Type' Error:Field validation"},
		},
		{
			name:            "same urls",
			args:            []string{"generate", "-c", "../testdata/same_urls.yml"},
			expectedOutputs: []string{`import "import1" has the same url for the source and the destination`},
		},
		{
			name:            "anymous source",
			args:            []string{"generate", "-c", "../testdata/anonymous_source.yml"},
			expectedOutputs: []string{"Config loaded", "Pipeline config generated!"},
		},
		{
			name:            "no credentials destination",
			args:            []string{"generate", "-c", "../testdata/no_credentials_destination.yml"},
			expectedOutputs: []string{`credentials token for destination in import "import1" is required`},
		},
		{
			name:            "no username for nuget import",
			args:            []string{"generate", "-c", "../testdata/nuget_no_username.yml"},
			expectedOutputs: []string{"NuGet credentials require a token and a username in authenticated registries"},
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			buff := new(strings.Builder)

			t.Cleanup(reset)
			log.SetOutput(buff)

			rootCmd.SetArgs(spec.args)
			err := rootCmd.Execute()
			require.Nil(t, err)

			output := buff.String()

			for _, expectedOutput := range spec.expectedOutputs {
				require.Contains(t, output, expectedOutput)
			}
		})
	}
}

const testdataExpectedExactMatchPipelineConfigPath = "../testdata/expected_exact_match_pipeline_config.yml"

func TestGenerateExactMatch(t *testing.T) {
	args := []string{"generate", "-c", "../testdata/exact_match.yml", "-p", testdataPipelineConfigPath}

	t.Cleanup(reset)
	log.SetOutput(io.Discard)

	rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	require.Nil(t, err)

	result, err := os.ReadFile(testdataPipelineConfigPath)
	require.Nil(t, err)

	expected, err := os.ReadFile(testdataExpectedExactMatchPipelineConfigPath)
	require.Nil(t, err)

	require.Equal(t, string(expected), string(result))
}

func reset() {
	viper.Reset()
	errorReadingConfig = false
	os.Remove(testdataPipelineConfigPath)
	log.SetOutput(os.Stderr)
	os.Remove(defaultPipelineConfigFilePath)
}
