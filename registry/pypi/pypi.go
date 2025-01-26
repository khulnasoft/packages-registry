package pypi

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/khulnasoft/packages-registry/config"
)

// Registy represents a PyPI registry given an import.
type Registry struct {
	pkgsImport config.Import
}

// NewRegistry will create a new PyPI registry given an import.
// Some validations are executed and could return an error.
func NewRegistry(pkgsImport config.Import) (*Registry, error) {
	registry := &Registry{
		pkgsImport: pkgsImport,
	}

	if err := registry.validate(); err != nil {
		return nil, err
	}

	return registry, nil
}

// ImageName returns the default image name for PyPI package imports.
func (r *Registry) ImageName() string {
	return "python:alpine"
}

// AdditionalEnvVars returns the additional environment variables.
// The PyPI scripts don't need any.
func (r *Registry) AdditionalEnvVars(name, version string) map[string]string {
	return map[string]string{}
}

// Scripts returns the script lines to execute a PyPI package import.
// pip is used for the download and twine is used for the upload. See:
// - https://pip.pypa.io/en/stable/cli/pip_download/
// - https://twine.readthedocs.io/en/stable/index.html#twine-upload
func (r *Registry) Scripts() ([]string, error) {
	install, err := r.installScript()
	if err != nil {
		return nil, err
	}

	return []string{
		install,
		"cd pkgs",
		"python -m pip install twine",
		r.pushScript(),
	}, nil
}

func (r *Registry) installScript() (string, error) {
	fullUrl, err := r.getFullUrl(r.pkgsImport.Source)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`python -m pip download "$PACKAGE_NAME==$PACKAGE_VERSION" -d pkgs --no-cache-dir --no-deps -i %s`, fullUrl), nil
}

func (r *Registry) pushScript() string {
	cmd := new(strings.Builder)

	cmd.WriteString(fmt.Sprintf("python -m twine upload --repository-url %s ", r.pkgsImport.Destination.URL))

	user, pw := r.getUsernameAndPassword(r.pkgsImport.Destination.Credentials)

	if len(user) != 0 {
		cmd.WriteString(fmt.Sprintf(`-u "%s" `, user))
	}

	if len(pw) != 0 {
		cmd.WriteString(fmt.Sprintf(`-p "%s" `, pw))
	}

	cmd.WriteString("./*")

	return cmd.String()
}

func (r *Registry) getFullUrl(reg config.Registry) (string, error) {
	address, err := url.Parse(reg.URL)
	if err != nil {
		return "", err
	}

	user, pw := r.getUsernameAndPassword(reg.Credentials)
	if len(user) != 0 && len(pw) != 0 {
		address.User = url.UserPassword(user, pw)
	}

	return address.String(), nil
}

func (r *Registry) getUsernameAndPassword(cr config.Credentials) (string, string) {
	return cr.AdditionalParameters["username"], cr.Token
}

func (r *Registry) validate() error {
	if err := r.validateCredentials(r.pkgsImport.Source.Credentials); err != nil {
		return err
	}

	return r.validateCredentials(r.pkgsImport.Destination.Credentials)
}

var errInvalidCredentials = errors.New("PyPI credentials require a token and a username in authenticated registries")

func (r *Registry) validateCredentials(credentials config.Credentials) error {
	if len(credentials.Token) == 0 || len(credentials.AdditionalParameters["username"]) != 0 {
		return nil
	}

	return errInvalidCredentials
}
