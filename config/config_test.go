package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name                 string
		configFixture        string
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:                 "with incoherent file",
			configFixture:        "incoherent.yml",
			expectError:          true,
			expectedErrorMessage: "1 error(s) decoding",
		},
		{
			name:          "with multi imports",
			configFixture: "multiple_imports.yml",
			expectError:   false,
		},
		{
			name:                 "with same urls",
			configFixture:        "same_urls.yml",
			expectError:          true,
			expectedErrorMessage: `import "import1" has the same url for the source and the destination`,
		},
		{
			name:          "with single import",
			configFixture: "single_import.yml",
			expectError:   false,
		},
		{
			name:                 "with unknown type",
			configFixture:        "unknown_type.yml",
			expectError:          true,
			expectedErrorMessage: "Key: 'Configuration.Imports[import1].Type' Error:Field validation for 'Type' failed on the 'oneof' tag",
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			file, err := os.Open(fmt.Sprintf("../testdata/%s", spec.configFixture))
			if err != nil {
				t.Logf("Can't find fixture file %q", spec.configFixture)
				t.Fail()
			}

			viper.SetConfigType("yml")
			err = viper.ReadConfig(file)

			require.NoError(t, err)

			config, err := Load()

			if spec.expectError {
				require.Error(t, err)
				require.Nil(t, config)
				require.Contains(t, err.Error(), spec.expectedErrorMessage)
			} else {
				require.Nil(t, err)
				require.NotNil(t, config)
			}
		})
	}
}
