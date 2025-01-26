package nuget

import (
	"errors"
	"fmt"
	"strings"

	"github.com/khulnasoft/packages-registry/config"
	"github.com/khulnasoft/packages-registry/util"
)

// Registy represents a NuGet registry given an import.
type Registry struct {
	pkgsImport config.Import
}

// NewRegistry will create a new NuGet registry given an import.
// Some validations are executed and could return an error.
func NewRegistry(pkgsImport config.Import) (*Registry, error) {
	if pkgsImport.Type != "nuget" {
		err := fmt.Errorf("NuGet Registry received the wrong import type: %q", pkgsImport.Type)
		return nil, err
	}

	registry := &Registry{
		pkgsImport: pkgsImport,
	}

	if err := registry.validate(); err != nil {
		return nil, err
	}

	return registry, nil
}

// ImageName returns the default image name for NuGet package imports.
func (r *Registry) ImageName() string {
	return "mono:6"
}

// AdditionalEnvVars returns the additional environment variables.
// The NuGet scripts don't need any.
func (r *Registry) AdditionalEnvVars(name, version string) map[string]string {
	return map[string]string{}
}

const (
	sourceRegistryLabel      = "pkgs_importer_source"
	destinationRegistryLabel = "pkgs_importer_destination"
)

// Scripts returns the script lines to execute a NuGet package import.
// Authentication and import is done by the usual nuget commands. See:
// - https://learn.microsoft.com/en-us/nuget/reference/cli-reference/cli-ref-sources
// - https://learn.microsoft.com/en-us/nuget/reference/cli-reference/cli-ref-install
//   - https://learn.microsoft.com/en-us/nuget/reference/cli-reference/cli-ref-sources
//   - https://learn.microsoft.com/en-us/nuget/reference/cli-reference/cli-ref-install
//   - https://learn.microsoft.com/en-us/nuget/reference/cli-reference/cli-ref-push
func (r *Registry) Scripts() ([]string, error) {
	scripts := make([]string, 0, 12)

	scripts = append(scripts, r.removeDefaults())
	scripts = append(scripts, r.configureAccess(r.pkgsImport.Source, sourceRegistryLabel))
	scripts = append(scripts, r.installScript(sourceRegistryLabel)...)
	scripts = append(scripts, r.resetAccess(sourceRegistryLabel))
	scripts = append(scripts, r.processPackage())
	scripts = append(scripts, r.configureAccess(r.pkgsImport.Destination, destinationRegistryLabel))
	scripts = append(scripts, r.pushScript(destinationRegistryLabel))

	return scripts, nil
}

func (r *Registry) removeDefaults() string {
	return "nuget sources Remove -Name nuget.org"
}

func (r *Registry) configureAccess(registry config.Registry, label string) string {
	cmd := new(strings.Builder)
	cmd.WriteString(fmt.Sprintf(`nuget sources Add -Name %s -Source "%s"`, label, registry.URL))

	if len(registry.Credentials.Token) != 0 {
		cmd.WriteString(fmt.Sprintf(` -password "%s"`, registry.Credentials.Token))
	}

	additionalParameters := registry.Credentials.AdditionalParameters
	for _, k := range util.OrderedMapKeysOf(additionalParameters) {
		v := additionalParameters[k]
		cmd.WriteString(fmt.Sprintf(" -%s %s", k, v))
	}

	return cmd.String()
}

func (r *Registry) installScript(label string) []string {
	return []string{
		"mkdir _pkg",
		fmt.Sprintf("nuget install $PACKAGE_NAME -Version $PACKAGE_VERSION -NoCache -DirectDownload -NonInteractive -DependencyVersion Ignore -Source %s -OutputDirectory _pkg", label),
	}
}

func (r *Registry) resetAccess(label string) string {
	return fmt.Sprintf("nuget sources Remove -Name %s", label)
}

func (r *Registry) processPackage() string {
	return "cd _pkg && cd $(ls -d */|head -n 1)"
}

func (r *Registry) pushScript(label string) string {
	return fmt.Sprintf("nuget push $(ls *.nupkg | head -n 1) -Source %s", label)
}

func (r *Registry) validate() error {
	if err := r.validateCredentials(r.pkgsImport.Source.Credentials); err != nil {
		return err
	}

	return r.validateCredentials(r.pkgsImport.Destination.Credentials)
}

var errInvalidCredentials = errors.New("NuGet credentials require a token and a username in authenticated registries")

func (r *Registry) validateCredentials(credentials config.Credentials) error {
	if len(credentials.Token) == 0 || len(credentials.AdditionalParameters["username"]) != 0 {
		return nil
	}

	return errInvalidCredentials
}
