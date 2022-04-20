package phpmemcachedhandler

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type MemcachedConfigWriter struct {
	logger scribe.Emitter
}

func NewMemcachedConfigWriter(logger scribe.Emitter) MemcachedConfigWriter {
	return MemcachedConfigWriter{
		logger: logger,
	}
}

func (c MemcachedConfigWriter) Write(memcachedConfig MemcachedConfig, layerPath, cnbPath string) (string, error) {
	tmpl, err := template.New("php-memcached.ini").ParseFiles(filepath.Join(cnbPath, "config", "php-memcached.ini"))
	if err != nil {
		return "", fmt.Errorf("failed to parse PHP memcached config template: %w", err)
	}

	c.logger.Debug.Subprocess("Including memcached servers: %s", memcachedConfig.Servers)
	c.logger.Debug.Subprocess("Including memcached username: %s", memcachedConfig.Username)
	if memcachedConfig.Password != "" {
		c.logger.Debug.Subprocess("Including a non-empty password")
	}

	// Configuration set by this buildpack
	var b bytes.Buffer
	err = tmpl.Execute(&b, memcachedConfig)
	if err != nil {
		// not tested
		return "", err
	}

	f, err := os.OpenFile(filepath.Join(layerPath, "php-memcached.ini"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, &b)
	if err != nil {
		// not tested
		return "", err
	}

	return f.Name(), nil
}
