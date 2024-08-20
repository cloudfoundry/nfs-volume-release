package driveradminhttp_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/nfsv3driver/driveradmin"
	"code.cloudfoundry.org/nfsv3driver/driveradmin/driveradminhttp"
	"code.cloudfoundry.org/nfsv3driver/nfsdriverfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/rata"
)

var _ = Describe("Volman Driver Handlers", func() {

	Context("when generating http handlers", func() {
		var (
			testLogger           = lagertest.NewTestLogger("HandlersTest")
			fakeDriverAdmin      = &nfsdriverfakes.FakeDriverAdmin{}
			handler              http.Handler
			httpRequest          *http.Request
			httpResponseRecorder *httptest.ResponseRecorder
			route                rata.Route
		)

		BeforeEach(func() {
			var err error
			handler, err = driveradminhttp.NewHandler(testLogger, fakeDriverAdmin)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			var err error
			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			httpRequest, err = http.NewRequest("GET", path, nil)
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder = httptest.NewRecorder()
			handler.ServeHTTP(httpResponseRecorder, httpRequest)
		})

		Context("Evacuate", func() {
			BeforeEach(func() {
				fakeDriverAdmin.EvacuateReturns(driveradmin.ErrorResponse{})

				var found bool
				route, found = driveradmin.Routes.FindRouteByName(driveradmin.EvacuateRoute)
				Expect(found).To(BeTrue())
			})

			It("should produce a handler with an evacuate route", func() {
				Expect(httpResponseRecorder.Code).To(Equal(200))
				Expect(httpResponseRecorder.Body).Should(MatchJSON(`{"Err":""}`))
			})

			Context("when invoking evacuate returns an error", func() {
				BeforeEach(func() {
					fakeDriverAdmin.EvacuateReturns(driveradmin.ErrorResponse{
						Err: "unable to evacuate",
					})
				})

				It("should return an http 500 response and an error string", func() {
					Expect(httpResponseRecorder.Code).To(Equal(500))
					Expect(httpResponseRecorder.Body).Should(MatchJSON(`{"Err":"unable to evacuate"}`))
				})
			})
		})

		Context("Ping", func() {
			BeforeEach(func() {
				fakeDriverAdmin.EvacuateReturns(driveradmin.ErrorResponse{})

				var found bool
				route, found = driveradmin.Routes.FindRouteByName(driveradmin.PingRoute)
				Expect(found).To(BeTrue())
			})

			It("should produce a handler with an ping route", func() {
				Expect(httpResponseRecorder.Code).To(Equal(200))
				Expect(httpResponseRecorder.Body).Should(MatchJSON(`{"Err":""}`))
			})

			Context("when invoking ping returns an error", func() {
				BeforeEach(func() {
					fakeDriverAdmin.PingReturns(driveradmin.ErrorResponse{
						Err: "unable to ping",
					})
				})

				It("should return an http 500 response and an error string", func() {
					Expect(httpResponseRecorder.Code).To(Equal(500))
					Expect(httpResponseRecorder.Body).Should(MatchJSON(`{"Err":"unable to ping"}`))
				})
			})
		})

	})
})
