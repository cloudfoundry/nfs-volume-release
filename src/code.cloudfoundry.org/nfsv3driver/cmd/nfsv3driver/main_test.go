package main_test

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	var (
		session                *gexec.Session
		command                *exec.Cmd
		expectedStartOutput    string
		expectedStartErrOutput string
	)

	BeforeEach(func() {
		command = exec.Command(driverPath)
		expectedStartOutput = "started"
		expectedStartErrOutput = ""
	})

	JustBeforeEach(func() {
		var err error
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Eventually(session.Out).Should(gbytes.Say(expectedStartOutput))
		Eventually(session.Err).Should(gbytes.Say(expectedStartErrOutput))
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		session.Kill().Wait()
	})

	Context("with a driver path", func() {
		var dir string

		BeforeEach(func() {
			var err error
			dir, err = os.MkdirTemp("", "driversPath")
			Expect(err).ToNot(HaveOccurred())

			command.Args = append(command.Args, "-driversPath="+dir)
			command.Args = append(command.Args, "-transport=tcp-json")
		})

		It("listens on tcp/7589 by default", func() {
			EventuallyWithOffset(1, func() error {
				_, err := net.Dial("tcp", "0.0.0.0:7589")
				return err
			}, 5).ShouldNot(HaveOccurred())

			specFile := filepath.Join(dir, "nfsv3driver.json")
			specFileContents, err := os.ReadFile(specFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(specFileContents)).To(MatchJSON(`{
				"Name": "nfsv3driver",
				"Addr": "http://127.0.0.1:7589",
				"TLSConfig": null,
				"UniqueVolumeIds": true
			}`))
		})

		It("listens on tcp/7590 for admin reqs by default", func() {
			EventuallyWithOffset(1, func() error {
				_, err := net.Dial("tcp", "0.0.0.0:7590")
				return err
			}, 5).ShouldNot(HaveOccurred())
		})

		Context("when command line arguments are provided", func() {
			BeforeEach(func() {
				command.Args = append(command.Args, "-listenAddr=0.0.0.0:7591")
				command.Args = append(command.Args, "-adminAddr=0.0.0.0:7592")
			})

			It("listens on provided arguments", func() {
				EventuallyWithOffset(1, func() error {
					_, err := net.Dial("tcp", "0.0.0.0:7591")
					return err
				}, 5).ShouldNot(HaveOccurred())
			})

			It("listens on provided arguments", func() {
				EventuallyWithOffset(1, func() error {
					_, err := net.Dial("tcp", "0.0.0.0:7592")
					return err
				}, 5).ShouldNot(HaveOccurred())
			})

			Context("when they are invalid", func() {

				BeforeEach(func() {
					command.Args = []string{"invalidargs"}
					expectedStartOutput = "fatal-err-aborting"
				})

				It("should error", func() {
					EventuallyWithOffset(1, func() error {
						_, err := net.Dial("tcp", "0.0.0.0:7595")
						return err
					}, 5).Should(HaveOccurred())
				})
			})
		})

		Context("given correct LDAP arguments set in the environment", func() {
			BeforeEach(func() {
				Expect(os.Setenv("LDAP_SVC_USER", "user")).To(Succeed())
				Expect(os.Setenv("LDAP_SVC_PASS", "password")).To(Succeed())
				Expect(os.Setenv("LDAP_USER_FQDN", "cn=Users,dc=corp,dc=testdomain,dc=com")).To(Succeed())
				Expect(os.Setenv("LDAP_HOST", "ldap.testdomain.com")).To(Succeed())
				Expect(os.Setenv("LDAP_PORT", "7593")).To(Succeed())
				Expect(os.Setenv("LDAP_PROTO", "tcp")).To(Succeed())

				command.Args = append(command.Args, "-listenAddr=0.0.0.0:7593")
				command.Args = append(command.Args, "-adminAddr=0.0.0.0:7594")
			})

			AfterEach(func() {
				Expect(os.Unsetenv("LDAP_SVC_USER")).To(Succeed())
				Expect(os.Unsetenv("LDAP_SVC_PASS")).To(Succeed())
				Expect(os.Unsetenv("LDAP_USER_FQDN")).To(Succeed())
				Expect(os.Unsetenv("LDAP_HOST")).To(Succeed())
				Expect(os.Unsetenv("LDAP_PORT")).To(Succeed())
				Expect(os.Unsetenv("LDAP_PROTO")).To(Succeed())
			})

			It("listens on provided arguments", func() {
				EventuallyWithOffset(1, func() error {
					_, err := net.Dial("tcp", "0.0.0.0:7593")
					return err
				}, 5).ShouldNot(HaveOccurred())
			})
		})

		Context("given incomplete LDAP arguments set in the environment", func() {
			BeforeEach(func() {
				Expect(os.Setenv("LDAP_HOST", "ldap.testdomain.com")).To(Succeed())
				Expect(os.Setenv("LDAP_PORT", "389")).To(Succeed())
				Expect(os.Setenv("LDAP_PROTO", "tcp")).To(Succeed())
				command.Args = append(command.Args, "-listenAddr=0.0.0.0:7595")
				command.Args = append(command.Args, "-adminAddr=0.0.0.0:7596")
				expectedStartOutput = ""
				expectedStartErrOutput = "required LDAP parameters are not set"
			})

			AfterEach(func() {
				Expect(os.Unsetenv("LDAP_HOST")).To(Succeed())
				Expect(os.Unsetenv("LDAP_PORT")).To(Succeed())
				Expect(os.Unsetenv("LDAP_PROTO")).To(Succeed())
			})

			It("fails to start", func() {
				EventuallyWithOffset(1, func() error {
					_, err := net.Dial("tcp", "0.0.0.0:7595")
					return err
				}, 5).Should(HaveOccurred())
			})
		})

		Context("given LDAP_TIMEOUT are set in the the environment", func() {
			BeforeEach(func() {
				Expect(os.Setenv("LDAP_SVC_USER", "user")).To(Succeed())
				Expect(os.Setenv("LDAP_SVC_PASS", "password")).To(Succeed())
				Expect(os.Setenv("LDAP_USER_FQDN", "cn=Users,dc=corp,dc=testdomain,dc=com")).To(Succeed())
				Expect(os.Setenv("LDAP_HOST", "ldap.testdomain.com")).To(Succeed())
				Expect(os.Setenv("LDAP_PORT", "389")).To(Succeed())
				Expect(os.Setenv("LDAP_PROTO", "tcp")).To(Succeed())
				Expect(os.Setenv("LDAP_TIMEOUT", "60")).To(Succeed())
				command.Args = append(command.Args, "-listenAddr=0.0.0.0:7593")
				command.Args = append(command.Args, "-adminAddr=0.0.0.0:7594")
			})

			AfterEach(func() {
				Expect(os.Unsetenv("LDAP_SVC_USER")).To(Succeed())
				Expect(os.Unsetenv("LDAP_SVC_PASS")).To(Succeed())
				Expect(os.Unsetenv("LDAP_USER_FQDN")).To(Succeed())
				Expect(os.Unsetenv("LDAP_HOST")).To(Succeed())
				Expect(os.Unsetenv("LDAP_PORT")).To(Succeed())
				Expect(os.Unsetenv("LDAP_PROTO")).To(Succeed())
				Expect(os.Unsetenv("LDAP_TIMEOUT")).To(Succeed())
			})

			It("listens on tcp/7589 by default", func() {
				EventuallyWithOffset(1, func() error {
					_, err := net.Dial("tcp", "0.0.0.0:7593")
					return err
				}, 5).ShouldNot(HaveOccurred())
			})
		})
	})
})
