package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCredentialsUseBase64Token(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]string
		expected bool
	}{
		{
			name:     "with no params",
			expected: false,
		},
		{
			name:     "with no base 64 key",
			params:   map[string]string{"test": "foo", "bar": "123"},
			expected: false,
		},
		{
			name:     "with base 64 key",
			params:   map[string]string{"test": "foo", "bar": "123", Base64TokenKey: "1"},
			expected: true,
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			credentials := Credentials{
				AdditionalParameters: spec.params,
			}

			require.Equal(t, spec.expected, credentials.UseBase64Token())
		})
	}
}

func TestRegistryRequireCredentialsToken(t *testing.T) {
	tests := []struct {
		name         string
		token        string
		errorMessage string
	}{
		{
			name:         "with token set",
			token:        "test",
			errorMessage: "",
		},
		{
			name:         "without token set",
			token:        "",
			errorMessage: "credentials token for destination in import test is required",
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			registry := Registry{
				Credentials: Credentials{Token: spec.token},
			}

			err := registry.requireCredentialsToken("destination", "test")

			if len(spec.errorMessage) != 0 {
				require.Error(t, err, spec.errorMessage)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestImportValidate(t *testing.T) {
	tests := []struct {
		name         string
		configImport Import
		errorMessage string
	}{
		{
			name: "with valid import",
			configImport: Import{
				Source:      Registry{URL: "https://source.registry"},
				Destination: Registry{URL: "https://destination.registry", Credentials: Credentials{Token: "token"}},
			},
			errorMessage: "",
		},
		{
			name: "with same urls",
			configImport: Import{
				Source:      Registry{URL: "https://same.registry"},
				Destination: Registry{URL: "https://same.registry", Credentials: Credentials{Token: "token"}},
			},
			errorMessage: "import test has the same url for the source and the destination",
		},
		{
			name: "with no destination registry token",
			configImport: Import{
				Source:      Registry{URL: "https://source.registry"},
				Destination: Registry{URL: "https://destination.registry"},
			},
			errorMessage: "credentials token for destination in import test is required",
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := spec.configImport.validate("test")

			if len(spec.errorMessage) != 0 {
				require.Error(t, err, spec.errorMessage)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
