package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/nfsbroker/fakes"
	fuzz "github.com/google/gofuzz"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/pivotal-cf/brokerapi/v11/domain/apiresponses"
	"github.com/tedsuo/ifrit"
	ginkgomon "github.com/tedsuo/ifrit/ginkgomon_v2"
)

var _ = Describe("nfsbroker Main", func() {
	Context("Parse VCAP_SERVICES tests", func() {

		BeforeEach(func() {
			*cfServiceName = "postgresql"
		})
	})

	Context("Missing required args", func() {
		var process ifrit.Process

		It("shows usage when dataDir or dbDriver are not provided", func() {
			var args []string
			volmanRunner := failRunner{
				Name:       "nfsbroker",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "Either dataDir or credhubURL parameters must be provided.",
			}
			process = ifrit.Invoke(volmanRunner)
		})

		It("shows usage when servicesConfig is not provided", func() {
			args := []string{"-credhubURL", "some-credhub"}
			volmanRunner := failRunner{
				Name:       "nfsbroker",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "servicesConfig parameter must be provided.",
			}
			process = ifrit.Invoke(volmanRunner)
		})

		AfterEach(func() {
			ginkgomon.Kill(process) // this is only if incorrect implementation leaves process running
		})
	})

	Context("credhub /info returns error", func() {
		var volmanRunner *ginkgomon.Runner
		var credhubServer *ghttp.Server

		DescribeTable("should log a helpful diagnostic error message ", func(statusCode int) {
			listenAddr := "0.0.0.0:" + strconv.Itoa(8999+GinkgoParallelProcess())

			credhubServer = ghttp.NewServer()
			credhubServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/info"),
				ghttp.RespondWith(statusCode, "", http.Header{"X-Squid-Err": []string{"some-error"}}),
			))
			defer credhubServer.Close()

			var args []string
			args = append(args, "-listenAddr", listenAddr)
			args = append(args, "-credhubURL", credhubServer.URL())
			args = append(args, "-servicesConfig", "./default_services.json")

			volmanRunner = ginkgomon.New(ginkgomon.Config{
				Name:       "nfsbroker",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "starting",
			})

			invoke := ifrit.Invoke(volmanRunner)
			defer ginkgomon.Kill(invoke)

			time.Sleep(2 * time.Second)
			Eventually(volmanRunner.ExitCode).Should(Equal(2))
			Eventually(volmanRunner.Buffer()).Should(gbytes.Say(fmt.Sprintf(".*Attempted to connect to credhub. Expected 200. Got %d.*X-Squid-Err:\\[some-error\\].*", statusCode)))

		},
			Entry("300", http.StatusMultipleChoices),
			Entry("400", http.StatusBadRequest),
			Entry("403", http.StatusForbidden),
			Entry("500", http.StatusInternalServerError))

		It("should timeout after 30 seconds", func() {
			listenAddr := "0.0.0.0:" + strconv.Itoa(8999+GinkgoParallelProcess())

			var closeChan = make(chan interface{}, 1)

			credhubServer = ghttp.NewServer()
			credhubServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/info"),
				func(w http.ResponseWriter, r *http.Request) {
					<-closeChan
				},
			))

			var args []string
			args = append(args, "-listenAddr", listenAddr)
			args = append(args, "-credhubURL", credhubServer.URL())
			args = append(args, "-servicesConfig", "./default_services.json")

			volmanRunner = ginkgomon.New(ginkgomon.Config{
				Name:       "nfsbroker",
				Command:    exec.Command(binaryPath, args...),
				StartCheck: "starting",
			})

			invoke := ifrit.Invoke(volmanRunner)
			defer func() {
				close(closeChan)
				credhubServer.Close()
				ginkgomon.Kill(invoke)
			}()

			Eventually(volmanRunner.ExitCode, "35s", "1s").Should(Equal(2))
			Eventually(volmanRunner.Buffer, "35s", "1s").Should(gbytes.Say(".*Unable to connect to credhub."))
		})
	})

	Context("Has required args", func() {
		var (
			args               []string
			listenAddr         string
			username, password string
			volmanRunner       *ginkgomon.Runner
			planID             = "0da18102-48dc-46d0-98b3-7a4ff6dc9c54"
			serviceOfferingID  = "9db9cca4-8fd5-4b96-a4c7-0a48f47c3bad"
			serviceInstanceID  = "service-instance-id"

			process ifrit.Process

			credhubServer *ghttp.Server
			uaaServer     *ghttp.Server
		)

		BeforeEach(func() {
			listenAddr = "0.0.0.0:" + strconv.Itoa(7999+GinkgoParallelProcess())
			username = "admin"
			password = "password"

			os.Setenv("USERNAME", username)
			os.Setenv("PASSWORD", password)

			credhubServer = ghttp.NewServer()
			uaaServer = ghttp.NewServer()

			infoResponse := credhubInfoResponse{
				AuthServer: credhubInfoResponseAuthServer{
					URL: "some-auth-server-url",
				},
			}

			credhubServer.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/info"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, infoResponse),
			), ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/info"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, infoResponse),
			))

			args = append(args, "-credhubURL", credhubServer.URL())
			args = append(args, "-listenAddr", listenAddr)
			args = append(args, "-allowedOptions", "source,uid,gid,auto_cache,readonly,version,mount,cache")
			args = append(args, "-servicesConfig", "./test_default_services.json")
		})

		JustBeforeEach(func() {
			volmanRunner = ginkgomon.New(ginkgomon.Config{
				Name:              "nfsbroker",
				Command:           exec.Command(binaryPath, args...),
				StartCheck:        "started",
				StartCheckTimeout: 20 * time.Second,
			})
			process = ginkgomon.Invoke(volmanRunner)
		})

		AfterEach(func() {
			ginkgomon.Kill(process)
			credhubServer.Close()
			uaaServer.Close()
		})

		httpDoWithAuth := func(method, endpoint string, body io.Reader) (*http.Response, error) {
			req, err := http.NewRequest(method, "http://"+listenAddr+endpoint, body)
			req.Header.Add("X-Broker-Api-Version", "2.14")
			Expect(err).NotTo(HaveOccurred())

			req.SetBasicAuth(username, password)
			return http.DefaultClient.Do(req)
		}

		It("should check for a proxy", func() {
			Eventually(volmanRunner.Buffer()).Should(gbytes.Say("no-proxy-found"))
		})

		It("should listen on the given address", func() {
			resp, err := httpDoWithAuth("GET", "/v2/catalog", nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp.StatusCode).To(Equal(200))
		})

		It("should pass services config through to catalog", func() {
			resp, err := httpDoWithAuth("GET", "/v2/catalog", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			bytes, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			var catalog apiresponses.CatalogResponse
			err = json.Unmarshal(bytes, &catalog)
			Expect(err).NotTo(HaveOccurred())

			Expect(catalog.Services).To(HaveLen(2))

			Expect(catalog.Services[0].Name).To(Equal("nfs-legacy"))
			Expect(catalog.Services[0].ID).To(Equal("nfsbroker"))
			Expect(catalog.Services[0].Plans[0].ID).To(Equal("Existing"))
			Expect(catalog.Services[0].Plans[0].Name).To(Equal("Existing"))
			Expect(catalog.Services[0].Plans[0].Description).To(Equal("A preexisting filesystem"))

			Expect(catalog.Services[1].Name).To(Equal("nfs"))
			Expect(catalog.Services[1].ID).To(Equal("997f8f26-e10c-11e7-80c1-9a214cf093ae"))
			Expect(catalog.Services[1].Plans[0].ID).To(Equal("09a09260-1df5-4445-9ed7-1ba56dadbbc8"))
			Expect(catalog.Services[1].Plans[0].Name).To(Equal("Existing"))
			Expect(catalog.Services[1].Plans[0].Description).To(Equal("A preexisting filesystem"))
		})

		Context("#update", func() {

			It("should respond with a 422", func() {
				updateDetailsJson, err := json.Marshal(domain.UpdateDetails{
					ServiceID: "service-id",
				})
				Expect(err).NotTo(HaveOccurred())
				reader := strings.NewReader(string(updateDetailsJson))
				resp, err := httpDoWithAuth("PATCH", "/v2/service_instances/12345", reader)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(422))

				responseBody, err := io.ReadAll(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(responseBody)).To(ContainSubstring("This service does not support instance updates. Please delete your service instance and create a new one with updated configuration."))
			})

		})

		Context("#bind", func() {
			var (
				bindingID = "456"
			)
			BeforeEach(func() {
				infoResponse := credhubInfoResponse{
					AuthServer: credhubInfoResponseAuthServer{
						URL: uaaServer.URL(),
					},
				}

				uaaServer.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token"),
					ghttp.RespondWith(http.StatusOK, `{ "access_token" : "111", "refresh_token" : "", "token_type" : "" }`),
				))

				credhubServer.RouteToHandler("GET", "/info", ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/info"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, infoResponse),
				))

				credhubServer.RouteToHandler("GET", "/api/v1/data", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.Contains(r.URL.RawQuery, bindingID) {
						w.WriteHeader(404)
					} else if strings.Contains(r.URL.RawQuery, fmt.Sprintf("current=true&name=%%2Fnfsbroker%%2F%s", serviceInstanceID)) {
						_, err := w.Write([]byte(`{ "data" : [ { "type": "value", "version_created_at": "2019", "id": "1", "name": "/some-name", "value": { "ServiceFingerPrint": "foobar" } } ] }`))
						if err != nil {
							w.WriteHeader(500)
						}
					}
				}))

				credhubServer.RouteToHandler("GET", "/version", ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/version"),
					ghttp.RespondWith(http.StatusOK, `{ "version" : "0.0.0" }`),
				))

				credhubServer.RouteToHandler("PUT", "/api/v1/data", ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v1/data"),
					ghttp.RespondWith(http.StatusCreated, `{ "type" : "json", "version_created_at" : "", "id" : "", "name" : "", "value" : { } }`),
				))
			})

			Context("allowed parameters", func() {
				It("should accept the parameter", func() {
					rawParametersMap := map[string]string{
						"uid":      "1",
						"gid":      "1",
						"mount":    "somemount",
						"readonly": "true",
						"cache":    "true",
						"version":  "4.2",
					}

					rawParameters, err := json.Marshal(rawParametersMap)
					Expect(err).NotTo(HaveOccurred())
					provisionDetailsJsons, err := json.Marshal(domain.BindDetails{
						ServiceID:     serviceOfferingID,
						PlanID:        planID,
						AppGUID:       "222",
						RawParameters: rawParameters,
					})
					Expect(err).NotTo(HaveOccurred())
					reader := strings.NewReader(string(provisionDetailsJsons))
					endpoint := fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", serviceInstanceID, bindingID)
					resp, err := httpDoWithAuth("PUT", endpoint, reader)

					Expect(err).NotTo(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(201))
				})
			})

			Context("invalid cache", func() {
				var (
					bindDetailJson []byte
					cache          = ""
				)

				BeforeEach(func() {
					fuzz.New().Fuzz(&cache)
					cache = strings.ReplaceAll(cache, "%", "")

					rawParametersMap := map[string]string{
						"cache": cache,
					}

					rawParameters, err := json.Marshal(rawParametersMap)
					Expect(err).NotTo(HaveOccurred())

					bindDetailJson, err = json.Marshal(domain.BindDetails{
						ServiceID:     serviceOfferingID,
						PlanID:        planID,
						AppGUID:       "222",
						RawParameters: rawParameters,
					})

					Expect(err).NotTo(HaveOccurred())
				})

				It("should respond with 400", func() {
					reader := strings.NewReader(string(bindDetailJson))
					endpoint := fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", serviceInstanceID, bindingID)
					resp, err := httpDoWithAuth("PUT", endpoint, reader)

					Expect(err).NotTo(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(400))

					expectedResponse := map[string]string{
						"description": fmt.Sprintf("- validation mount options failed: %s is not a valid value for cache\n", cache),
					}
					expectedJsonResponse, err := json.Marshal(expectedResponse)
					Expect(err).NotTo(HaveOccurred())

					responseBody, err := io.ReadAll(resp.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(responseBody)).To(MatchJSON(expectedJsonResponse))
				})
			})
		})
	})

	Context("#IsRetired", func() {
		var (
			fakeRetiredStore *fakes.FakeRetiredStore
			retired          bool
			err              error
		)

		JustBeforeEach(func() {
			retired, err = IsRetired(fakeRetiredStore)
		})

		BeforeEach(func() {
			fakeRetiredStore = &fakes.FakeRetiredStore{}
		})

		Context("when the store is not a RetireableStore", func() {
			BeforeEach(func() {
				fakeRetiredStore.IsRetiredReturns(false, nil)
			})

			It("should return false", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(retired).To(BeFalse())
			})
		})

		Context("when the store is a RetiredStore", func() {
			Context("when the store is retired", func() {
				BeforeEach(func() {
					fakeRetiredStore.IsRetiredReturns(true, nil)
				})

				It("should return true", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(retired).To(BeTrue())
				})
			})

			Context("when the store is not retired", func() {
				BeforeEach(func() {
					fakeRetiredStore.IsRetiredReturns(false, nil)
				})

				It("should return false", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(retired).To(BeFalse())
				})
			})

			Context("when the IsRetired check fails", func() {
				BeforeEach(func() {
					fakeRetiredStore.IsRetiredReturns(false, errors.New("is-retired-failed"))
				})

				It("should return true", func() {
					Expect(err).To(MatchError("is-retired-failed"))
				})
			})
		})
	})
})

