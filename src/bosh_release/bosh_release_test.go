package bosh_release_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"time"
)

var _ = Describe("BoshReleaseTest", func() {
	BeforeEach(func() {
		ensureDeploy()

		By("bosh -d bosh_release_test start -n nfsv3driver", func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "start", "-n", "nfsv3driver")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
		})

		stubSleep()
	})

	AfterEach(func() {
		unstubSleep()
	})

	Context("with no other cloud foundry control components", func() {
		It("should succeed deploying", func() {
			session, err := deploy("./operations/remove-credhub.yml", "./operations/remove-nfsbrokerpush.yml")
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	It("should have a nfsv3driver process running", func() {
		state := findProcessState("nfsv3driver")

		Expect(state).To(Equal("running"))
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

		Context("when the rep process takes longer than 15 minutes to exit", func() {
			BeforeEach(func() {

				By("bosh -d bosh_release_test scp"+repBuildPackagePath+"nfsv3driver:/tmp/rep", func() {
					cmd := exec.Command("bosh", "-d", "bosh_release_test", "scp", repBuildPackagePath, "nfsv3driver:/tmp/rep")
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
				})

				By("bosh -d bosh_release_test ssh nfsv3driver -c sudo chmod +x /tmp/rep && sudo mv /tmp/rep /bin/rep", func() {
					cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo chmod +x /tmp/rep && sudo mv /tmp/rep /bin/rep")
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
				})

				By("bosh -d bosh_release_test ssh nfsv3driver -c rep", func() {
					cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "rep")
					_, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			AfterEach(func() {
				By("bosh -d bosh_release_test ssh nfsv3driver -c sudo pkill -f 'rep'", func() {
					cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo pkill -f 'rep'")
					session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
				})
			})

			It("should timeout and fail drain", func() {
				By("stopping nfsv3driver")
				cmd := exec.Command("bosh", "-d", "bosh_release_test", "stop", "-n", "nfsv3driver")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session.Out, 16*time.Minute).Should(gbytes.Say("drain scripts failed. Failed Jobs: nfsv3driver"))
				Eventually(session, 16*time.Minute).Should(gexec.Exit(1), string(session.Out.Contents()))
			})
		})
	})
})

func unstubSleep() {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo rm -f /usr/bin/sleep")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
}

func stubSleep() {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo touch /usr/bin/sleep && sudo chmod +x /usr/bin/sleep")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
}

func releaseDpkgLock() {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", "sudo pkill lock_dpkg")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func expectDpkgNotInstalled(dpkgName string) {
	cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "-c", fmt.Sprintf( "dpkg -l | grep ' %s '", dpkgName))
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(1), string(session.Out.Contents()))
}

func expectDpkgInstalled(dpkgName string, version string) {
	packageDebugMessage := fmt.Sprintf("Expecting dpkg %s %s to be installed", dpkgName, version)
	By(packageDebugMessage)

  cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", fmt.Sprintf("dpkg -s %s | grep Version | grep -o '[0-9].*' | grep -E '^%s$'", dpkgName, version))
  session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
  Expect(err).NotTo(HaveOccurred(), packageDebugMessage)
  Eventually(session).Should(gexec.Exit(0), packageDebugMessage + string(session.Out.Contents()))
}
