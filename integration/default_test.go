package integration_test

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
		source string
		name   string
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image              occam.Image
			container          occam.Container
			memcachedContainer occam.Container
			binding            string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "default_app"))
			Expect(err).NotTo(HaveOccurred())
			binding = filepath.Join(source, "binding")

			memcachedContainer, err = docker.Container.Run.
				WithPublish("11211").
				Execute("memcached")
			Expect(err).NotTo(HaveOccurred())

			ipAddress, err := memcachedContainer.IPAddressForNetwork("bridge")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(filepath.Join(source, "binding", "servers"), []byte(ipAddress), os.ModePerm)).To(Succeed())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(memcachedContainer.ID)).To(Succeed())
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("sets up a memcached session handler for PHP", func() {
			var (
				logs fmt.Stringer
				err  error
			)

			image, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					phpBuildpack,
					buildpack,
					phpBuiltinServerBuildpack,
				).
				WithEnv(map[string]string{
					"BP_PHP_WEB_DIR":       "htdocs",
					"BP_LOG_LEVEL":         "DEBUG",
					"SERVICE_BINDING_ROOT": "/bindings",
				}).
				WithVolumes(fmt.Sprintf("%s:/bindings/php-memcached-session", binding)).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Getting the layer associated with the memcached configuration",
				fmt.Sprintf("    /layers/%s/php-memcached-config", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
			))
			Expect(logs).To(ContainLines(
				"  Resolving the php-memcached-session service binding",
			))
			Expect(logs).To(ContainLines(
				"  Parsing the php-memcached-session service binding",
			))

			ipAddress, err := memcachedContainer.IPAddressForNetwork("bridge")
			Expect(err).NotTo(HaveOccurred())

			Expect(logs).To(ContainLines(
				"  Writing the memcached configuration",
				fmt.Sprintf("    Including memcached servers: %s", ipAddress),
				"    Including memcached username: ",
				fmt.Sprintf("    Memcached configuration written to: /layers/%s/php-memcached-config/php-memcached.ini", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
			))
			Expect(logs).To(ContainLines(
				"  Configuring launch environment",
				MatchRegexp(fmt.Sprintf(`    PHP_INI_SCAN_DIR -> "\$PHP_INI_SCAN_DIR:\/layers\/%s\/php-memcached-config"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))),
			))

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			jar, err := cookiejar.New(nil)
			Expect(err).NotTo(HaveOccurred())

			client := &http.Client{
				Jar: jar,
			}

			Eventually(container).Should(Serve(ContainSubstring("1")).WithClient(client).OnPort(8080).WithEndpoint("/index.php"))
			Eventually(container).Should(Serve(ContainSubstring("2")).WithClient(client).OnPort(8080).WithEndpoint("/index.php"))
		})
	})
}
