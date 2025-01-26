package npm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/khulnasoft/packages-registry/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScripts(t *testing.T) {
	tests := []struct {
		name                        string
		sourceUrl                   string
		sourceToken                 string
		sourceAdditionalParams      map[string]string
		destinationUrl              string
		destinationToken            string
		destinationAdditionalParams map[string]string
	}{
		{
			name:             "with no additional config",
			sourceUrl:        "http://source.test",
			sourceToken:      "TOKEN_FOR_SOURCE",
			destinationUrl:   "https://destination.test",
			destinationToken: "TOKEN_FOR_DESTINATION",
		},
		{
			name:                        "with additional config",
			sourceUrl:                   "http://source.test",
			sourceToken:                 "TOKEN_FOR_SOURCE",
			sourceAdditionalParams:      map[string]string{"test": "foo", "bar": "123"},
			destinationUrl:              "https://destination.test",
			destinationToken:            "TOKEN_FOR_DESTINATION",
			destinationAdditionalParams: map[string]string{"test": "foo", "bar": "123"},
		},
		{
			name:                        "with base 64 token",
			sourceUrl:                   "http://source.test",
			sourceToken:                 "TOKEN_FOR_SOURCE",
			sourceAdditionalParams:      map[string]string{"_base64_token": "1"},
			destinationUrl:              "https://destination.test",
			destinationToken:            "TOKEN_FOR_DESTINATION",
			destinationAdditionalParams: map[string]string{"_base64_token": "1"},
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			pkgsImport := config.Import{
				Source: config.Registry{
					URL: spec.sourceUrl,
					Credentials: config.Credentials{
						Token:                spec.sourceToken,
						AdditionalParameters: spec.sourceAdditionalParams,
					},
				},
				Destination: config.Registry{
					URL: spec.destinationUrl,
					Credentials: config.Credentials{
						Token:                spec.destinationToken,
						AdditionalParameters: spec.destinationAdditionalParams,
					},
				},
			}

			registry := Registry{
				pkgsImport: pkgsImport,
			}

			scripts, err := registry.Scripts()
			assert.NoError(t, err)
			joinedScripts := strings.Join(scripts, "\n")

			assertRegistryAccess(t, joinedScripts, spec.sourceUrl, spec.sourceToken, spec.sourceAdditionalParams)
			assertRegistryAccess(t, joinedScripts, spec.destinationUrl, spec.destinationToken, spec.destinationAdditionalParams)

			if len(spec.sourceAdditionalParams) != 0 {
				assertAdditionalParams(t, joinedScripts, spec.sourceAdditionalParams)
			}

			if len(spec.destinationAdditionalParams) != 0 {
				assertAdditionalParams(t, joinedScripts, spec.destinationAdditionalParams)
			}
		})
	}
}

func assertRegistryAccess(t *testing.T, scripts string, registryUrl string, token string, params map[string]string) {
	require.Contains(t, scripts, fmt.Sprintf("registry = %s", registryUrl))

	url := strings.TrimPrefix(registryUrl, "https:")
	url = strings.TrimPrefix(url, "http:")
	if params["_base64_token"] == "1" {
		require.Contains(t, scripts, fmt.Sprintf("%s:_auth = %s", url, token))
	} else {
		require.Contains(t, scripts, fmt.Sprintf("%s:_authToken = %s", url, token))
	}
}

func assertAdditionalParams(t *testing.T, scripts string, params map[string]string) {
	for k, v := range params {
		if k != "_base64_token" {
			require.Contains(t, scripts, fmt.Sprintf("%s = %s", k, v))
		}
	}
}

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name        string
		importType  string
		expectError bool
	}{
		{
			name:        "with wrong import type",
			importType:  "unknown",
			expectError: true,
		},
		{
			name:        "with correct import type",
			importType:  "npm",
			expectError: false,
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			pkgsImport := config.Import{
				Type: spec.importType,
			}

			registry, err := NewRegistry(pkgsImport)

			if spec.expectError {
				require.Nil(t, registry)
				require.Error(t, err)
			} else {
				require.NotNil(t, registry)
				require.Nil(t, err)
			}
		})
	}
}

func TestImageName(t *testing.T) {
	require.Equal(t, "node:alpine", new(Registry).ImageName())
}

func TestAdditionalEnvVars(t *testing.T) {
	require.Equal(t, map[string]string{}, new(Registry).AdditionalEnvVars("name", "version"))
}
