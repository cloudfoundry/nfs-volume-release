package main_test

import (
	"encoding/json"
	"testing"

	"code.cloudfoundry.org/cf-networking-helpers/portauthority"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func TestNfsV3Driver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NFS V3 Main Suite")
}

var (
	driverPath            string
	portAllocator         portauthority.PortAllocator
	listenPort, adminPort uint16
)

var _ = SynchronizedBeforeSuite(func() []byte {
	path, err := Build("code.cloudfoundry.org/nfsv3driver/cmd/nfsv3driver")
	Expect(err).ToNot(HaveOccurred())

	payload, err := json.Marshal(map[string]string{
		"driver_path": path,
	})

	Expect(err).NotTo(HaveOccurred())
	return payload

}, func(payload []byte) {

	context := map[string]string{}

	err := json.Unmarshal(payload, &context)
	Expect(err).NotTo(HaveOccurred())

	driverPath = context["driver_path"]

	node := GinkgoParallelProcess()

	startPort := 1070 * node
	portRange := 50
	endPort := startPort + portRange

	portAllocator, err = portauthority.New(startPort, endPort)
	Expect(err).NotTo(HaveOccurred())

	listenPort, err = portAllocator.ClaimPorts(1)
	Expect(err).NotTo(HaveOccurred())

	adminPort, err = portAllocator.ClaimPorts(1)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	CleanupBuildArtifacts()
})