type failRunner struct {
	Command           *exec.Cmd
	Name              string
	AnsiColorCode     string
	StartCheck        string
	StartCheckTimeout time.Duration
	Cleanup           func()
	session           *gexec.Session
	sessionReady      chan struct{}
}

func (r failRunner) Run(sigChan <-chan os.Signal, ready chan<- struct{}) error {
	defer GinkgoRecover()

	allOutput := gbytes.NewBuffer()

	debugWriter := gexec.NewPrefixedWriter(
		fmt.Sprintf("\x1b[32m[d]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
		GinkgoWriter,
	)

	var err error
	r.session, err = gexec.Start(
		r.Command,
		gexec.NewPrefixedWriter(
			fmt.Sprintf("\x1b[32m[o]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
			io.MultiWriter(allOutput, GinkgoWriter),
		),
		gexec.NewPrefixedWriter(
			fmt.Sprintf("\x1b[91m[e]\x1b[%s[%s]\x1b[0m ", r.AnsiColorCode, r.Name),
			io.MultiWriter(allOutput, GinkgoWriter),
		),
	)

	Ω(err).ShouldNot(HaveOccurred())

	fmt.Fprintf(debugWriter, "spawned %s (pid: %d)\n", r.Command.Path, r.Command.Process.Pid)

	if r.sessionReady != nil {
		close(r.sessionReady)
	}

	startCheckDuration := r.StartCheckTimeout
	if startCheckDuration == 0 {
		startCheckDuration = 5 * time.Second
	}

	var startCheckTimeout <-chan time.Time
	if r.StartCheck != "" {
		startCheckTimeout = time.After(startCheckDuration)
	}

	detectStartCheck := allOutput.Detect(r.StartCheck)

	for {
		select {
		case <-detectStartCheck: // works even with empty string
			allOutput.CancelDetects()
			startCheckTimeout = nil
			detectStartCheck = nil
			close(ready)

		case <-startCheckTimeout:
			// clean up hanging process
			r.session.Kill().Wait()

			// fail to start
			return fmt.Errorf(
				"did not see %s in command's output within %s. full output:\n\n%s",
				r.StartCheck,
				startCheckDuration,
				string(allOutput.Contents()),
			)

		case signal := <-sigChan:
			r.session.Signal(signal)

		case <-r.session.Exited:
			if r.Cleanup != nil {
				r.Cleanup()
			}

			Expect(string(allOutput.Contents())).To(ContainSubstring(r.StartCheck))
			Expect(r.session.ExitCode()).To(Not(Equal(0)), "Expected process to exit with non-zero, got: 0")
			return nil
		}
	}
}

type credhubInfoResponse struct {
	AuthServer credhubInfoResponseAuthServer `json:"auth-server"`
}

type credhubInfoResponseAuthServer struct {
	URL string `json:"url"`
}
