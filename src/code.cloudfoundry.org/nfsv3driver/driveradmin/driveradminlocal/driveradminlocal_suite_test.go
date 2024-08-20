package driveradminlocal_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLocalDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Driver Admin API Local Suite")
}
