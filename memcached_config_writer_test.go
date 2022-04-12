package phpmemcachedhandler_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/scribe"
	phpmemcachedhandler "github.com/paketo-buildpacks/php-memcached-session-handler"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testMemcachedConfigWriter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layerDir              string
		cnbDir                string
		memcachedConfig       phpmemcachedhandler.MemcachedConfig
		memcachedConfigWriter phpmemcachedhandler.MemcachedConfigWriter
	)

	it.Before(func() {
		var err error
		layerDir, err = os.MkdirTemp("", "php-memcached-layer")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chmod(layerDir, os.ModePerm)).To(Succeed())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(cnbDir, "config"), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(cnbDir, "config", "php-memcached.ini"), []byte(`
session.save_path = "{{.Servers}}"
memcached.sess_sasl_username="{{.Username}}"
memcached.sess_sasl_password="{{.Password}}"
`), os.ModePerm)).To(Succeed())

		memcachedConfig = phpmemcachedhandler.MemcachedConfig{
			Servers:  "some-servers",
			Username: "some-username",
			Password: "some-password",
		}
		logEmitter := scribe.NewEmitter(bytes.NewBuffer(nil))
		memcachedConfigWriter = phpmemcachedhandler.NewMemcachedConfigWriter(logEmitter)
	})

	it.After(func() {
		Expect(os.RemoveAll(layerDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("writes a memcached config ini file into the memcached config layer", func() {
		memcachedConfigFilePath, err := memcachedConfigWriter.Write(memcachedConfig, layerDir, cnbDir)
		Expect(err).NotTo(HaveOccurred())

		Expect(memcachedConfigFilePath).To(Equal(filepath.Join(layerDir, "php-memcached.ini")))

		contents, err := os.ReadFile(memcachedConfigFilePath)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(contents)).To(ContainSubstring(`session.save_path = "some-servers"`))
	})

	context("when there is no username", func() {
		it.Before(func() {
			memcachedConfig.Username = ""
		})

		it("writes a memcached config ini file into the memcached config layer with an empty string username", func() {
			memcachedConfigFilePath, err := memcachedConfigWriter.Write(memcachedConfig, layerDir, cnbDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(memcachedConfigFilePath).To(Equal(filepath.Join(layerDir, "php-memcached.ini")))

			contents, err := os.ReadFile(memcachedConfigFilePath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(ContainSubstring(`memcached.sess_sasl_username=""`))
		})
	})

	context("when there is no password", func() {
		it.Before(func() {
			memcachedConfig.Password = ""
		})

		it("writes a memcached config ini file into the memcached config layer with an empty string password", func() {
			memcachedConfigFilePath, err := memcachedConfigWriter.Write(memcachedConfig, layerDir, cnbDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(memcachedConfigFilePath).To(Equal(filepath.Join(layerDir, "php-memcached.ini")))

			contents, err := os.ReadFile(memcachedConfigFilePath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(ContainSubstring(`memcached.sess_sasl_password=""`))
		})
	})

	context("failure cases", func() {
		context("when template is not parseable", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(cnbDir, "config", "php-memcached.ini"), []byte(`{{.`), os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := memcachedConfigWriter.Write(memcachedConfig, layerDir, cnbDir)
				Expect(err).To(MatchError(ContainSubstring("failed to parse PHP memcached config template")))
			})
		})

		context("when memcached config file can't be opened for writing", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(layerDir, "php-memcached.ini"), nil, 0400)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := memcachedConfigWriter.Write(memcachedConfig, layerDir, cnbDir)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
