package config

import (
	"fmt"
)

// Represents credentials to use when interacting with the given Registry.
// Credentials can have additional fields to support specific needs.
type Credentials struct {
	Token                string            `validate:"omitempty,required"` // the credential token. Required.
	AdditionalParameters map[string]string `mapstructure:",remain"`        // all other fields as a map
}

// The key for the additional parameter that indicates that the token is a base64 token.
const Base64TokenKey = "_base64_token"

// UseBase64Token is a helper function that will try to read the Base64TokenKey and check if it is set to "1".
func (c *Credentials) UseBase64Token() bool {
	return c.AdditionalParameters[Base64TokenKey] == "1"
}

// Represents a package registry.
type Registry struct {
	URL         string      `validate:"required,url"` // the url where the registry is located. Required.
	Credentials Credentials // the credentials to be used. Optionnal.
}

func (r *Registry) requireCredentialsToken(registryLabel string, importName string) error {
	if len(r.Credentials.Token) == 0 {
		return fmt.Errorf("credentials token for %s in import %q is required", registryLabel, importName)
	}
	return nil
}

// Represents an import operation to carry. It's mainly caracterized by a source and a destination
// registry.
type Import struct {
	Type        string   `validate:"required,oneof=npm nuget maven pypi"` // The import type. Only npm, nuget, maven and pypi are valid values.
	Image       string   // The image to use for the jobs that execute this import. Optionnal.
	Source      Registry `validate:"required"` // The source registry. Required.
	Destination Registry `validate:"required"` // The destination registry. Required.
}

func (i *Import) validate(importName string) error {
	if i.Source.URL == i.Destination.URL {
		return fmt.Errorf("import %q has the same url for the source and the destination", importName)
	}

	return i.Destination.requireCredentialsToken("destination", importName)
}

// The root struct of the configuration. At its core, it's a set of imports where each import has a
// name.
type Configuration struct {
	Imports map[string]Import `mapstructure:",remain" validate:"dive"` // Map of imports
}

func (c *Configuration) validate() error {
	for name, i := range c.Imports {
		if err := i.validate(name); err != nil {
			return err
		}
	}

	return nil
}
