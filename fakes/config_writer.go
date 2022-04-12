package fakes

import (
	"sync"

	phpmemcachedhandler "github.com/paketo-buildpacks/php-memcached-session-handler"
)

type ConfigWriter struct {
	WriteCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			MemcachedConfig phpmemcachedhandler.MemcachedConfig
			LayerPath       string
			CnbPath         string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(phpmemcachedhandler.MemcachedConfig, string, string) (string, error)
	}
}

func (f *ConfigWriter) Write(param1 phpmemcachedhandler.MemcachedConfig, param2 string, param3 string) (string, error) {
	f.WriteCall.mutex.Lock()
	defer f.WriteCall.mutex.Unlock()
	f.WriteCall.CallCount++
	f.WriteCall.Receives.MemcachedConfig = param1
	f.WriteCall.Receives.LayerPath = param2
	f.WriteCall.Receives.CnbPath = param3
	if f.WriteCall.Stub != nil {
		return f.WriteCall.Stub(param1, param2, param3)
	}
	return f.WriteCall.Returns.String, f.WriteCall.Returns.Error
}
