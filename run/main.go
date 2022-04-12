package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
	phpmemcachedhandler "github.com/paketo-buildpacks/php-memcached-session-handler"
)

func main() {
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	serviceResolver := servicebindings.NewResolver()

	packit.Run(
		phpmemcachedhandler.Detect(
			serviceResolver,
		),
		phpmemcachedhandler.Build(
			phpmemcachedhandler.NewMemcachedConfigParser(),
			serviceResolver,
			phpmemcachedhandler.NewMemcachedConfigWriter(logEmitter),
			logEmitter,
		),
	)
}
