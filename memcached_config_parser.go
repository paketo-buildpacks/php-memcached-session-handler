package phpmemcachedhandler

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/fs"
)

type MemcachedConfig struct {
	Servers  string
	Username string
	Password string
}

type MemcachedConfigParser struct {
}

func NewMemcachedConfigParser() MemcachedConfigParser {
	return MemcachedConfigParser{}
}

func (p MemcachedConfigParser) Parse(dir string) (MemcachedConfig, error) {
	serversFilepath := filepath.Join(dir, "servers")

	serversFileExists, err := fs.Exists(serversFilepath)
	if err != nil {
		return MemcachedConfig{}, err
	}

	servers := "127.0.0.1"
	if serversFileExists {
		serversBytes, err := os.ReadFile(serversFilepath)
		if err != nil {
			return MemcachedConfig{}, err
		}

		servers = strings.TrimSpace(string(serversBytes))
	}

	usernameFilepath := filepath.Join(dir, "username")

	usernameFileExists, err := fs.Exists(usernameFilepath)
	if err != nil {
		// untested
		return MemcachedConfig{}, err
	}

	username := ""
	if usernameFileExists {
		usernameBytes, err := os.ReadFile(usernameFilepath)
		if err != nil {
			return MemcachedConfig{}, err
		}

		username = strings.TrimSpace(string(usernameBytes))
	}

	passwordFilepath := filepath.Join(dir, "password")

	passwordFileExists, err := fs.Exists(passwordFilepath)
	if err != nil {
		// untested
		return MemcachedConfig{}, err
	}

	password := ""
	if passwordFileExists {
		passwordBytes, err := os.ReadFile(passwordFilepath)
		if err != nil {
			return MemcachedConfig{}, err
		}

		password = strings.TrimSpace(string(passwordBytes))
	}

	return MemcachedConfig{
		Servers:  servers,
		Username: username,
		Password: password,
	}, nil
}
