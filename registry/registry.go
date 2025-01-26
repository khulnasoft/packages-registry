// Package registry is a collection of registries that contains customization for a given package
// format.
package registry

import (
	"fmt"

	"github.com/khulnasoft/packages-registry/config"
	"github.com/khulnasoft/packages-registry/registry/maven"
	"github.com/khulnasoft/packages-registry/registry/npm"
	"github.com/khulnasoft/packages-registry/registry/nuget"
	"github.com/khulnasoft/packages-registry/registry/pypi"
)

// Registry is the interface that all package format dedicated registries will comply to.
type Registry interface {
	Scripts() ([]string, error)                         // Returns the set of scripts needed to execute the import of a single package.
	ImageName() string                                  // Returns the default docker image name that provides the necessary CLI tools.
	AdditionalEnvVars(string, string) map[string]string // Returns the additional environment variables that the pipeline jobs might need.
}

// GetRegistry will read the given import type and return the correct registry for the right package
// format. Returns an error if such registry can't be found.
func GetRegistry(pkgsImport config.Import, importName string) (Registry, error) {
	switch pkgsImport.Type {
	case "npm":
		return npm.NewRegistry(pkgsImport)
	case "nuget":
		return nuget.NewRegistry(pkgsImport)
	case "maven":
		return maven.NewRegistry(pkgsImport, importName)
	case "pypi":
		return pypi.NewRegistry(pkgsImport)
	}

	return nil, fmt.Errorf("no registry object for type %q in import %q", pkgsImport.Type, importName)
}
