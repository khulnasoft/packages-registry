package registry

import (
	"testing"

	"github.com/khulnasoft/packages-registry/config"
	"github.com/stretchr/testify/require"
)

func TestGetRegistry(t *testing.T) {
	tests := []struct {
		name           string
		importType     string
		expectRegistry bool
		expectError    bool
	}{
		{
			name:           "with type npm",
			importType:     "npm",
			expectRegistry: true,
			expectError:    false,
		},
		{
			name:           "with type nuget",
			importType:     "nuget",
			expectRegistry: true,
			expectError:    false,
		},
		{
			name:           "with type maven",
			importType:     "maven",
			expectRegistry: true,
			expectError:    false,
		},
		{
			name:           "with type pypi",
			importType:     "pypi",
			expectRegistry: true,
			expectError:    false,
		},
		{
			name:           "with an unknown type",
			importType:     "you_dont_know_me",
			expectRegistry: false,
			expectError:    true,
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			pkgsImport := config.Import{Type: spec.importType}

			registry, err := GetRegistry(pkgsImport, "import1")

			if spec.expectRegistry {
				require.Implements(t, (*Registry)(nil), registry)
			} else {
				require.Nil(t, registry)
			}

			if spec.expectError {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
