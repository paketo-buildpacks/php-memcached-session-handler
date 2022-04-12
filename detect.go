package phpmemcachedhandler

import (
	"fmt"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
)

//go:generate faux --interface DetectBindingResolver --output fakes/detect_binding_resolver.go
type BuildPlanMetadata struct {
	Launch bool
}

type DetectBindingResolver interface {
	Resolve(typ, provider, platformDir string) ([]servicebindings.Binding, error)
}

func Detect(bindingResolver DetectBindingResolver) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		memcachedBindings, err := bindingResolver.Resolve(MemcachedBindingType, "", context.Platform.Path)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if len(memcachedBindings) < 1 {
			return packit.DetectResult{}, packit.Fail.WithMessage(fmt.Sprintf("no service bindings of type `%s` provided", MemcachedBindingType))
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "php",
						Metadata: BuildPlanMetadata{
							Launch: true,
						},
					},
				},
				Provides: []packit.BuildPlanProvision{},
			},
		}, nil
	}
}
