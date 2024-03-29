package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/onsi/gomega/format"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	buildpack                 string
	phpBuildpack              string
	offlinePhpBuildpack       string
	phpBuiltinServerBuildpack string

	root string

	buildpackInfo struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}
)

func TestIntegration(t *testing.T) {
	docker := occam.NewDocker()
	Expect := NewWithT(t).Expect

	format.MaxLength = 0

	var config struct {
		Php              string `json:"php"`
		PhpBuiltinServer string `json:"php-builtin-server"`
	}

	file, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()

	Expect(json.NewDecoder(file).Decode(&config)).To(Succeed())

	file, err = os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&buildpackInfo)
	Expect(err).NotTo(HaveOccurred())

	root, err = filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	buildpackStore := occam.NewBuildpackStore()

	buildpack, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	phpBuildpack, err = buildpackStore.Get.
		Execute(config.Php)
	Expect(err).NotTo(HaveOccurred())

	offlinePhpBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		Execute(config.Php)
	Expect(err).NotTo(HaveOccurred())

	phpBuiltinServerBuildpack, err = buildpackStore.Get.
		Execute(config.PhpBuiltinServer)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault)
	suite("Offline", testOffline)
	suite("TestReproducibleLayerRebuild", testReproducibleLayerRebuild)
	suite.Run(t)

	Expect(docker.Image.Remove.WithForce().Execute("memcached:latest")).To(Succeed())
}
