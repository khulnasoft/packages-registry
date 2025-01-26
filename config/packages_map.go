package config

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// GetPackagesMap will try to read and return all the packages attached to an import,
// referenced by the name.
func GetPackagesMap(name string) (map[string][]string, error) {
	cacheKey := fmt.Sprintf("%s.cached_packages", name)

	if pkgs := viper.GetStringMapStringSlice(cacheKey); len(pkgs) != 0 {
		return pkgs, nil
	}

	packagesMapValue, err := getPackagesMapValue(name)
	if err != nil {
		return nil, err
	}

	viper.Set(cacheKey, packagesMapValue)

	return packagesMapValue, nil
}

func getPackagesMapValue(name string) (map[string][]string, error) {
	key := fmt.Sprintf("%s.packages", name)
	value := viper.Get(key)

	if value != nil && reflect.TypeOf(value).Kind() == reflect.String {
		stringValue := viper.GetString(key)
		if strings.HasSuffix(stringValue, ".csv") {
			return getPackagesMapFromCSV(stringValue)
		} else {
			return nil, fmt.Errorf("packages of import %q (value %q) is not a csv file path", name, stringValue)
		}
	}

	return viper.GetStringMapStringSlice(key), nil
}

func getPackagesMapFromCSV(csvFilepath string) (map[string][]string, error) {
	file, err := os.Open(csvFilepath)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 2
	reader.ReuseRecord = true
	packages := map[string][]string{}

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		name := record[0]
		version := record[1]

		if packages[name] == nil {
			packages[name] = make([]string, 0, 1)
		}
		packages[name] = append(packages[name], version)
	}

	return packages, nil
}
