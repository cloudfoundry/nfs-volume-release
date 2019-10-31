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

	Context("when credhub has been set to zero instances", func() {
		It("should succeed deploying", func() {
			session, err := deploy("./operations/scale-credhub-to-zero.yml")
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Context("missing a credhub", func() {
		It("should succeed deploying", func() {
			session, err := deploy("./operations/remove-credhub.yml")
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	It("should have a nfsv3driver process running", func() {
		expectDpkgInstalled("rpcbind", "0.2.3-0.2")
		expectDpkgInstalled("keyutils", "1.5.9-8ubuntu1")
		expectDpkgInstalled("libevent-2.1-6", "2.1.8-stable-4build1")
		expectDpkgInstalled("libnfsidmap2", "0.25-5")
		expectDpkgInstalled("libkrb5support0", "1.15.1-2")
		expectDpkgInstalled("libk5crypto3", "1.15.1-2")
		expectDpkgInstalled("libkrb5-3", "1.15.1-2")
		expectDpkgInstalled("libgssapi-krb5-2", "1.15.1-2")
		expectDpkgInstalled("nfs-common", "1:1.3.4-2.1ubuntu4")

		state := findProcessState("nfsv3driver")

		Expect(state).To(Equal("running"))
	})

	Context("when an existing package is already installed", func() {
		BeforeEach(func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo dpkg -P nfs-common")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			cmd = exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo apt-get update && sudo apt-get install -y nfs-common")
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
		})

		It("should upgrade the nfs-common package to the version specified in the pre-install script", func() {
			expectDpkgInstalled("nfs-common", "1:1.2.8-9ubuntu12.2")

			cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo /var/vcap/jobs/nfsv3driver/bin/pre-start")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			expectDpkgInstalled("nfs-common", "1:1.3.4-2.1ubuntu4")
		})
	})

	Context("when nfsv3driver is disabled", func() {
		BeforeEach(func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "delete-deployment", "-n")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			ensureDeploy("./operations/disable-nfsv3driver.yml")
		})


		It("should not install packages or run rpcbind", func() {
			expectDpkgNotInstalled("rpcbind")
			expectDpkgNotInstalled("keyutils")
			expectDpkgNotInstalled("libevent-2.1-6")
			expectDpkgNotInstalled("libnfsidmap2")
			expectDpkgNotInstalled("nfs-common")

			Expect(findProcessState("nfsv3driver")).To(Equal(""))
		})
	})

	Context("when another process has a dpkg lock", func() {

		BeforeEach(func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo dpkg -P nfs-common")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			cmd = exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo rm -f /tmp/lock_dpkg")
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			cmd = exec.Command("bosh", "-d", "bosh_release_test", "scp", dpkgLockBuildPackagePath, "nfsv3driver:/tmp/lock_dpkg")
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))

			cmd = exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo /tmp/lock_dpkg")
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("locked /var/lib/dpkg/lock"))
		})

		AfterEach(func() {
			releaseDpkgLock()
		})

		It("should successfully dpkg install", func() {
			unstubSleep()

			cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo /var/vcap/jobs/nfsv3driver/bin/pre-start")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("dpkg: error: dpkg status database is locked by another process"))
			releaseDpkgLock()
			Eventually(session).Should(gexec.Exit(0))
		})

		It("should eventually timeout when the dpkg lock is not released in a reasonable time", func() {
			cmd := exec.Command("bosh", "-d", "bosh_release_test", "ssh", "nfsv3driver", "-c", "sudo /var/vcap/jobs/nfsv3driver/bin/pre-start")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("dpkg: error: dpkg status database is locked by another process"))
			Eventually(session, 6 * time.Minute, 1 * time.Second).Should(gexec.Exit(1))
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
