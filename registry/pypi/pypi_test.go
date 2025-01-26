package pypi

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/khulnasoft/packages-registry/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ScriptsTest struct {
	name                string
	sourceUrl           string
	sourceToken         string
	sourceUsername      string
	destinationUrl      string
	destinationToken    string
	destinationUsername string
}

func TestScripts(t *testing.T) {
	tests := []ScriptsTest{
		{
			name:                "public source",
			sourceUrl:           "http://source.test",
			destinationUrl:      "https://destination.test",
			destinationToken:    "TOKEN_FOR_DESTINATION",
			destinationUsername: "username_for_destination",
		},
		{
			name:                "private source",
			sourceUrl:           "http://source.test",
			sourceToken:         "TOKEN_FOR_SOURCE",
			sourceUsername:      "username_for_source",
			destinationUrl:      "https://destination.test",
			destinationToken:    "TOKEN_FOR_DESTINATION",
			destinationUsername: "username_for_destination",
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			pkgsImport := config.Import{
				Source: config.Registry{
					URL: spec.sourceUrl,
					Credentials: config.Credentials{
						Token:                spec.sourceToken,
						AdditionalParameters: map[string]string{"username": spec.sourceUsername},
					},
				},
				Destination: config.Registry{
					URL: spec.destinationUrl,
					Credentials: config.Credentials{
						Token:                spec.destinationToken,
						AdditionalParameters: map[string]string{"username": spec.destinationUsername},
					},
				},
			}

			registry := Registry{
				pkgsImport: pkgsImport,
			}

			scripts, err := registry.Scripts()
			assert.NoError(t, err)

			err = assertSourceAccess(t, scripts, spec)
			assert.NoError(t, err)
			assertDestinationAccess(t, scripts, spec)
		})
	}
}

func assertSourceAccess(t *testing.T, scripts []string, spec ScriptsTest) error {
	address, err := url.Parse(spec.sourceUrl)
	if err != nil {
		return err
	}

	if len(spec.sourceUsername) != 0 && len(spec.sourceToken) != 0 {
		address.User = url.UserPassword(spec.sourceUsername, spec.sourceToken)
	}

	expected := fmt.Sprintf(`python -m pip download "$PACKAGE_NAME==$PACKAGE_VERSION" -d pkgs --no-cache-dir --no-deps -i %s`, address.String())

	require.Contains(t, scripts, expected)

	return nil
}

func assertDestinationAccess(t *testing.T, scripts []string, spec ScriptsTest) {
	expected := fmt.Sprintf(`python -m twine upload --repository-url %s -u "%s" -p "%s" ./*`, spec.destinationUrl, spec.destinationUsername, spec.destinationToken)

	require.Contains(t, scripts, expected)
}

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name         string
		errorMessage string
		pkgsImport   config.Import
	}{
		{
			name: "with correct import type",
			pkgsImport: config.Import{
				Type: "pypi",
			},
		},
		{
			name: "with registries with correct credentials",
			pkgsImport: config.Import{
				Type: "pypi",
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
				Type: "pypi",
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
				Type: "pypi",
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
	require.Equal(t, "python:alpine", new(Registry).ImageName())
}

func TestAdditionalEnvVars(t *testing.T) {
	require.Equal(t, map[string]string{}, new(Registry).AdditionalEnvVars("name", "version"))
}
