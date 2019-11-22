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
		expectDpkgInstalled("rpcbind", "0.2.3\\-0.2")
		expectDpkgInstalled("keyutils", "1.5.9\\-8ubuntu1")
		expectDpkgInstalled("libevent-2.0-5", "2.0.21\\-stable\\-2ubuntu0.16.04.1")
		expectDpkgInstalled("libnfsidmap2", "0.25\\-5")
		expectDpkgInstalled("libkrb5support0", "1.13.2\\+dfsg\\-5ubuntu2.1")
		expectDpkgInstalled("libk5crypto3", "1.13.2\\+dfsg\\-5ubuntu2.1")
		expectDpkgInstalled("libkrb5-3", "1.13.2\\+dfsg\\-5ubuntu2.1")
		expectDpkgInstalled("libgssapi-krb5-2", "1.13.2\\+dfsg\\-5ubuntu2.1")
		expectDpkgInstalled("nfs-common", "1:1.2.8\\-9ubuntu12.1")

		state := findProcessState("nfsv3driver")

		Expect(state).To(Equal("running"))
	})

	Context("when an existing package is already installed", func() {
		BeforeEach(func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo dpkg -P nfs-common")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			cmd = exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo apt-get install -y nfs-common")
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
		})

		It("should upgrade the nfs-common package to the version specified in the pre-install script", func() {
			expectDpkgInstalled("nfs-common", "1:1.2.8\\-9ubuntu12.2")

			cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo /var/vcap/jobs/nfsv3driver/bin/pre-start")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			expectDpkgInstalled("nfs-common", "1:1.2.8\\-9ubuntu12.1")
		})
	})

	Context("nfsv3driver drain", func() {
		It("should successfully drain", func() {
			By("bosh stopping the nfsv3driver")
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "stop", "-n", "nfsv3driver")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
		})

		Context("when nfsv3driver is not reachable", func() {
			BeforeEach(func() {
				By("drain cannot reach the nfsv3driver")
				cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo iptables -t filter -A OUTPUT -p tcp --dport 7590  -j DROP")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
			})

			AfterEach(func() {
				cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo iptables -t filter -D OUTPUT -p tcp --dport 7590  -j DROP")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))

				cmd = exec.Command("bosh", "-d", "bosh_release_test", "start", "-n", "nfsv3driver")
				session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
			})

			It("should successfully drain", func() {
				cmd := exec.Command("bosh", "-d", "bosh_release_test", "stop", "-n", "nfsv3driver")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
			})
		})
	})
})

func expectDpkgInstalled(dpkgName string, version string) {
	packageDebugMessage := fmt.Sprintf("Expecting dpkg %s %s to be installed", dpkgName, version)
	By(packageDebugMessage)

  cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf("dpkg -s %s | grep Version | grep -o '[0-9].*' | grep -E '^%s$'", dpkgName, version))
  session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
  Expect(err).NotTo(HaveOccurred(), packageDebugMessage)
  Eventually(session).Should(gexec.Exit(0), packageDebugMessage + string(session.Out.Contents()))
}
