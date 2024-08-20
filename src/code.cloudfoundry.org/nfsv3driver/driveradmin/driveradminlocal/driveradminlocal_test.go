package driveradminlocal_test

import (
	"context"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/nfsv3driver/driveradmin"
	"code.cloudfoundry.org/nfsv3driver/driveradmin/driveradminlocal"
	"code.cloudfoundry.org/nfsv3driver/nfsdriverfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Driver Admin Local", func() {
	var logger lager.Logger
	var ctx context.Context
	var env dockerdriver.Env
	var driverAdminLocal *driveradminlocal.DriverAdminLocal
	var err driveradmin.ErrorResponse

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("driveradminlocal")
		ctx = context.TODO()
		env = driverhttp.NewHttpDriverEnv(logger, ctx)
	})

	Context("created", func() {
		BeforeEach(func() {
			driverAdminLocal = driveradminlocal.NewDriverAdminLocal()
		})

		Describe("Evacuate", func() {
			JustBeforeEach(func() {
				err = driverAdminLocal.Evacuate(env)
			})
			Context("when the driver evacuates with no process set", func() {
				It("should fail", func() {
					Expect(err.Err).To(ContainSubstring("server process not found"))
				})
			})
			Context("when the driver evacuates with a process set", func() {
				var fakeProcess *nfsdriverfakes.FakeProcess

				BeforeEach(func() {
					fakeProcess = &nfsdriverfakes.FakeProcess{}
					driverAdminLocal.SetServerProc(fakeProcess)
				})

				It("should signal the process to terminate", func() {
					Expect(err.Err).To(BeEmpty())
					Expect(fakeProcess.SignalCallCount()).NotTo(Equal(0))
				})
				Context("when there is a drainable server registered", func() {
					var fakeDrainable *nfsdriverfakes.FakeDrainable
					BeforeEach(func() {
						fakeDrainable = &nfsdriverfakes.FakeDrainable{}
						driverAdminLocal.RegisterDrainable(fakeDrainable)
					})
					It("should drain", func() {
						Expect(fakeDrainable.DrainCallCount()).NotTo(Equal(0))
					})
				})

			})
		})

		Describe("Ping", func() {
			Context("when the driver pings", func() {
				BeforeEach(func() {
					err = driverAdminLocal.Ping(env)
				})

				It("should not fail", func() {
					Expect(err.Err).To(Equal(""))
				})
			})
		})
	})
})
