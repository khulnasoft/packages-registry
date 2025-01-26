package nuget

import (
	"fmt"
	"strings"
	"testing"

	"github.com/khulnasoft/packages-registry/config"
	"github.com/khulnasoft/packages-registry/util"
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
			name:                        "with no additional config",
			sourceUrl:                   "http://source.test",
			sourceToken:                 "TOKEN_FOR_SOURCE",
			sourceAdditionalParams:      map[string]string{"username": "username_for_source"},
			destinationUrl:              "https://destination.test",
			destinationToken:            "TOKEN_FOR_DESTINATION",
			destinationAdditionalParams: map[string]string{"username": "username_for_destination"},
		},
		{
			name:                        "with additional config",
			sourceUrl:                   "http://source.test",
			sourceToken:                 "TOKEN_FOR_SOURCE",
			sourceAdditionalParams:      map[string]string{"username": "username_for_source", "test": "foo", "bar": "123"},
			destinationUrl:              "https://destination.test",
			destinationToken:            "TOKEN_FOR_DESTINATION",
			destinationAdditionalParams: map[string]string{"username": "username_for_destination", "test": "foo", "bar": "123"},
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

			assertRegistryAccess(t, joinedScripts, sourceRegistryLabel, spec.sourceUrl, spec.sourceToken, spec.sourceAdditionalParams)
			assertRegistryAccess(t, joinedScripts, destinationRegistryLabel, spec.destinationUrl, spec.destinationToken, spec.destinationAdditionalParams)
		})
	}
}

func assertRegistryAccess(t *testing.T, scripts string, label string, registryUrl string, token string, params map[string]string) {
	expected := fmt.Sprintf(`nuget sources Add -Name %s -Source "%s" -password "%s"`, label, registryUrl, token)

	if len(params) > 0 {
		for _, k := range util.OrderedMapKeysOf(params) {
			v := params[k]
			expected = fmt.Sprintf("%s -%s %s", expected, k, v)
		}
	}

	require.Contains(t, scripts, expected)
}

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name         string
		errorMessage string
		pkgsImport   config.Import
	}{
		{
			name:         "with wrong import type",
			errorMessage: `NuGet Registry received the wrong import type: "wrong_type"`,
			pkgsImport: config.Import{
				Type: "wrong_type",
			},
		},
		{
			name:         "with correct import type",
			errorMessage: "",
			pkgsImport: config.Import{
				Type: "nuget",
			},
		},
		{
			name:         "with regisitries with correct credentials",
			errorMessage: "",
			pkgsImport: config.Import{
				Type: "nuget",
				Source: config.Registry{
					URL: "http://source.registry",
				},
				Destination: config.Registry{
					URL: "http://destination.registry",
					Credentials: config.Credentials{
						AdditionalParameters: map[string]string{
							"username": "user",
						},
						Token: "1234567890",
					},
				},
			},
		},
		{
			name:         "with source registry with partial credentials",
			errorMessage: errInvalidCredentials.Error(),
			pkgsImport: config.Import{
				Type: "nuget",
				Source: config.Registry{
					URL: "http://source.registry",
					Credentials: config.Credentials{
						Token: "1234567890",
					},
				},
				Destination: config.Registry{
					URL: "http://destination.registry",
				},
			},
		},
		{
			name:         "with destination registry with partial credentials",
			errorMessage: errInvalidCredentials.Error(),
			pkgsImport: config.Import{
				Type: "nuget",
				Source: config.Registry{
					URL: "http://source.registry",
				},
				Destination: config.Registry{
					URL: "http://destination.registry",
					Credentials: config.Credentials{
						Token: "1234567890",
					},
				},
			},
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			registry, err := NewRegistry(spec.pkgsImport)

			if len(spec.errorMessage) != 0 {
				require.Nil(t, registry)
				require.ErrorContains(t, err, spec.errorMessage)
			} else {
				require.NotNil(t, registry)
				require.Nil(t, err)
			}
		})
	}
}

func TestImageName(t *testing.T) {
	require.Equal(t, "mono:6", new(Registry).ImageName())
}

func TestAdditionalEnvVars(t *testing.T) {
	require.Equal(t, map[string]string{}, new(Registry).AdditionalEnvVars("name", "version"))
}
