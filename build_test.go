package phpmemcachedhandler_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/paketo-buildpacks/packit/v2/servicebindings"
	phpmemcachedhandler "github.com/paketo-buildpacks/php-memcached-session-handler"
	"github.com/paketo-buildpacks/php-memcached-session-handler/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layerDir   string
		workingDir string
		cnbDir     string

		buffer               *bytes.Buffer
		configParser         *fakes.ConfigParser
		buildBindingResolver *fakes.BuildBindingResolver
		configWriter         *fakes.ConfigWriter

		parsedMemcachedConfig phpmemcachedhandler.MemcachedConfig

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layerDir, err = os.MkdirTemp("", "layer")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		buffer = bytes.NewBuffer(nil)
		logEmitter := scribe.NewEmitter(buffer)

		configParser = &fakes.ConfigParser{}
		buildBindingResolver = &fakes.BuildBindingResolver{}
		configWriter = &fakes.ConfigWriter{}

		buildBindingResolver.ResolveOneCall.Returns.Binding = servicebindings.Binding{
			Path: "some-binding-path",
		}

		parsedMemcachedConfig = phpmemcachedhandler.MemcachedConfig{
			Servers:  "some-servers",
			Username: "some-username",
			Password: "some-password",
		}

		configParser.ParseCall.Returns.MemcachedConfig = parsedMemcachedConfig

		build = phpmemcachedhandler.Build(configParser, buildBindingResolver, configWriter, logEmitter)
	})

	it.After(func() {
		Expect(os.RemoveAll(layerDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("writes a memcached configuration for Php", func() {
		result, err := build(packit.BuildContext{
			Layers: packit.Layers{
				Path: layerDir,
			},
			Platform: packit.Platform{
				Path: "some-platform-path",
			},
			CNBPath: "some-cnb-path",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		layer := result.Layers[0]

		Expect(layer.Name).To(Equal("php-memcached-config"))
		Expect(layer.Path).To(Equal(filepath.Join(layerDir, "php-memcached-config")))
		Expect(layer.Launch).To(BeTrue())
		Expect(layer.LaunchEnv).To(Equal(packit.Environment{
			"PHP_INI_SCAN_DIR.append": filepath.Join(layerDir, "php-memcached-config"),
			"PHP_INI_SCAN_DIR.delim":  ":",
		}))

		Expect(buildBindingResolver.ResolveOneCall.Receives.Typ).To(Equal("php-memcached-session"))
		Expect(buildBindingResolver.ResolveOneCall.Receives.Provider).To(Equal(""))
		Expect(buildBindingResolver.ResolveOneCall.Receives.PlatformDir).To(Equal("some-platform-path"))

		Expect(configParser.ParseCall.Receives.Dir).To(Equal("some-binding-path"))

		Expect(configWriter.WriteCall.Receives.MemcachedConfig).To(Equal(parsedMemcachedConfig))
		Expect(configWriter.WriteCall.Receives.LayerPath).To(Equal(filepath.Join(layerDir, "php-memcached-config")))
		Expect(configWriter.WriteCall.Receives.CnbPath).To(Equal("some-cnb-path"))
	})

	context("failure cases", func() {
		context("when the memcached layer cannot be retrieved", func() {
			it.Before(func() {
				Expect(os.Chmod(layerDir, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					Layers: packit.Layers{
						Path: layerDir,
					},
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the memcached layer cannot be reset", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(layerDir, "php-memcached-config", "some-dir"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(filepath.Join(layerDir, "php-memcached-config"), 0500)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(layerDir, "php-memcached-config"), os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					Layers: packit.Layers{
						Path: layerDir,
					},
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the memcached binding cannot be resolved", func() {
			it.Before(func() {
				buildBindingResolver.ResolveOneCall.Returns.Error = errors.New("failed to resolve php-memcached-session binding")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					Layers: packit.Layers{
						Path: layerDir,
					},
				})
				Expect(err).To(MatchError("failed to resolve php-memcached-session binding"))
			})
		})

		context("when the memcached binding cannot be parsed", func() {
			it.Before(func() {
				configParser.ParseCall.Returns.Error = errors.New("failed to parse binding")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					Layers: packit.Layers{
						Path: layerDir,
					},
				})
				Expect(err).To(MatchError("failed to parse binding"))
			})
		})

		context("when the memcached configuration cannot be written", func() {
			it.Before(func() {
				configWriter.WriteCall.Returns.Error = errors.New("failed to write config")
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					Layers: packit.Layers{
						Path: layerDir,
					},
				})
				Expect(err).To(MatchError("failed to write config"))
			})
		})
	})
}
