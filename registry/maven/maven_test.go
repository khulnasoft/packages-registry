package maven

import (
	"fmt"
	"strings"
	"testing"

	"github.com/khulnasoft/packages-registry/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ScriptsTest struct {
	name                  string
	sourceUrl             string
	sourceToken           string
	sourceUsername        string
	sourceHeaderName      string
	destinationUrl        string
	destinationToken      string
	destinationUsername   string
	destinationHeaderName string
}

func TestScripts(t *testing.T) {
	tests := []ScriptsTest{
		{
			name:                "source with username and token, destination with username and token",
			sourceUrl:           "http://source.test",
			sourceUsername:      "USER_FOR_SOURCE",
			sourceToken:         "TOKEN_FOR_SOURCE",
			destinationUrl:      "https://destination.test",
			destinationUsername: "USER_FOR_DESTINATION",
			destinationToken:    "TOKEN_FOR_DESTINATION",
		},
		{
			name:                "source without credentials, destination with username and token",
			sourceUrl:           "http://source.test",
			destinationUrl:      "https://destination.test",
			destinationUsername: "USER_FOR_DESTINATION",
			destinationToken:    "TOKEN_FOR_DESTINATION",
		},
		{
			name:                  "source with header and token, destination with header and token",
			sourceUrl:             "http://source.test",
			sourceHeaderName:      "HEADER_FOR_SOURCE",
			sourceToken:           "TOKEN_FOR_SOURCE",
			destinationUrl:        "https://destination.test",
			destinationHeaderName: "HEADER_FOR_DESTINATION",
			destinationToken:      "TOKEN_FOR_DESTINATION",
		},
		{
			name:                  "source without credentials, destination with header and token",
			sourceUrl:             "http://source.test",
			destinationUrl:        "https://destination.test",
			destinationHeaderName: "HEADER_FOR_DESTINATION",
			destinationToken:      "TOKEN_FOR_DESTINATION",
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			sourceAdditionalParams := map[string]string{}
			destinationAdditionalParams := map[string]string{}
			if len(spec.sourceUsername) != 0 {
				sourceAdditionalParams["username"] = spec.sourceUsername
			}
			if len(spec.sourceHeaderName) != 0 {
				sourceAdditionalParams["header_name"] = spec.sourceHeaderName
			}
			if len(spec.destinationUsername) != 0 {
				destinationAdditionalParams["username"] = spec.destinationUsername
			}
			if len(spec.destinationHeaderName) != 0 {
				destinationAdditionalParams["header_name"] = spec.destinationHeaderName
			}
			pkgsImport := config.Import{
				Source: config.Registry{
					URL: spec.sourceUrl,
					Credentials: config.Credentials{
						Token:                spec.sourceToken,
						AdditionalParameters: sourceAdditionalParams,
					},
				},
				Destination: config.Registry{
					URL: spec.destinationUrl,
					Credentials: config.Credentials{
						Token:                spec.destinationToken,
						AdditionalParameters: destinationAdditionalParams,
					},
				},
			}

			registry := Registry{
				pkgsImport: pkgsImport,
			}

			scripts, err := registry.Scripts()
			assert.NoError(t, err)
			joinedScripts := strings.Join(scripts, "\n")

			if len(spec.sourceUsername) != 0 || len(spec.sourceHeaderName) != 0 {
				assertRegistryAccess(t, joinedScripts, "source", spec)
			}
			assertRegistryAccess(t, joinedScripts, "destination", spec)
		})
	}
}

