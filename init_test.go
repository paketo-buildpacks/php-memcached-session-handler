package phpmemcachedhandler_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPhpMemcachedHandler(t *testing.T) {
	suite := spec.New("php-memcached-handler", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("MemcachedConfigParser", testMemcachedConfigParser)
	suite("MemcachedConfigWriter", testMemcachedConfigWriter)
	suite.Run(t)
}
