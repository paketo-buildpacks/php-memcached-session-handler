package phpmemcachedhandler

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
)

//go:generate faux --interface BuildBindingResolver --output fakes/build_binding_resolver.go
//go:generate faux --interface ConfigParser --output fakes/config_parser.go
//go:generate faux --interface ConfigWriter --output fakes/config_writer.go

type BuildBindingResolver interface {
	ResolveOne(typ, provider, platformDir string) (servicebindings.Binding, error)
}

type ConfigParser interface {
	Parse(dir string) (MemcachedConfig, error)
}

type ConfigWriter interface {
	Write(memcachedConfig MemcachedConfig, layerPath, cnbPath string) (string, error)
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
func Build(memcachedBindingConfigParser ConfigParser, bindingResolver BuildBindingResolver, memcachedConfigWriter ConfigWriter, logger scribe.Emitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		logger.Debug.Process("Getting the layer associated with the memcached configuration")
		phpMemcachedLayer, err := context.Layers.Get(PhpMemcachedLayer)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Subprocess(phpMemcachedLayer.Path)
		logger.Debug.Break()

		phpMemcachedLayer, err = phpMemcachedLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Debug.Process("Resolving the %s service binding", MemcachedBindingType)
		binding, err := bindingResolver.ResolveOne(MemcachedBindingType, "", context.Platform.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Break()

		logger.Debug.Process("Parsing the %s service binding", MemcachedBindingType)
		memcachedConfig, err := memcachedBindingConfigParser.Parse(binding.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Debug.Break()

		// Use go templating to write the config file
		logger.Process("Writing the memcached configuration")
		memcachedConfigPath, err := memcachedConfigWriter.Write(memcachedConfig, phpMemcachedLayer.Path, context.CNBPath)
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Subprocess("Memcached configuration written to: %s", memcachedConfigPath)
		logger.Break()

		phpMemcachedLayer.LaunchEnv.Append("PHP_INI_SCAN_DIR",
			phpMemcachedLayer.Path,
			string(os.PathListSeparator),
		)
		logger.EnvironmentVariables(phpMemcachedLayer)

		phpMemcachedLayer.Launch = true

		return packit.BuildResult{
			Layers: []packit.Layer{phpMemcachedLayer},
		}, nil
	}
}