func assertRegistryAccess(t *testing.T, scripts string, kind string, spec ScriptsTest) {
	var expected string
	basicAuthTemplate := `echo "<settings><servers><server><id>%s</id><username>%s</username><password>%s</password></server></servers></settings>" > settings.xml`
	customHeaderTemplate := `echo "<settings><servers><server><id>%s</id><configuration><httpHeaders><property><name>%s</name><value>%s</value></property></httpHeaders></configuration></server></servers></settings>`
	if kind == "source" {
		if len(spec.sourceUsername) != 0 {
			expected = fmt.Sprintf(basicAuthTemplate, sourceRegistryLabel, spec.sourceUsername, spec.sourceToken)
		}
		if len(spec.sourceHeaderName) != 0 {
			expected = fmt.Sprintf(customHeaderTemplate, sourceRegistryLabel, spec.sourceHeaderName, spec.sourceToken)
		}
	}

	if kind == "destination" {
		if len(spec.destinationUsername) != 0 {
			expected = fmt.Sprintf(basicAuthTemplate, destinationRegistryLabel, spec.destinationUsername, spec.destinationToken)
		}
		if len(spec.destinationHeaderName) != 0 {
			expected = fmt.Sprintf(customHeaderTemplate, destinationRegistryLabel, spec.destinationHeaderName, spec.destinationToken)
		}
	}

	require.Contains(t, scripts, expected)
}

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name         string
		errorMessage string
		pkgsImport   config.Import
		pkgs         map[string][]string
	}{
		{
			name:         "with correct import type",
			errorMessage: "",
			pkgsImport: config.Import{
				Type: "maven",
			},
		},
		{
			name:         "with regisitries with correct credentials: username and token",
			errorMessage: "",
			pkgsImport: config.Import{
				Type: "maven",
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
			pkgs: map[string][]string{
				"my.company:package1": {"1.2.3"},
				"my.company:package2": {"1.2.3"},
			},
		},
		{
			name:         "with regisitries with correct credentials: header_name and token",
			errorMessage: "",
			pkgsImport: config.Import{
				Type: "maven",
				Source: config.Registry{
					URL: "http://source.registry",
				},
				Destination: config.Registry{
					URL: "http://destination.registry",
					Credentials: config.Credentials{
						AdditionalParameters: map[string]string{
							"header_name": "header",
						},
						Token: "1234567890",
					},
				},
			},
			pkgs: map[string][]string{
				"my.company:package1": {"1.2.3"},
				"my.company:package2": {"1.2.3"},
			},
		},
		{
			name:         "with source registry with partial credentials",
			errorMessage: errInvalidCredentials.Error(),
			pkgsImport: config.Import{
				Type: "maven",
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
				Type: "maven",
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
		{
			name:         "with packages with no separator",
			errorMessage: "my.company.invalid is an invalid Maven package name. It must contain : between the group ID and the artifact ID.",
			pkgsImport: config.Import{
				Type: "maven",
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
			pkgs: map[string][]string{
				"my.company:package1": {"1.2.3"},
				"my.company.invalid":  {"1.2.3"},
			},
		},
		{
			name:         "with packages with no group ID",
			errorMessage: ":package2 is an invalid Maven package name. It must contain : between the group ID and the artifact ID.",
			pkgsImport: config.Import{
				Type: "maven",
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
			pkgs: map[string][]string{
				"my.company:package1": {"1.2.3"},
				":package2":           {"1.2.3"},
			},
		},
		{
			name:         "with packages with no artifact ID",
			errorMessage: "my.company: is an invalid Maven package name. It must contain : between the group ID and the artifact ID.",
			pkgsImport: config.Import{
				Type: "maven",
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
			pkgs: map[string][]string{
				"my.company:package1": {"1.2.3"},
				"my.company:":         {"1.2.3"},
			},
		},
		{
			name: "with packages with valid packaging",
			pkgsImport: config.Import{
				Type: "maven",
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
			pkgs: map[string][]string{
				"my.company:package1": {"1.2.3:ear"},
				"my.company:package2": {"1.2.3:war"},
			},
		},
		{
			name:         "with packages with invalid packaging",
			errorMessage: "zip is an invalid Maven packaging string. It must be one of : [pom jar maven-plugin ejb war ear rar aar].",
			pkgsImport: config.Import{
				Type: "maven",
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
			pkgs: map[string][]string{
				"my.company:package1": {"1.2.3:zip"},
				"my.company:package2": {"1.2.3"},
			},
		},
		{
			name:         "with packages with packaging and classifier",
			errorMessage: "1.2.3:jar:javadoc is an invalid Maven version string. It must be in the form of : version[:packaging].",
			pkgsImport: config.Import{
				Type: "maven",
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
			pkgs: map[string][]string{
				"my.company:package1": {"1.2.3:jar:javadoc"},
				"my.company:package2": {"1.2.3"},
			},
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			t.Cleanup(viper.Reset)
			viper.Set("import1.packages", spec.pkgs)

			registry, err := NewRegistry(spec.pkgsImport, "import1")

			if spec.errorMessage != "" {
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
	require.Equal(t, "maven:eclipse-temurin", new(Registry).ImageName())
}

func TestAdditionalEnvVars(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		expectedHeaders map[string]string
	}{
		{
			name:            "with standard version string",
			version:         "1.2.3",
			expectedHeaders: map[string]string{"PACKAGE_PACKAGING": "jar"},
		},
		{
			name:            "with a packaging suffix",
			version:         "1.2.3:war",
			expectedHeaders: map[string]string{"PACKAGE_PACKAGING": "war"},
		},
		{
			name:            "with a packaging and a classifier suffix",
			version:         "1.2.3:war:sources",
			expectedHeaders: map[string]string{"PACKAGE_PACKAGING": "jar"}, // classifiers not supported, we return the default headers
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			require.Equal(t, spec.expectedHeaders, new(Registry).AdditionalEnvVars("name", spec.version))
		})
	}
}
