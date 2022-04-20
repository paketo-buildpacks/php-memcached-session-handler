package phpmemcachedhandler_test

import (
	"os"
	"path/filepath"
	"testing"

	phpmemcachedhandler "github.com/paketo-buildpacks/php-memcached-session-handler"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testMemcachedConfigParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string

		parser phpmemcachedhandler.MemcachedConfigParser
	)

	it.Before(func() {
		var err error

		workingDir, err = os.MkdirTemp("", "workingDir")
		Expect(err).NotTo(HaveOccurred())

		parser = phpmemcachedhandler.NewMemcachedConfigParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("parses with default values", func() {
		config, err := parser.Parse(workingDir)
		Expect(err).NotTo(HaveOccurred())

		Expect(config).To(Equal(phpmemcachedhandler.MemcachedConfig{
			Servers:  "127.0.0.1",
			Username: "",
			Password: "",
		}))
	})

	context("when the servers file exists", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, "servers"), []byte("some-servers"), os.ModePerm)).To(Succeed())
		})

		it("uses the value from the servers file", func() {
			config, err := parser.Parse(workingDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(config.Servers).To(Equal("some-servers"))
		})

		context("when there is whitespace", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "servers"), []byte("  \tsome-servers\n\n"), os.ModePerm)).To(Succeed())
			})

			it("strips whitespace", func() {
				config, err := parser.Parse(workingDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(config.Servers).To(Equal("some-servers"))
			})
		})
	})

	context("when the username file exists", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, "username"), []byte("some-username"), os.ModePerm)).To(Succeed())
		})

		it("uses the value from the username file", func() {
			config, err := parser.Parse(workingDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(config.Username).To(Equal("some-username"))
		})

		context("when there is whitespace", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "username"), []byte("  \tsome-username\n\n"), os.ModePerm)).To(Succeed())
			})

			it("strips whitespace", func() {
				config, err := parser.Parse(workingDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(config.Username).To(Equal("some-username"))
			})
		})
	})

	context("when the password file exists", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(workingDir, "password"), []byte("some-password"), os.ModePerm)).To(Succeed())
		})

		it("uses the value from the password file", func() {
			config, err := parser.Parse(workingDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(config.Password).To(Equal("some-password"))
		})

		context("when there is whitespace", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "password"), []byte("  \tsome-password\n\n"), os.ModePerm)).To(Succeed())
			})

			it("strips whitespace", func() {
				config, err := parser.Parse(workingDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(config.Password).To(Equal("some-password"))
			})
		})
	})

	context("failure cases", func() {
		context("when there is an error determining if the files exist", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := parser.Parse(workingDir)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when there is an error reading the servers file", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "servers"), []byte("some-servers"), 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := parser.Parse(workingDir)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when there is an error reading the username file", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "username"), []byte("some-username"), 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := parser.Parse(workingDir)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when there is an error reading the password file", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "password"), []byte("some-password"), 0000)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := parser.Parse(workingDir)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
