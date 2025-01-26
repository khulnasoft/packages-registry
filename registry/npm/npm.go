// Package npm is a set of simple functions that provide all the custom parts to handle
// NPM package imports from registry A to registry B.
package npm

import (
	"fmt"
	"strings"

	"github.com/khulnasoft/packages-registry/config"
	"github.com/khulnasoft/packages-registry/util"
)

// Registy represents an npm registry given an import.
type Registry struct {
	pkgsImport config.Import
}

// NewRegistry will create a new npm registry given an import.
// Some validations are executed and could return an error.
func NewRegistry(pkgsImport config.Import) (*Registry, error) {
	if pkgsImport.Type != "npm" {
		err := fmt.Errorf("NPM Registry received the wrong import type: %q", pkgsImport.Type)
		return nil, err
	}

	registry := &Registry{
		pkgsImport: pkgsImport,
	}

	return registry, nil
}

// ImageName returns the default image name for npm package imports.
func (r *Registry) ImageName() string {
	return "node:alpine"
}

// AdditionalEnvVars returns the additional environment variables.
// The NPM scripts don't need any.
func (r *Registry) AdditionalEnvVars(name, version string) map[string]string {
	return map[string]string{}
}

// Scripts returns the script lines to execute an npm package import.
// Authentication is done by managing an .npmrc file.
// The import itself will use the usual npm commands. See:
// - https://docs.npmjs.com/cli/v9/commands/npm-pack#description
// - https://docs.npmjs.com/cli/v9/commands/npm-publish#description
func (r *Registry) Scripts() ([]string, error) {
	scripts := make([]string, 0, 15)

	scripts = append(scripts, r.configureAccess(r.pkgsImport.Source)...)
	scripts = append(scripts, r.packScript()...)
	scripts = append(scripts, r.resetAccess())
	scripts = append(scripts, r.processPackage()...)
	scripts = append(scripts, r.configureAccess(r.pkgsImport.Destination)...)
	scripts = append(scripts, r.publishScript())

	return scripts, nil
}

func (r *Registry) configureAccess(registry config.Registry) []string {
	scripts := make([]string, 0, 3)

	scripts = append(scripts, fmt.Sprintf(`echo "registry = %s" >> .npmrc`, registry.URL))

	if len(registry.Credentials.Token) != 0 {
		scripts = append(scripts, r.setAuthToken(registry))
	}

	additionalParameters := registry.Credentials.AdditionalParameters
	for _, k := range util.OrderedMapKeysOf(additionalParameters) {
		v := additionalParameters[k]
		if k != config.Base64TokenKey {
			scripts = append(scripts, r.setConfigValueScript(k, v))
		}
	}

	return scripts
}

func (r *Registry) setAuthToken(registry config.Registry) string {
	schemaLessUrl := strings.TrimPrefix(registry.URL, "https:")
	schemaLessUrl = strings.TrimPrefix(schemaLessUrl, "http:")
	var authKey string

	if registry.Credentials.UseBase64Token() {
		authKey = "_auth"
	} else {
		authKey = "_authToken"
	}

	key := fmt.Sprintf("%s:%s", schemaLessUrl, authKey)

	return r.setConfigValueScript(key, registry.Credentials.Token)
}

func (r *Registry) setConfigValueScript(key, value string) string {
	return fmt.Sprintf(`echo "%s = %s" >> .npmrc`, key, value)
}

func (r *Registry) resetAccess() string {
	return "rm -f .npmrc"
}

func (r *Registry) packScript() []string {
	// https://docs.npmjs.com/cli/v9/commands/npm-pack#description
	return []string{
		"mkdir _pkg",
		`npm pack --pack-destination="_pkg" $PACKAGE_NAME@$PACKAGE_VERSION`,
	}
}

func (r *Registry) processPackage() []string {
	return []string{
		"cd _pkg",
		"ls *.tgz | xargs tar zxvf",
		"cd package",
		"npm pkg delete publishConfig",
		"npm pack",
	}
}

func (r *Registry) publishScript() string {
	// https://docs.npmjs.com/cli/v9/commands/npm-publish#description
	return "ls *.tgz | xargs npm publish"
}
