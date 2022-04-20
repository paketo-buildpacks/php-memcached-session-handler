package fakes

import (
	"sync"

	phpmemcachedhandler "github.com/paketo-buildpacks/php-memcached-session-handler"
)

type ConfigParser struct {
	ParseCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Dir string
		}
		Returns struct {
			MemcachedConfig phpmemcachedhandler.MemcachedConfig
			Error           error
		}
		Stub func(string) (phpmemcachedhandler.MemcachedConfig, error)
	}
}

func (f *ConfigParser) Parse(param1 string) (phpmemcachedhandler.MemcachedConfig, error) {
	f.ParseCall.mutex.Lock()
	defer f.ParseCall.mutex.Unlock()
	f.ParseCall.CallCount++
	f.ParseCall.Receives.Dir = param1
	if f.ParseCall.Stub != nil {
		return f.ParseCall.Stub(param1)
	}
	return f.ParseCall.Returns.MemcachedConfig, f.ParseCall.Returns.Error
}
