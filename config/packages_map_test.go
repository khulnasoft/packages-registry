package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestGetPackagesMap(t *testing.T) {
	tests := []struct {
		name     string
		packages map[string][]string
	}{
		{
			name:     "import no packages",
			packages: map[string][]string{},
		},
		{
			name:     "import some packages",
			packages: map[string][]string{"my_pkg": {"1.2.3"}},
		},
		{
			name:     "import some packages with many versions",
			packages: map[string][]string{"my_pkg": {"1.2.3", "3.4.5"}},
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			t.Cleanup(viper.Reset)
			viper.Set("import1.packages", spec.packages)
			packages, err := GetPackagesMap("import1")
			require.Nil(t, err)
			require.Equal(t, spec.packages, packages)
			require.Equal(t, spec.packages, viper.Get("import1.cached_packages"))

			packages, err = GetPackagesMap("import1")
			require.Nil(t, err)
			require.Equal(t, spec.packages, packages)
		})
	}
}

func TestGetPackagesMapWithCsv(t *testing.T) {
	tests := []struct {
		name        string
		csvFilePath interface{}
		packages    map[string][]string
		erroneous   bool
	}{
		{
			name:        "no existing csv file",
			csvFilePath: "testdata/doesnt_exist.csv",
			erroneous:   true,
		},
		{
			name:        "any string",
			csvFilePath: "testdata/test",
			erroneous:   true,
		},
		{
			name:        "not a string",
			csvFilePath: 33,
			packages:    map[string][]string{},
		},
		{
			name:        "simply csv",
			csvFilePath: "../testdata/csv/simple.csv",
			packages: map[string][]string{
				"package1":       {"1.2.3"},
				"@test/package2": {"3.2.1"},
				"package3":       {"2.3.5"},
			},
		},
		{
			name:        "complex csv",
			csvFilePath: "../testdata/csv/complex.csv",
			packages: map[string][]string{
				"package1":       {"1.2.3", "9.0.1", "8.4.76"},
				"@test/package2": {"3.2.1", "6.2.3"},
				"package3":       {"2.3.5"},
			},
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			t.Cleanup(viper.Reset)
			viper.Set("import1.packages", spec.csvFilePath)
			packages, err := GetPackagesMap("import1")
			if spec.erroneous {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
				require.Equal(t, spec.packages, packages)
				require.Equal(t, spec.packages, viper.Get("import1.cached_packages"))
			}
		})
	}
}
