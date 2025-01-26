package maven

import (
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/khulnasoft/packages-registry/config"
	"golang.org/x/exp/slices"
)

// Registy represents a Maven registry given an import.
type Registry struct {
	pkgsImport config.Import
}

// NewRegistry will create a new Maven registry given an import.
// Validations are executed and could return an error.
func NewRegistry(pkgsImport config.Import, importName string) (*Registry, error) {
	registry := &Registry{
		pkgsImport: pkgsImport,
	}

	if err := registry.validate(importName); err != nil {
		return nil, err
	}

	return registry, nil
}

// ImageName returns the default image name for Maven package imports.
func (r *Registry) ImageName() string {
	return "maven:eclipse-temurin"
}

const mavenCoordinatesSeparator = ":"

// AdditionalEnvVars returns the additional environment variables.
func (r *Registry) AdditionalEnvVars(name, version string) map[string]string {
	count := strings.Count(version, mavenCoordinatesSeparator)

	if count == 1 {
		packaging := strings.Split(version, mavenCoordinatesSeparator)[1]
		return map[string]string{"PACKAGE_PACKAGING": packaging}
	}

	return map[string]string{"PACKAGE_PACKAGING": "jar"}
}

const (
	sourceRegistryLabel      = "pkgs_importer_source"
	destinationRegistryLabel = "pkgs_importer_destination"
)

// Scripts returns the script lines to execute an npm package import.
// Authentication is done by managing a settings.xml file.
// The import itself will use the following maven commands:
// - https://maven.apache.org/plugins/maven-dependency-plugin/get-mojo.html
// - https://maven.apache.org/plugins/maven-deploy-plugin/deploy-file-mojo.html
func (r *Registry) Scripts() ([]string, error) {
	scripts := make([]string, 0, 6)

	if r.pkgsImport.Source.Credentials.Token != "" {
		accessScripts, err := r.configureAccess(r.pkgsImport.Source.Credentials, sourceRegistryLabel)
		if err != nil {
			return []string{}, err
		}
		scripts = append(scripts, accessScripts)
	}
	scripts = append(scripts, r.pullScript(sourceRegistryLabel))
	scripts = append(scripts, r.cdIntoPackageDirectory()...)

	accessScripts, err := r.configureAccess(r.pkgsImport.Destination.Credentials, destinationRegistryLabel)
	if err != nil {
		return []string{}, err
	}
	scripts = append(scripts, accessScripts)
	scripts = append(scripts, r.pushScript(destinationRegistryLabel))

	return scripts, nil
}

const (
	xmlSettingsBasicAuthTemplate    = "<settings><servers><server><id>{{.Label}}</id><username>{{.Username}}</username><password>{{.Password}}</password></server></servers></settings>"
	xmlSettingsCustomHeaderTemplate = "<settings><servers><server><id>{{.Label}}</id><configuration><httpHeaders><property><name>{{.HeaderName}}</name><value>{{.HeaderValue}}</value></property></httpHeaders></configuration></server></servers></settings>"
	settingsFile                    = "settings.xml"
)

type BasicAuthValues struct {
	Label, Username, Password string
}

type CustomHeaderValues struct {
	Label, HeaderName, HeaderValue string
}

func (r *Registry) configureAccess(credentials config.Credentials, label string) (string, error) {
	if credentials.AdditionalParameters["username"] != "" {
		values := BasicAuthValues{
			Label:    label,
			Username: credentials.AdditionalParameters["username"],
			Password: credentials.Token,
		}

		return r.configureAccessWithTemplate(xmlSettingsBasicAuthTemplate, values)
	}
	if credentials.AdditionalParameters["header_name"] != "" {
		values := CustomHeaderValues{
			Label:       label,
			HeaderName:  credentials.AdditionalParameters["header_name"],
			HeaderValue: credentials.Token,
		}

		return r.configureAccessWithTemplate(xmlSettingsCustomHeaderTemplate, values)
	}
	return "", nil
}

