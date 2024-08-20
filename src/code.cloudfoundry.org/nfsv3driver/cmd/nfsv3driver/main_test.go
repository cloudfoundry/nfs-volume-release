package main_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

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
		listenAddr, adminAddr  string
	)

	BeforeEach(func() {
		listenAddr = fmt.Sprintf("0.0.0.0:%d", listenPort)
		adminAddr = fmt.Sprintf("0.0.0.0:%d", adminPort)
		command = exec.Command(driverPath)
		expectedStartOutput = "started"
		expectedStartErrOutput = ""
	})

	JustBeforeEach(func() {
		var err error
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		time.Sleep(time.Second)
		Eventually(session.Out).Should(gbytes.Say(expectedStartOutput))
		Eventually(session.Err).Should(gbytes.Say(expectedStartErrOutput))
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		session.Kill()
		Eventually(session.Exited).Should(BeClosed())
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

		Context("when command line arguments are provided", func() {
			BeforeEach(func() {
				command.Args = append(command.Args, fmt.Sprintf("-listenAddr=%s", listenAddr))
				command.Args = append(command.Args, fmt.Sprintf("-adminAddr=%s", adminAddr))
			})

			It("listens on provided arguments", func() {
				EventuallyWithOffset(1, func() error {
					_, err := net.Dial("tcp", listenAddr)
					return err
				}, 5).ShouldNot(HaveOccurred())
			})

			It("listens on provided arguments", func() {
				EventuallyWithOffset(1, func() error {
					_, err := net.Dial("tcp", adminAddr)
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
						_, err := net.Dial("tcp", listenAddr)
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

				command.Args = append(command.Args, fmt.Sprintf("-listenAddr=%s", listenAddr))
				command.Args = append(command.Args, fmt.Sprintf("-adminAddr=%s", adminAddr))
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
					_, err := net.Dial("tcp", listenAddr)
					return err
				}, 5).ShouldNot(HaveOccurred())
			})
		})

		Context("given incomplete LDAP arguments set in the environment", func() {
			BeforeEach(func() {
				Expect(os.Setenv("LDAP_HOST", "ldap.testdomain.com")).To(Succeed())
				Expect(os.Setenv("LDAP_PORT", "389")).To(Succeed())
				Expect(os.Setenv("LDAP_PROTO", "tcp")).To(Succeed())
				command.Args = append(command.Args, fmt.Sprintf("-listenAddr=%s", listenAddr))
				command.Args = append(command.Args, fmt.Sprintf("-adminAddr=%s", adminAddr))
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
					_, err := net.Dial("tcp", listenAddr)
					return err
				}, 5).Should(HaveOccurred())
			})
		})
	})
})
