package nfsv3driver_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/goshims/osshim/os_fake"
	"code.cloudfoundry.org/goshims/syscallshim/syscall_fake"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/nfsv3driver"
	"code.cloudfoundry.org/nfsv3driver/nfsdriverfakes"
	vmo "code.cloudfoundry.org/volume-mount-options"
	"code.cloudfoundry.org/volumedriver"
	"code.cloudfoundry.org/volumedriver/invoker"
	"code.cloudfoundry.org/volumedriver/invokerfakes"
	nfsfakes "code.cloudfoundry.org/volumedriver/volumedriverfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("MapfsMounter", func() {

	var (
		logger      *lagertest.TestLogger
		testContext context.Context
		env         dockerdriver.Env
		err         error

		fakeInvoker    *invokerfakes.FakeInvoker
		fakeIdResolver *nfsdriverfakes.FakeIdResolver

		fakeInvokeResult *invokerfakes.FakeInvokeResult

		subject          volumedriver.Mounter
		fakeOs           *os_fake.FakeOs
		fakeMountChecker *nfsfakes.FakeMountChecker
		fakeSyscall      *syscall_fake.FakeSyscall

		opts      map[string]interface{}
		mapfsPath string
		mask      vmo.MountOptsMask
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("mapfs-mounter")
		mapfsPath = "/var/vcap/packages/mapfs/bin/mapfs"
		testContext = context.TODO()
		env = driverhttp.NewHttpDriverEnv(logger, testContext)
		opts = map[string]interface{}{}
		opts["uid"] = "2000"
		opts["gid"] = "2000"

		fakeInvoker = &invokerfakes.FakeInvoker{}
		fakeInvokeResult = &invokerfakes.FakeInvokeResult{}
		fakeInvoker.InvokeReturns(fakeInvokeResult)
		fakeInvokeResult.WaitReturns(nil)

		fakeOs = &os_fake.FakeOs{}
		fakeSyscall = &syscall_fake.FakeSyscall{}
		fakeOs.OpenFileReturns(&os_fake.FakeFile{}, nil)
		fakeMountChecker = &nfsfakes.FakeMountChecker{}
		fakeMountChecker.ExistsReturns(true, nil)

		fakeOs.StatReturns(nil, nil)
		fakeOs.IsExistReturns(true)

		fakeSyscall.StatStub = func(path string, st *syscall.Stat_t) error {
			st.Mode = 0777
			st.Uid = 1000
			st.Gid = 1000
			return nil
		}

		mask, err = nfsv3driver.NewMapFsVolumeMountMask()
		Expect(err).NotTo(HaveOccurred())

		subject = nfsv3driver.NewMapfsMounter(fakeInvoker, fakeOs, fakeSyscall, fakeMountChecker, "my-fs", "my-mount-options,timeo=600,retrans=2,actimeo=0", nil, mask, mapfsPath)
	})

	Context("#Mount", func() {
		var (
			source, target string
		)
		BeforeEach(func() {
			source = "source"
			target = "target"
		})
		JustBeforeEach(func() {
			err = subject.Mount(env, source, target, opts)
		})

		Context("when version is specified", func() {
			BeforeEach(func() {
				opts["version"] = "4.1"
			})

			It("should use version specified", func() {
				Expect(err).NotTo(HaveOccurred())
				_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("mount"))
				Expect(len(args)).To(BeNumerically(">", 5))
				Expect(args).To(ContainElement("-t"))
				Expect(args).To(ContainElement("my-fs"))
				Expect(args).To(ContainElement("-o"))
				Expect(args).To(ContainElement("my-mount-options,timeo=600,retrans=2,actimeo=0,vers=4.1"))
				Expect(args).To(ContainElement("source"))
				Expect(args).To(ContainElement("target_mapfs"))
			})
			Context("when NFS version 3.X is specified", func() {
				When("version specified is numerically equal to a valid version 3", func() {
					BeforeEach(func() {
						opts["version"] = "3.0"
					})
					It("should automatically correct the version to 3 if the values are numerically equal", func() {
						Expect(err).NotTo(HaveOccurred())
						_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(0)
						Expect(cmd).To(Equal("mount"))
						Expect(len(args)).To(BeNumerically(">", 5))
						Expect(args).To(ContainElement("-t"))
						Expect(args).To(ContainElement("my-fs"))
						Expect(args).To(ContainElement("-o"))
						Expect(args).To(ContainElement("my-mount-options,timeo=600,retrans=2,actimeo=0,vers=3"))
						Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("NFSv3 does not have a minor version available")))
					})
				})
			})
			DescribeTable("when version is invalid", func(expectErr string, version string) {
				opts["version"] = version

				err = subject.Mount(env, source, target, opts)
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
				Expect(err).To(MatchError(expectErr))
			},
				Entry("version with additional options", "\"version\" must be a positive numeric value", "4.1,4.2"),
				Entry("not a number", "\"version\" must be a positive numeric value", "foo"),
				Entry("negative number", "\"version\" must be a positive numeric value", "-1"),
				Entry("not a valid version", "\"version\" must be a positive numeric value", "0"),
				Entry("specifying nfsv3 with minor version of 3.1", "NFSv3 does not use minor versions. NFSv 3.1 does not exist", "3.1"),
				Entry("specifying nfsv3 with minor version of 3.9", "NFSv3 does not use minor versions. NFSv 3.9 does not exist", "3.9"),
			)
		})
		Context("when experimental is specified", func() {
			BeforeEach(func() {
				opts["experimental"] = "true"
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when source is specified", func() {
			BeforeEach(func() {
				opts["source"] = "some-source"
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when mount is specified", func() {
			BeforeEach(func() {
				opts["mount"] = "some-mount"
			})

			It("should succeed", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when mount succeeds", func() {
			It("should use the mapfs mounter", func() {
				Expect(fakeInvoker.InvokeCallCount()).NotTo(BeZero())
			})

			It("should return without error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("should create an intermediary mount point", func() {
				Expect(fakeOs.MkdirAllCallCount()).NotTo(BeZero())
				dirName, mode := fakeOs.MkdirAllArgsForCall(0)
				Expect(dirName).To(Equal("target_mapfs"))
				Expect(mode).To(Equal(os.ModePerm))
			})

			It("should use the passed in variables", func() {
				_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("mount"))
				Expect(len(args)).To(BeNumerically(">", 5))
				Expect(args).To(ContainElement("-t"))
				Expect(args).To(ContainElement("my-fs"))
				Expect(args).To(ContainElement("-o"))
				Expect(args).To(ContainElement("my-mount-options,timeo=600,retrans=2,actimeo=0"))
				Expect(args).To(ContainElement("source"))
				Expect(args).To(ContainElement("target_mapfs"))
			})

			It("should launch mapfs to mount the target", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
				_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(1)
				expectedText, duration := fakeInvokeResult.WaitForArgsForCall(0)

				Expect(cmd).To(Equal(mapfsPath))
				Expect(args).To(ContainElement("-uid"))
				Expect(args).To(ContainElement("2000"))
				Expect(args).To(ContainElement("-gid"))
				Expect(args).To(ContainElement("target"))
				Expect(args).To(ContainElement("target_mapfs"))

				Expect(expectedText).To(Equal("Mounted!"))
				Expect(duration).To(Equal(time.Minute * 5))
			})

			Context("when mkdir fails", func() {
				BeforeEach(func() {
					fakeOs.MkdirAllReturns(errors.New("failed-to-create-dir"))
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(dockerdriver.SafeError)
					Expect(ok).To(BeTrue())
				})
			})

			Context("when the mount is readonly", func() {
				BeforeEach(func() {
					opts["readonly"] = true
				})

				It("should not append 'ro' to the kernel mount options, since garden manages the ro mount", func() {
					_, _, args, _ := fakeInvoker.InvokeArgsForCall(0)
					Expect(len(args)).To(BeNumerically(">", 3))
					Expect(args[2]).To(Equal("-o"))
					Expect(args[3]).NotTo(ContainSubstring(",ro"))
				})
				It("should not append 'actimeo=0' to the kernel mount options", func() {
					Expect(err).NotTo(HaveOccurred())
					_, _, args, _ := fakeInvoker.InvokeArgsForCall(0)
					Expect(len(args)).To(BeNumerically(">", 3))
					Expect(args[2]).To(Equal("-o"))
					Expect(args[3]).NotTo(ContainSubstring("actimeo=0"))
				})
			})

			Context("cache option", func() {
				Context("when the mount is cache true", func() {
					BeforeEach(func() {
						opts["cache"] = true
					})

					It("should not append 'actimeo=0' to the kernel mount options", func() {
						Expect(err).NotTo(HaveOccurred())
						_, _, args, _ := fakeInvoker.InvokeArgsForCall(0)
						Expect(len(args)).To(BeNumerically(">", 3))
						Expect(args[2]).To(Equal("-o"))
						Expect(args[3]).NotTo(ContainSubstring("actimeo=0"))
						Expect(args[3]).NotTo(ContainSubstring("cache"))
					})
				})

				Context("when the mount is cache false", func() {
					BeforeEach(func() {
						opts["cache"] = false
					})

					It("should append 'actimeo=0' to the kernel mount options", func() {
						Expect(err).NotTo(HaveOccurred())
						_, _, args, _ := fakeInvoker.InvokeArgsForCall(0)
						Expect(len(args)).To(BeNumerically(">", 3))
						Expect(args[2]).To(Equal("-o"))
						Expect(args[3]).To(ContainSubstring("actimeo=0"))
						Expect(args[3]).NotTo(ContainSubstring("cache"))
					})
				})

				Context("when the mount is cache unset", func() {
					BeforeEach(func() {
						Expect(opts).NotTo(HaveKey("cache"))
					})

					It("should append 'actimeo=0' to the kernel mount options", func() {
						Expect(err).NotTo(HaveOccurred())
						_, _, args, _ := fakeInvoker.InvokeArgsForCall(0)
						Expect(len(args)).To(BeNumerically(">", 3))
						Expect(args[2]).To(Equal("-o"))
						Expect(args[3]).To(ContainSubstring("actimeo=0"))
						Expect(args[3]).NotTo(ContainSubstring("cache"))
					})
				})

				Context("when the mount is cache false and readonly", func() {
					BeforeEach(func() {
						opts["cache"] = false
						opts["readonly"] = true
					})

					It("should append 'actimeo=0' to the kernel mount options", func() {
						Expect(err).NotTo(HaveOccurred())
						_, _, args, _ := fakeInvoker.InvokeArgsForCall(0)
						Expect(len(args)).To(BeNumerically(">", 3))
						Expect(args[2]).To(Equal("-o"))
						Expect(args[3]).To(ContainSubstring("actimeo=0"))
						Expect(args[3]).NotTo(ContainSubstring("cache"))
					})
				})

				Context("when the mount is cache invalid", func() {
					BeforeEach(func() {
						opts["cache"] = "foobar"
					})

					It("should return an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(MatchError("Invalid 'cache' option"))

						_, ok := err.(dockerdriver.SafeError)
						Expect(ok).To(BeTrue())
					})
				})
			})

			DescribeTable("when the mount has a legacy format", func(legacySourceFormat string, expectedShareFormat string) {
				fakeInvoker = &invokerfakes.FakeInvoker{}
				fakeInvoker.InvokeReturns(fakeInvokeResult)
				subject = nfsv3driver.NewMapfsMounter(fakeInvoker, fakeOs, fakeSyscall, fakeMountChecker, "my-fs", "my-mount-options,timeo=600,retrans=2,actimeo=0", nil, mask, mapfsPath)

				err = subject.Mount(env, legacySourceFormat, target, opts)
				Expect(err).NotTo(HaveOccurred())

				_, _, args, _ := fakeInvoker.InvokeArgsForCall(0)
				Expect(len(args)).To(BeNumerically(">", 4))
				Expect(args[4]).To(Equal(expectedShareFormat))
			},
				Entry("with subdirectories", "nfs://server/some/share/path/", "server:/some/share/path/"),
				Entry("with subdirectories without a trailing slash", "nfs://server/some/share/path", "server:/some/share/path"),
				Entry("without subdirectories", "nfs://server/", "server:/"),
				Entry("without subdirectories without a trailing slash", "nfs://server", "server:/"),
			)

			Context("when the share value is invalid", func() {
				BeforeEach(func() {
					source = "nfs:// "
				})
				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("Invalid 'share' option"))
				})
			})

			Context("when the target has a trailing slash", func() {
				BeforeEach(func() {
					target = "/some/target/"
				})
				It("should rewrite the target to remove the slash", func() {
					Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
					_, _, args, _ := fakeInvoker.InvokeArgsForCall(1)
					Expect(args[5]).To(Equal("/some/target"))
					Expect(args[6]).To(Equal("/some/target_mapfs"))
				})
			})

			Context("when other options are specified", func() {
				BeforeEach(func() {
					opts["auto_cache"] = true
				})
				It("should include those options on the mapfs invoke call", func() {
					Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
					_, _, args, _ := fakeInvoker.InvokeArgsForCall(1)
					Expect(args).To(ContainElement("-auto_cache"))
				})
			})
		})

		Context("when there is no uid", func() {
			BeforeEach(func() {
				delete(opts, "uid")

				fakeInvoker.InvokeStub = func(d dockerdriver.Env, actualCommand string, s []string, e ...string) invoker.InvokeResult {
					Expect(actualCommand).NotTo(ContainSubstring(mapfsPath), "should not launch mapfs")
					return fakeInvokeResult
				}

			})

			It("should create an intermediary mount point", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeOs.MkdirAllCallCount()).NotTo(BeZero())
				dirName, _ := fakeOs.MkdirAllArgsForCall(0)
				Expect(dirName).To(Equal("target_mapfs"))
			})

			It("should mount directly to the target", func() {
				Expect(err).NotTo(HaveOccurred())
				_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("mount"))
				Expect(args).To(ContainElement("source"))
				Expect(args).To(ContainElement("target"))
			})

			It("should not launch mapfs", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(Equal(1))
			})
		})
		Context("when there is no gid", func() {
			BeforeEach(func() {
				delete(opts, "gid")
			})
			It("should error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
			})
		})
		Context("when uid is an integer", func() {
			BeforeEach(func() {
				opts["uid"] = 2000
			})
			It("should not error", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
				_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(1)
				Expect(cmd).To(Equal(mapfsPath))
				Expect(args).To(ContainElement("-uid"))
				Expect(args).To(ContainElement("2000"))
				Expect(args).To(ContainElement("-gid"))
				Expect(args).To(ContainElement("target"))
				Expect(args).To(ContainElement("target_mapfs"))
			})
		})
		Context("when gid is an integer", func() {
			BeforeEach(func() {
				opts["gid"] = 2000
			})
			It("should not error", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(BeNumerically(">=", 1))
				_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(1)
				Expect(cmd).To(Equal(mapfsPath))
				Expect(args).To(ContainElement("-uid"))
				Expect(args).To(ContainElement("2000"))
				Expect(args).To(ContainElement("-gid"))
				Expect(args).To(ContainElement("target"))
				Expect(args).To(ContainElement("target_mapfs"))
			})
		})

		DescribeTable("when uid is provided invalid values it should error", func(invalidUid interface{}) {
			opts["uid"] = invalidUid
			err = subject.Mount(env, source, target, opts)
			Expect(err).To(HaveOccurred())
			_, ok := err.(dockerdriver.SafeError)
			Expect(ok).To(BeTrue())
			Expect(err.Error()).To(Equal("Invalid 'uid' option (0, negative, or non-integer)"))

		},
			Entry("when uid is not an integer", "foo"),
			Entry("when uid is negative", -1),
			Entry("when uid is not an integer", 1.2),
			Entry("when uid is zero", 0),
		)

		DescribeTable("when gid is provided invalid values it should error", func(invalidGid interface{}) {
			opts["gid"] = invalidGid
			err = subject.Mount(env, source, target, opts)
			Expect(err).To(HaveOccurred())
			_, ok := err.(dockerdriver.SafeError)
			Expect(ok).To(BeTrue())
			Expect(err.Error()).To(Equal("Invalid 'gid' option (0, negative, or non-integer)"))

		},
			Entry("when gid is not an integer", "foo"),
			Entry("when gid is negative", -1),
			Entry("when gid is not an integer", 1.2),
			Entry("when gid is zero", 0),
		)

		Context("when the specified uid doesn't have read access", func() {
			BeforeEach(func() {
				fakeSyscall.StatStub = func(path string, st *syscall.Stat_t) error {
					st.Mode = 0750
					st.Uid = 1000
					st.Gid = 1000
					return nil
				}
			})
			It("should fail and clean up the intermediate mount", func() {
				Expect(fakeSyscall.StatCallCount()).NotTo(BeZero())
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
				Expect(err.Error()).To(ContainSubstring("access"))

				Expect(fakeInvoker.InvokeCallCount()).To(Equal(2))
				_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(1)
				Expect(cmd).To(Equal("umount"))
				Expect(len(args)).To(BeNumerically(">", 0))
				Expect(args[0]).To(Equal("target_mapfs"))
				Expect(fakeOs.RemoveCallCount()).To(Equal(1))

				Expect(logger.LogMessages()).NotTo(ContainElement(ContainSubstring("intermediate-unmount-failed")))
				Expect(logger.LogMessages()).NotTo(ContainElement(ContainSubstring("intermediate-remove-failed")))
			})

			Context("when it fails to unmount the intermediate directory", func() {
				BeforeEach(func() {
					result := &invokerfakes.FakeInvokeResult{}
					result.WaitReturns(errors.New("intermediate-unmount-failed"))
					fakeInvoker.InvokeReturnsOnCall(1, result)
				})

				It("should log the error", func() {
					Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("intermediate-unmount-failed")))
					Expect(fakeOs.RemoveCallCount()).To(Equal(0))
				})
			})

			Context("when it fails to remove the intermediate directory", func() {

				BeforeEach(func() {
					fakeOs.RemoveReturns(errors.New("intermediate-remove-failed"))
				})
				It("should log the error", func() {
					Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("intermediate-remove-failed")))
				})
			})
		})
		Context("when stat() fails during access check", func() {
			BeforeEach(func() {
				fakeSyscall.StatStub = func(path string, st *syscall.Stat_t) error {
					return errors.New("this is nacho share.")
				}
			})
			It("should succeed and log a warning", func() {
				Expect(fakeSyscall.StatCallCount()).NotTo(BeZero())
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.TestSink.Buffer().Contents()).To(ContainSubstring("nacho share"))
			})

		})
		Context("when stat returns ambiguous results", func() {
			var (
				uid = uint32(1000)
				gid = uint32(1000)
			)
			BeforeEach(func() {
				fakeSyscall.StatStub = func(path string, st *syscall.Stat_t) error {
					st.Mode = 0750
					st.Uid = uid
					st.Gid = gid
					return nil
				}
			})

			Context("when uid is unknown", func() {
				BeforeEach(func() {
					uid = nfsv3driver.UnknownId
				})
				It("should succeed", func() {
					Expect(fakeSyscall.StatCallCount()).NotTo(BeZero())
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when gid is unknown", func() {
				BeforeEach(func() {
					gid = nfsv3driver.UnknownId
				})
				It("should succeed", func() {
					Expect(fakeSyscall.StatCallCount()).NotTo(BeZero())
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
		Context("when idresolver isn't present but username is passed", func() {
			BeforeEach(func() {
				delete(opts, "uid")
				delete(opts, "gid")
				opts["username"] = "test-user"
				opts["password"] = "test-pw"
			})

			It("should error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
				Expect(err.Error()).To(ContainSubstring("LDAP is not configured"))
			})
		})

		Context("when mount errors", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns(fakeInvokeResult)
				fakeInvokeResult.WaitReturns(fmt.Errorf("error"))
			})

			It("should return error", func() {
				Expect(fakeInvokeResult.WaitCallCount()).To(Equal(1))
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
			})

			It("should remove the intermediary mountpoint", func() {
				Expect(logger.LogMessages()).NotTo(ContainElement(ContainSubstring("remove-failed")))

				Expect(fakeOs.RemoveCallCount()).To(Equal(1))
			})

			Context("when the intermediate mount directory remove fails", func() {
				BeforeEach(func() {
					fakeOs.RemoveReturns(errors.New("remove-failed"))
				})

				It("should log an error", func() {
					Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("remove-failed")))
				})
			})

		})

		Context("when kernel mount succeeds, but mapfs mount fails", func() {
			Context("mapfs mount cmd fails", func() {
				BeforeEach(func() {
					fakeInvokeResult.WaitForReturns(errors.New("mount error"))
				})

				It("should invoke unmount", func() {
					Expect(fakeInvoker.InvokeCallCount()).To(Equal(3))
					_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(2)
					Expect(cmd).To(Equal("umount"))
					Expect(len(args)).To(BeNumerically(">", 0))
					Expect(args[0]).To(Equal("target_mapfs"))

					Expect(fakeInvokeResult.WaitCallCount()).To(Equal(2))
				})

				It("should remove the intermediary mountpoint", func() {
					Expect(fakeOs.RemoveCallCount()).To(Equal(1))
					Expect(logger.LogMessages()).NotTo(ContainElement(ContainSubstring("unmount-failed")))
					Expect(logger.LogMessages()).NotTo(ContainElement(ContainSubstring("remove-failed")))
				})

				It("should safely return the mount error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(dockerdriver.SafeError)
					Expect(ok).To(BeTrue())
					Expect(err).To(MatchError("mount error"))
				})

				Context("when unmount fails", func() {
					BeforeEach(func() {
						fakeInvokeResult.WaitReturnsOnCall(1, errors.New(""))
					})
					It("should log the error", func() {
						Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("unmount-failed")))
						Expect(fakeOs.RemoveCallCount()).To(Equal(0))
					})

					It("should return the mount error safely", func() {
						Expect(err).To(HaveOccurred())
						_, ok := err.(dockerdriver.SafeError)
						Expect(ok).To(BeTrue())
						Expect(err).To(MatchError("mount error"))
					})
				})

				Context("when remove fails", func() {
					BeforeEach(func() {
						fakeOs.RemoveReturns(errors.New("remove-failed"))
					})

					It("should log the error", func() {
						Expect(logger.LogMessages()).To(ContainElement(ContainSubstring("remove-failed")))
					})

					It("should return the mount error safely", func() {
						Expect(err).To(HaveOccurred())
						_, ok := err.(dockerdriver.SafeError)
						Expect(ok).To(BeTrue())
						Expect(err).To(MatchError("mount error"))
					})
				})
			})
		})

		Context("when provided a username to map to a uid", func() {
			BeforeEach(func() {
				fakeIdResolver = &nfsdriverfakes.FakeIdResolver{}

				subject = nfsv3driver.NewMapfsMounter(fakeInvoker, fakeOs, fakeSyscall, fakeMountChecker, "my-fs", "my-mount-options", fakeIdResolver, mask, mapfsPath)
				fakeIdResolver.ResolveReturns("100", "100", nil)

				delete(opts, "uid")
				delete(opts, "gid")
				opts["username"] = "test-user"
				opts["password"] = "test-pw"
			})

			It("does not show the credentials in the options", func() {
				Expect(err).NotTo(HaveOccurred())
				_, _, args, _ := fakeInvoker.InvokeArgsForCall(0)
				Expect(strings.Join(args, " ")).To(Not(ContainSubstring("username")))
				Expect(strings.Join(args, " ")).To(Not(ContainSubstring("password")))
			})

			It("shows gid and uid", func() {
				Expect(err).NotTo(HaveOccurred())
				_, _, args, _ := fakeInvoker.InvokeArgsForCall(1)
				Expect(strings.Join(args, " ")).To(ContainSubstring("-uid 100"))
				Expect(strings.Join(args, " ")).To(ContainSubstring("-gid 100"))
			})

			Context("when username is passed but password is not passed", func() {
				BeforeEach(func() {
					delete(opts, "password")
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(dockerdriver.SafeError)
					Expect(ok).To(BeTrue())
					Expect(err.Error()).To(ContainSubstring("LDAP password is missing"))
				})
			})

			Context("when uid is NaN", func() {
				BeforeEach(func() {
					fakeIdResolver.ResolveReturns("uid-not-a-number", "1", nil)
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Invalid 'uid' option (0, negative, or non-integer)"))
				})
			})

			Context("when gid is NaN", func() {
				BeforeEach(func() {
					fakeIdResolver.ResolveReturns("1", "gid-not-a-number", nil)
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Invalid 'gid' option (0, negative, or non-integer)"))
				})
			})

			Context("when uid is passed", func() {
				BeforeEach(func() {
					opts["uid"] = "100"
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(dockerdriver.SafeError)
					Expect(ok).To(BeTrue())
					Expect(err.Error()).To(ContainSubstring("Not allowed options"))
				})
			})

			Context("when gid is passed", func() {
				BeforeEach(func() {
					opts["gid"] = "100"
				})

				It("should error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(dockerdriver.SafeError)
					Expect(ok).To(BeTrue())
					Expect(err.Error()).To(ContainSubstring("Not allowed options"))
				})
			})

			Context("when unable to resolve username", func() {
				BeforeEach(func() {
					fakeIdResolver.ResolveReturns("", "", errors.New("unable to resolve"))
				})

				It("return an error that is not a SafeError since it might contain sensitive information", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("unable to resolve"))
					Expect(err).NotTo(BeAssignableToTypeOf(dockerdriver.SafeError{}))
				})
			})
		})
	})

	Context("#Unmount", func() {
		var target string
		BeforeEach(func() {
			target = "target"
		})

		JustBeforeEach(func() {
			err = subject.Unmount(env, target)
		})

		Context("when unmount succeeds", func() {
			It("should return without error", func() {
				Expect(logger.Buffer()).NotTo(gbytes.Say("failed"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should invoke unmount on both the mapfs and target mountpoints", func() {
				Expect(fakeInvoker.InvokeCallCount()).To(Equal(2))

				_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(0)
				Expect(cmd).To(Equal("umount"))
				Expect(len(args)).To(Equal(2))
				Expect(args[0]).To(Equal("-l"))
				Expect(args[1]).To(Equal("target"))

				_, cmd, args, _ = fakeInvoker.InvokeArgsForCall(1)
				Expect(cmd).To(Equal("umount"))
				Expect(len(args)).To(Equal(2))
				Expect(args[0]).To(Equal("-l"))
				Expect(args[1]).To(Equal("target_mapfs"))
			})

			It("should delete the mapfs mount point", func() {
				Expect(fakeOs.RemoveCallCount()).ToNot(BeZero())
				Expect(fakeOs.RemoveArgsForCall(0)).To(Equal("target_mapfs"))
			})

			Context("when the target has a trailing slash", func() {
				BeforeEach(func() {
					target = "/some/target/"
				})

				It("should rewrite the target to remove the slash", func() {
					Expect(fakeInvoker.InvokeCallCount()).To(Equal(2))

					_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(0)
					Expect(cmd).To(Equal("umount"))
					Expect(len(args)).To(Equal(2))
					Expect(args[0]).To(Equal("-l"))
					Expect(args[1]).To(Equal("/some/target"))

					_, cmd, args, _ = fakeInvoker.InvokeArgsForCall(1)
					Expect(cmd).To(Equal("umount"))
					Expect(len(args)).To(Equal(2))
					Expect(args[0]).To(Equal("-l"))
					Expect(args[1]).To(Equal("/some/target_mapfs"))
				})
			})

			Context("when uid mapping was not used for the mount", func() {
				BeforeEach(func() {
					fakeMountChecker.ExistsStub = func(s string) (bool, error) {
						Expect(s).To(Equal("target_mapfs"))
						return false, nil
					}
				})

				It("should not attempt to unmount the intermediate mount", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeInvoker.InvokeCallCount()).To(Equal(1))
				})

				Context("when the mount checker returns an error", func() {
					BeforeEach(func() {
						fakeMountChecker.ExistsReturns(false, errors.New("mount-checker-failed"))
					})

					It("should log the error", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(logger.Buffer()).To(gbytes.Say("mount-checker-failed"))
					})
				})
			})

			Context("when the intermediate directory does not exist", func() {
				BeforeEach(func() {
					fakeOs.StatStub = func(name string) (os.FileInfo, error) {
						Expect(name).To(Equal("target_mapfs"))
						return nil, &os.PathError{Err: os.ErrNotExist}
					}
				})

				It("should succeeed", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeOs.RemoveCallCount()).To(Equal(0))
				})
			})
		})

		Context("umount cmd errors", func() {
			BeforeEach(func() {
				fakeInvokeResult.WaitReturns(fmt.Errorf("umount error"))
			})

			It("should return an error", func() {
				Expect(fakeInvokeResult.WaitCallCount()).To(Equal(1))

				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
				Expect(err).To(MatchError("umount error"))
			})
		})

		Context("when waiting for unmount of the intermediate mount fails", func() {
			BeforeEach(func() {
				fakeInvokeResult.WaitReturnsOnCall(0, nil)
				fakeInvokeResult.WaitReturnsOnCall(1, fmt.Errorf("mapfs umount error"))
			})

			It("should not return an error", func() {
				Expect(fakeInvokeResult.WaitCallCount()).To(Equal(2))
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.Buffer()).To(gbytes.Say("mapfs umount error"))
			})

			It("should not call Remove on the intermediate directory", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeOs.RemoveCallCount()).To(Equal(0))
			})
		})

		Context("when remove fails", func() {
			BeforeEach(func() {
				fakeOs.RemoveReturns(errors.New("failed-to-remove-dir"))
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				_, ok := err.(dockerdriver.SafeError)
				Expect(ok).To(BeTrue())
			})
		})
	})

	Context("#Check", func() {

		var (
			success bool
		)

		JustBeforeEach(func() {
			success = subject.Check(env, "target", "source")
		})

		Context("when check succeeds", func() {

			It("uses correct context", func() {
				env, _, _, _ := fakeInvoker.InvokeArgsForCall(0)
				Expect(fmt.Sprintf("%#v", env.Context())).To(ContainSubstring("timerCtx"))
			})

			It("reports valid mountpoint", func() {
				Expect(fakeInvokeResult.WaitCallCount()).To(Equal(1))
				Expect(success).To(BeTrue())
			})
		})

		Context("when check command error", func() {
			BeforeEach(func() {
				fakeInvokeResult.WaitReturns(fmt.Errorf("check command error"))
			})
			It("reports invalid mountpoint", func() {
				Expect(success).To(BeFalse())
			})
		})
	})

	Context("#Purge", func() {
		var pathToPurge string

		BeforeEach(func() {
			pathToPurge = "/foo/foo/foo"
			fakeMountChecker.ListReturns([]string{"/foo/foo/foo/mount_one_mapfs"}, nil)
			fakeInvokeResult.StdOutputReturnsOnCall(0, "stdout from pkill fakeinvoke result")
			fakeInvokeResult.StdOutputReturnsOnCall(1, "stdout from pgrep fakeinvoke result")
		})

		JustBeforeEach(func() {
			subject.Purge(env, pathToPurge)
		})

		It("kills the mapfs mount processes", func() {
			Expect(fakeInvokeResult.WaitCallCount()).To(Equal(33))
			Expect(fakeInvoker.InvokeCallCount()).To(Equal(33))

			_, proc, args, _ := fakeInvoker.InvokeArgsForCall(0)
			Expect(proc).To(Equal("pkill"))
			Expect(args[0]).To(Equal("mapfs"))
			Expect(logger.Buffer()).Should(gbytes.Say("pkill.*stdout from pkill fakeinvoke result"))

			_, proc, args, _ = fakeInvoker.InvokeArgsForCall(1)
			Expect(proc).To(Equal("pgrep"))
			Expect(args[0]).To(Equal("mapfs"))

			Expect(logger.Buffer()).Should(gbytes.Say("stdout from pgrep fakeinvoke result"))
		})

		Context("pkill process is no longer found by pgrep", func() {
			BeforeEach(func() {
				fakeInvokeResult.WaitReturnsOnCall(11, fmt.Errorf("pkill not found"))
			})

			It("continues", func() {
				Expect(fakeInvokeResult.WaitCallCount()).To(Equal(14))
			})
		})

		It("should unmount both the mounts", func() {
			Expect(fakeInvoker.InvokeCallCount()).To(Equal(33))
			Expect(fakeInvokeResult.WaitCallCount()).To(Equal(33))

			_, cmd, args, _ := fakeInvoker.InvokeArgsForCall(fakeInvoker.InvokeCallCount() - 2)
			Expect(cmd).To(Equal("umount"))
			Expect(len(args)).To(Equal(3))
			Expect(args[0]).To(Equal("-l"))
			Expect(args[1]).To(Equal("-f"))
			Expect(args[2]).To(Equal("/foo/foo/foo/mount_one"))

			_, cmd, args, _ = fakeInvoker.InvokeArgsForCall(fakeInvoker.InvokeCallCount() - 1)
			Expect(cmd).To(Equal("umount"))
			Expect(len(args)).To(Equal(3))
			Expect(args[0]).To(Equal("-l"))
			Expect(args[1]).To(Equal("-f"))
			Expect(args[2]).To(Equal("/foo/foo/foo/mount_one_mapfs"))
		})

		It("should remove both the mountpoints", func() {
			Expect(fakeOs.RemoveCallCount()).To(Equal(2))

			path := fakeOs.RemoveArgsForCall(0)
			Expect(path).To(Equal("/foo/foo/foo/mount_one"))

			path = fakeOs.RemoveArgsForCall(1)
			Expect(path).To(Equal("/foo/foo/foo/mount_one_mapfs"))
		})

		Context("pkill command fails", func() {
			BeforeEach(func() {
				fakeInvokeResult.WaitReturnsOnCall(0, fmt.Errorf("pkill command errored out"))
			})

			It("returns", func() {
				Expect(logger.Buffer()).Should(gbytes.Say("pkill.*err.*pkill command errored out.*output.*stdout from pkill fakeinvoke result"))
			})
		})

		Context("pgrep command fails", func() {
			BeforeEach(func() {
				fakeInvokeResult.WaitReturnsOnCall(1, fmt.Errorf("pgrep command errored out"))
			})

			It("continues checking if the mapfs has been killed", func() {
				Expect(logger.Buffer()).Should(gbytes.Say("mapfs-mounter.purge.pgrep.*pgrep command errored out"))
			})
		})

		Context("umount on mapfs command fails", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns(fakeInvokeResult)
				fakeInvokeResult.WaitReturnsOnCall(31, fmt.Errorf("umount command error"))
			})

			It("returns", func() {
				Expect(logger.Buffer()).Should(gbytes.Say("warning-umount-command-intermediate-failed.*umount command error"))
			})
		})

		Context("umount on linux dir command fails", func() {
			BeforeEach(func() {
				fakeInvoker.InvokeReturns(fakeInvokeResult)
				fakeInvokeResult.WaitReturnsOnCall(32, fmt.Errorf("umount command error"))
			})

			It("returns", func() {
				Expect(logger.Buffer()).Should(gbytes.Say("warning-umount-mapfs-failed.*umount command error"))
			})
		})

		Context("when given a path to purge that is a malformed URI", func() {
			BeforeEach(func() {
				pathToPurge = "foo("
			})

			It("should log an error", func() {
				Expect(logger.TestSink.Buffer()).Should(gbytes.Say("unable-to-list-mounts"))
				Expect(fakeMountChecker.ListCallCount()).To(Equal(0))
			})
		})

		Context("when list mounts invoke fails", func() {
			BeforeEach(func() {
				fakeMountChecker.ListReturns(nil, errors.New("list-failed"))
			})

			It("should log the error and not attempt any unmounts", func() {
				Expect(logger.Buffer()).To(gbytes.Say("list-failed"))
				Expect(logger.Buffer()).NotTo(gbytes.Say("mount-directory-list"))

			})
		})

	})
})