func (r *Registry) configureAccessWithTemplate(templateContent string, values interface{}) (string, error) {
	cmd := new(strings.Builder)

	t := template.Must(template.New("xmlSettings").Parse(templateContent))

	cmd.WriteString(`echo "`)

	if err := t.Execute(cmd, values); err != nil {
		return "", err
	}

	cmd.WriteString(fmt.Sprintf(`" > %s`, settingsFile))

	return cmd.String(), nil
}

const mavenRepoLocal = "deps"

func (r *Registry) pullScript(label string) string {
	cmd := new(strings.Builder)
	cmd.WriteString(fmt.Sprintf(`mvn dependency:get -Dmaven.repo.local=%s -Dtransitive=false -Dartifact=$PACKAGE_NAME:$PACKAGE_VERSION -DremoteRepositories=%s::::%s`, mavenRepoLocal, label, r.pkgsImport.Source.URL))

	if r.pkgsImport.Source.Credentials.Token != "" {
		cmd.WriteString(fmt.Sprintf(" -s %s", settingsFile))
	}

	return cmd.String()
}

func (r *Registry) cdIntoPackageDirectory() []string {
	return []string{
		`pkg_dir=$(echo $PACKAGE_NAME | cut -d ":" -f 1 | tr "." "/")/$(echo $PACKAGE_NAME | cut -d ":" -f 2)/$(echo $PACKAGE_VERSION | cut -d ":" -f 1)`,
		fmt.Sprintf(`cd $(find %s -path "*/$pkg_dir")`, mavenRepoLocal),
	}
}

func (r *Registry) pushScript(label string) string {
	return fmt.Sprintf(`mvn deploy:deploy-file -Durl=%s -DrepositoryId=%s -Dfile="$(find . -type f -name "*.$PACKAGE_PACKAGING")" -Dpackaging="$PACKAGE_PACKAGING" -DpomFile=$(ls *.pom | head -n 1) -s %s`, r.pkgsImport.Destination.URL, label, settingsFile)
}

func (r *Registry) validate(importName string) error {
	if err := r.validateCredentials(r.pkgsImport.Source.Credentials); err != nil {
		return err
	}

	if err := r.validateCredentials(r.pkgsImport.Destination.Credentials); err != nil {
		return err
	}

	return r.validatePackages(importName)
}

func (r *Registry) validatePackages(importName string) error {
	packagesMap, err := config.GetPackagesMap(importName)
	if err != nil {
		return err
	}

	for packageName, packageVersions := range packagesMap {
		if err := r.validatePackageName(packageName); err != nil {
			return err
		}
		for _, packageVersion := range packageVersions {
			if err := r.validatePackageVersion(packageVersion); err != nil {
				return err
			}
		}
	}

	return nil
}

var mavenGroupAndArtifactRegexp = regexp.MustCompile(`^.+:.+$`)

func (r *Registry) validatePackageName(name string) error {
	if !mavenGroupAndArtifactRegexp.MatchString(name) {
		return fmt.Errorf("%s is an invalid Maven package name. It must contain : between the group ID and the artifact ID.", name)
	}
	return nil
}

var validPackagings = []string{"pom", "jar", "maven-plugin", "ejb", "war", "ear", "rar", "aar"}

func (r *Registry) validatePackageVersion(version string) error {
	count := strings.Count(version, mavenCoordinatesSeparator)
	if count == 0 {
		return nil
	}

	if count == 1 {
		packaging := strings.Split(version, mavenCoordinatesSeparator)[1]
		if slices.Contains(validPackagings, packaging) {
			return nil
		}
		return fmt.Errorf("%s is an invalid Maven packaging string. It must be one of : %s.", packaging, validPackagings)
	}

	return fmt.Errorf("%s is an invalid Maven version string. It must be in the form of : version[:packaging].", version)
}

var errInvalidCredentials = errors.New("Maven credentials require a token and a username or a token and a header_name for authenticated registries")

func (r *Registry) validateCredentials(credentials config.Credentials) error {
	if credentials.Token == "" || credentials.AdditionalParameters["username"] != "" || credentials.AdditionalParameters["header_name"] != "" {
		return nil
	}

	return errInvalidCredentials
}
