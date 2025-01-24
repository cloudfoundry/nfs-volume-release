package perf_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	perf "code.cloudfoundry.org/mapfs-performance-acceptance-tests"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestPerf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mapfs Performance Acceptance Tests Suite")
}

var config perf.Config

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Second)

	var err error
	config, err = perf.LoadConfig()
	Expect(err).NotTo(HaveOccurred())

})

func writeDataToMountedDirectory(directory string) {
	nativeWriteFile, err := os.CreateTemp(directory, "perf_file")
	Expect(err).NotTo(HaveOccurred())

	defer func() {
		err := nativeWriteFile.Close()
		Expect(err).NotTo(HaveOccurred())
	}()

	cmd := exec.Command("dd", "if=/dev/zero", "bs=16k", numWrites, fmt.Sprintf("of=%s", nativeWriteFile.Name()))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 10*time.Second).Should(gexec.Exit(0), string(session.Out.Contents()))
}

func shapeTraffic(delayInMs string) {
	unshapeTraffic()
	tcHighLatencyCmd := exec.Command("tc", "qdisc", "add", "dev", "eth0", "root", "netem", "delay", delayInMs)
	session, err := gexec.Start(tcHighLatencyCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func unshapeTraffic() {
	tcHighLatencyCmd := exec.Command("tc", "qdisc", "del", "dev", "eth0", "root", "netem")
	session, err := gexec.Start(tcHighLatencyCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	tcCheckCmd := exec.Command("tc", "qdisc")
	session, err = gexec.Start(tcCheckCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	GinkgoWriter.Write(session.Out.Contents())
}
