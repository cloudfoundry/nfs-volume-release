package bosh_release_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os/exec"
)

var _ = Describe("BoshReleaseTest", func() {
	BeforeEach(func() {
		deploy()
	})

	It("should have a nfsv3driver process running", func() {
		state := findProcessState("nfsv3driver")

		Expect(state).To(Equal("running"))
	})

	Context("when nfsv3driver is disabled", func() {

		BeforeEach(func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "delete-deployment", "-n")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			deploy("./operations/disable-nfsv3driver.yml")
		})


		It("should not install packages or run rpcbind", func() {
			expectDpkgNotInstalled("rpcbind")
			expectDpkgNotInstalled("keyutils")
			expectDpkgNotInstalled("libevent-2.0-5")
			expectDpkgNotInstalled("libnfsidmap2")
			expectDpkgNotInstalled("nfs-common")

			Expect(findProcessState("nfsv3driver")).To(Equal(""))
		})
	})
})

func expectDpkgNotInstalled(dpkgName string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf( "dpkg -l | grep ' %s '", dpkgName))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(1), string(session.Out.Contents()))
}
