// Package khulnasoft centralizes all the logic and structs needed to transform a configuration
// into a valid KhulnaSoft pipeline configuration file.
package khulnasoft

import (
	"fmt"
	"os"

	"github.com/khulnasoft/packages-registry/config"
	"github.com/khulnasoft/packages-registry/registry"
	"github.com/khulnasoft/packages-registry/util"
	"gopkg.in/yaml.v3"
)

// Generator is the builder object that from the configuration will
// create a CI pipeline configuration file. As such, it has a single
// exported function: Generate.
type Generator struct {
	config *config.Configuration
}

const fiveMegaBytes int64 = 5 * 1024 * 1024

// Generate will generate the CI pipeline yaml config file and write it to the
// passed os.File pointer.
func (g *Generator) Generate(file *os.File) error {
	if err := g.generateYamlConfig(file); err != nil {
		return err
	}

	return g.validate(file, fiveMegaBytes)
}

func (g *Generator) validate(file *os.File, maxSize int64) error {
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() >= maxSize {
		return fmt.Errorf("the generated config file is %d bytes which is over the limit for the KhulnaSoft engine", stat.Size())
	}
	return nil
}

func (g *Generator) generateYamlConfig(file *os.File) error {
	importsCount := len(g.config.Imports)
	pipeline := newPipeline(importsCount, importsCount)

	for _, importName := range util.OrderedMapKeysOf(g.config.Imports) {
		i := g.config.Imports[importName]
		pipeline.Stages = append(pipeline.Stages, importName)

		registry, err := registry.GetRegistry(i, importName)
		if err != nil {
			return err
		}

		scripts, err := registry.Scripts()
		if err != nil {
			return err
		}

		image := registry.ImageName()

		if len(i.Image) != 0 {
			image = i.Image
		}

		if err := pipeline.AddHiddenJob(importName, image, scripts); err != nil {
			return err
		}

		packagesMap, err := config.GetPackagesMap(importName)
		if err != nil {
			return err
		}

		for _, name := range util.OrderedMapKeysOf(packagesMap) {
			versions := packagesMap[name]
			for _, version := range versions {
				envVars := registry.AdditionalEnvVars(name, version)
				pipeline.AddJob(
					pipeline.withStage(importName),
					pipeline.withImage(image),
					pipeline.withPackageNameAndVersion(name, version),
					pipeline.withAdditionalEnvVariables(envVars),
				)
			}
		}
	}

	return g.marshalYaml(file, pipeline)
}

func (g *Generator) marshalYaml(file *os.File, pipeline *Pipeline) error {
	buff, err := yaml.Marshal(pipeline)
	if err != nil {
		return err
	}

	if _, err = file.WriteString(string(buff)); err != nil {
		return err
	}

	return nil
}

func NewGenerator(config *config.Configuration) *Generator {
	return &Generator{
		config: config,
	}
}
