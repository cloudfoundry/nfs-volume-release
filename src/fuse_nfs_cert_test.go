package fuse_nfs_certs

import (
	"os/exec"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	"path/filepath"
	"syscall"

	"fmt"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	PCAP uint32 = 200
)

var _ = Describe("Certify with: ", func() {
	var (
		testLogger lager.Logger
		err        error

		output []byte
		attrs  syscall.SysProcAttr

		pcapMountPath string
		rootMountPath string
	)

	BeforeEach(func() {
		testLogger = lagertest.NewTestLogger("MainTest")

		attrs = syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: PCAP,
				Gid: PCAP,
			},
		}
	})

	Context("given a pcap user with uid:gid 200:200", func() {
		BeforeEach(func() {
			output, err = asRoot(testLogger, "groupadd", "-g", fmt.Sprintf("%d", PCAP), "pcap")
			Expect(err).NotTo(HaveOccurred())

			output, err = asRoot(testLogger, "useradd", "-u", fmt.Sprintf("%d", PCAP), "-g", fmt.Sprintf("%d", PCAP), "pcap")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			output, err = asRoot(testLogger, "userdel", "pcap")

			output, err = asRoot(testLogger, "groupdel", "pcap")
		})

		Context("given a fuse-nfs mount mapping pcap user to uid:gid 3000:3050", func() {
			BeforeEach(func() {
				// pcap mount
				pcapMountPath = filepath.Join("/tmp", "fuse_nfs_certs")
				output, err = asUser(testLogger, PCAP, PCAP, "mkdir", "-p", pcapMountPath)
				Expect(err).NotTo(HaveOccurred())

				output, err = asUser(testLogger, PCAP, PCAP, "fuse-nfs", "-n", "nfs://nfs.persi.cf-app.com/export/users?uid=3000&gid=3050", "-m", pcapMountPath)
				Expect(err).NotTo(HaveOccurred())

				// root mount
				rootMountPath = filepath.Join("/tmp", "fuse_nfs_certs_root")

				output, err = asRoot(testLogger, "mkdir", "-p", rootMountPath)
				Expect(err).NotTo(HaveOccurred())

				output, err = asRoot(testLogger, "mount", "-t", "nfs", "-o", "nfsvers=3,nolock", "nfs.persi.cf-app.com:/export/users", rootMountPath)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				output, err = asRoot(testLogger, "umount", rootMountPath)
				Expect(err).NotTo(HaveOccurred())

				err = os.Remove(rootMountPath)
				Expect(err).ToNot(HaveOccurred())

				output, err = asRoot(testLogger, "umount", "-f", pcapMountPath)
				Expect(err).NotTo(HaveOccurred())

				err = os.Remove(pcapMountPath)
				Expect(err).ToNot(HaveOccurred())
			})

			It("successfully creates a file with uid:gid pcap:pcap", func() {
				output, err = asUser(testLogger, PCAP, PCAP, "touch", filepath.Join(pcapMountPath, "tempfile"))
				Expect(err).NotTo(HaveOccurred())

				output, err = asUser(testLogger, PCAP, PCAP, "stat", "-c", "%u:%g", filepath.Join(pcapMountPath, "tempfile"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(Equal("200:200\n"))

				output, err = asUser(testLogger, PCAP, PCAP, "stat", "-c", "%u:%g", filepath.Join(rootMountPath, "tempfile"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(output)).To(Equal("3000:3050\n"))
			})
		})
	})
})

func asUser(logger lager.Logger, uid, gid uint32, cmd string, args ...string) ([]byte, error) {
	logger.Info(fmt.Sprintf("Executing command %s %#v", cmd, args))
	cmdHandle := exec.Command(cmd, args...)

	attrs := syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uid,
			Gid: gid,
		},
	}
	cmdHandle.SysProcAttr = &attrs

	output, err := cmdHandle.CombinedOutput()
	if err != nil {
		logger.Error(string(output), err)
	}

	return output, err
}

func asRoot(logger lager.Logger, cmd string, args ...string) ([]byte, error) {
	return asUser(logger, 0, 0, cmd, args...)
}
