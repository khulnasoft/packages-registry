package khulnasoft

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/khulnasoft/packages-registry/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type testImport struct {
	Name     string
	Import   config.Import
	Packages map[string]string
}

var singleImport = []testImport{
	{
		Name: "import1",
		Packages: map[string]string{
			"package1":          "2.3.4",
			"@import1/package1": "2.1.0",
		},
		Import: config.Import{
			Type: "npm",
			Source: config.Registry{
				URL: "http://source.test",
				Credentials: config.Credentials{
					Token: "1234567890",
				},
			},
			Destination: config.Registry{
				URL: "http://destination.test",
				Credentials: config.Credentials{
					Token: "1234567890",
				},
			},
		},
	},
}

var multipleImports = []testImport{
	{
		Name: "import1",
		Packages: map[string]string{
			"package1":          "2.3.4",
			"@import1/package1": "2.1.0",
			"same_package":      "6.5.3",
		},
		Import: config.Import{
			Type: "npm",
			Source: config.Registry{
				URL: "http://source.test",
				Credentials: config.Credentials{
					Token: "1234567890",
				},
			},
			Destination: config.Registry{
				URL: "http://destination1.test",
				Credentials: config.Credentials{
					Token: "1234567890",
				},
			},
		},
	},
	{
		Name: "import2",
		Packages: map[string]string{
			"package2":          "2.3.4",
			"@import2/package2": "2.1.0",
			"same_package":      "6.5.3",
		},
		Import: config.Import{
			Type: "npm",
			Source: config.Registry{
				URL: "http://source.test",
				Credentials: config.Credentials{
					Token: "1234567890",
				},
			},
			Destination: config.Registry{
				URL: "http://destination2.test",
				Credentials: config.Credentials{
					Token: "1234567890",
				},
			},
		},
	},
}

func TestGeneratePipelineConfig(t *testing.T) {
	tests := []struct {
		name    string
		imports []testImport
	}{
		{
			name:    "single import",
			imports: singleImport,
		},
		{
			name:    "multiple imports",
			imports: multipleImports,
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			t.Cleanup(viper.Reset)
			g := NewGenerator(configFrom(spec.imports))

			file, err := os.CreateTemp(os.TempDir(), "output*.yml")
			if err != nil {
				t.Errorf("Can't create temp file")
				t.Fail()
			}
			defer file.Close()
			defer os.Remove(file.Name())

			err = g.Generate(file)
			require.Nil(t, err)

			require.FileExists(t, file.Name())
			bytes, err := os.ReadFile(file.Name())
			if err != nil {
				t.Errorf("Can't read temp file")
				t.Fail()
			}

			content := string(bytes)
			require.Contains(t, content, "image: node:alpine")
			assertContainsPackages(t, content, spec.imports)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name            string
		fileContent     string
		maxSize         int64
		validationFails bool
	}{
		{
			name:            "with a valid file",
			fileContent:     "a",
			maxSize:         10,
			validationFails: false,
		},
		{
			name:            "with a invalid file",
			fileContent:     "test",
			maxSize:         1,
			validationFails: true,
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			file, err := os.CreateTemp(os.TempDir(), "output*.yml")
			if err != nil {
				t.Errorf("Can't create temp file")
				t.Fail()
			}
			defer file.Close()
			defer os.Remove(file.Name())

			n, err := file.Write([]byte(spec.fileContent))

			require.Nil(t, err)
			require.Greater(t, n, 0)

			g := Generator{}
			err = g.validate(file, spec.maxSize)

			if spec.validationFails {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func assertContainsPackages(t *testing.T, content string, imports []testImport) {
	for _, pkgsImport := range imports {
		for name, version := range pkgsImport.Packages {
			if strings.HasPrefix(name, "@") {
				require.Contains(t, content, fmt.Sprintf("PACKAGE_NAME: '%s'", name))
			} else {
				require.Contains(t, content, fmt.Sprintf("PACKAGE_NAME: %s", name))
			}
			require.Contains(t, content, fmt.Sprintf("PACKAGE_VERSION: %s", version))

			jobName := fmt.Sprintf("%s:%s:%s:", pkgsImport.Name, name, version)
			require.Contains(t, content, jobName)
		}
	}
}

func configFrom(imports []testImport) *config.Configuration {
	configImports := make(map[string]config.Import)

	for _, pkgsImport := range imports {
		configImports[pkgsImport.Name] = pkgsImport.Import
		viper.Set(fmt.Sprintf("%s.packages", pkgsImport.Name), pkgsImport.Packages)
	}

	return &config.Configuration{
		Imports: configImports,
	}
}
