// Package config centralize all logic around the config file (config.yml by default).
// The only available function is Load() which will load the file into dedicated structs.
// It also performns validations.
package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/khulnasoft/packages-registry/logger"
	"github.com/spf13/viper"
)

// Load will load the configuration yaml file into Configuration. It will also run several
// checks and return an error if one validation fails.
func Load() (*Configuration, error) {
	var config Configuration
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	logger.LogInfo("Config loaded")
	return &config, nil
}
