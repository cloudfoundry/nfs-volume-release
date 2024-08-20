package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/existingvolumebroker"
	evbutils "code.cloudfoundry.org/existingvolumebroker/utils"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagerflags"
	"code.cloudfoundry.org/nfsbroker/utils"
	"code.cloudfoundry.org/service-broker-store/brokerstore"
	vmo "code.cloudfoundry.org/volume-mount-options"
	vmou "code.cloudfoundry.org/volume-mount-options/utils"
	"github.com/pivotal-cf/brokerapi/v11"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
)

var dataDir = flag.String(
	"dataDir",
	"",
	"[REQUIRED] - Broker's state will be stored here to persist across reboots",
)

var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:8999",
	"host:port to serve service broker API",
)

var servicesConfig = flag.String(
	"servicesConfig",
	"",
	"[REQUIRED] - Path to services config to register with cloud controller",
)

var cfServiceName = flag.String(
	"cfServiceName",
	"",
	"(optional) For CF pushed apps, the service name in VCAP_SERVICES where we should find database credentials.  dbDriver must be defined if this option is set, but all other db parameters will be extracted from the service binding.",
)

var allowedOptions = flag.String(
	"allowedOptions",
	"auto_cache,uid,gid",
	"A comma separated list of parameters allowed to be set in config.",
)

var defaultOptions = flag.String(
	"defaultOptions",
	"auto_cache:true",
	"A comma separated list of defaults specified as param:value. If a parameter has a default value and is not in the allowed list, this default value becomes a fixed value that cannot be overridden",
)

var credhubURL = flag.String(
	"credhubURL",
	"",
	"(optional) CredHub server URL when using CredHub to store broker state",
)

var credhubCACertPath = flag.String(
	"credhubCACertPath",
	"",
	"(optional) Path to CA Cert for CredHub",
)

var uaaClientID = flag.String(
	"uaaClientID",
	"",
	"(optional) UAA client ID when using CredHub to store broker state",
)

var uaaClientSecret = flag.String(
	"uaaClientSecret",
	"",
	"(optional) UAA client secret when using CredHub to store broker state",
)

var uaaCACertPath = flag.String(
	"uaaCACertPath",
	"",
	"(optional) Path to CA Cert for UAA used for CredHub authorization",
)

var storeID = flag.String(
	"storeID",
	"nfsbroker",
	"(optional) Store ID used to namespace instance details and bindings (credhub only)",
)

var (
	username string
	password string
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes/retired_store_fake.go . RetiredStore
type RetiredStore interface {
	IsRetired() (bool, error)
	brokerstore.Store
}

func main() {
	parseCommandLine()
	parseEnvironment()

	checkParams()

	logger, logSink := newLogger()
	logger.Info("starting")
	defer logger.Info("ends")

	verifyCredhubIsReachable(logger)

	server := createServer(logger)

	if dbgAddr := debugserver.DebugAddress(flag.CommandLine); dbgAddr != "" {
		server = utils.ProcessRunnerFor(grouper.Members{
			{Name: "debug-server", Runner: debugserver.Runner(dbgAddr, logSink)},
			{Name: "broker-api", Runner: server},
		})
	}

	process := ifrit.Invoke(server)
	logger.Info("started")
	utils.UntilTerminated(logger, process)
}

func parseCommandLine() {
	lagerflags.AddFlags(flag.CommandLine)
	debugserver.AddFlags(flag.CommandLine)
	flag.Parse()
}

func parseEnvironment() {
	username, _ = os.LookupEnv("USERNAME")
	password, _ = os.LookupEnv("PASSWORD")
	uaaClientSecretString, _ := os.LookupEnv("UAA_CLIENT_SECRET")
	if uaaClientSecretString != "" {
		uaaClientSecret = &uaaClientSecretString
	}
	uaaClientIDString, _ := os.LookupEnv("UAA_CLIENT_ID")
	if uaaClientIDString != "" {
		uaaClientID = &uaaClientIDString
	}

}

func checkParams() {
	if *dataDir == "" && *credhubURL == "" {
		fmt.Fprint(os.Stderr, "\nERROR: Either dataDir or credhubURL parameters must be provided.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *servicesConfig == "" {
		fmt.Fprint(os.Stderr, "\nERROR: servicesConfig parameter must be provided.\n\n")
		flag.Usage()
		os.Exit(1)
	}
}

func newLogger() (lager.Logger, *lager.ReconfigurableSink) {
	lagerConfig := lagerflags.ConfigFromFlags()
	lagerConfig.RedactSecrets = true

	return lagerflags.NewFromConfig("nfsbroker", lagerConfig)
}

func verifyCredhubIsReachable(logger lager.Logger) {
	var client = &http.Client{
		Timeout: 30 * time.Second,
	}

	configureCACert(logger, client)

	evbutils.IsThereAProxy(&osshim.OsShim{}, logger)

	resp, err := client.Get(*credhubURL + "/info")
	if err != nil {
		logger.Fatal("Unable to connect to credhub", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Fatal(fmt.Sprintf("Attempted to connect to credhub. Expected 200. Got %d", resp.StatusCode), nil, lager.Data{"response_headers": fmt.Sprintf("%v", resp.Header)})
	}
}

func configureCACert(logger lager.Logger, client *http.Client) {
	if *credhubCACertPath != "" {
		certpool := x509.NewCertPool()

		certPEM, err := os.ReadFile(*credhubCACertPath)
		if err != nil {
			logger.Fatal("reading credhub ca cert path", err)
		}

		ok := certpool.AppendCertsFromPEM(certPEM)
		if !ok {
			logger.Fatal("appending certs from PEM", err)
		}
		// disable "G402 (CWE-295): TLS MinVersion too low. (Confidence: HIGH, Severity: HIGH)"
		// #nosec G402 - Enforcing a MinVersion for TLS could break numerous existing systems
		clientTLSConf := &tls.Config{
			RootCAs: certpool,
		}

		transport := &http.Transport{
			TLSClientConfig: clientTLSConf,
		}

		client.Transport = transport
	}
}

func parseVcapServices(logger lager.Logger, os osshim.Os) {
	// populate db parameters from VCAP_SERVICES and pitch a fit if there isn't one.
	services, hasValue := os.LookupEnv("VCAP_SERVICES")
	if !hasValue {
		logger.Fatal("missing-vcap-services-environment", errors.New("missing VCAP_SERVICES environment"))
	}

	stuff := map[string][]interface{}{}
	err := json.Unmarshal([]byte(services), &stuff)
	if err != nil {
		logger.Fatal("json-unmarshal-error", err)
	}

	stuff2, ok := stuff[*cfServiceName]
	if !ok {
		logger.Fatal("missing-service-binding", errors.New("VCAP_SERVICES missing specified db service"), lager.Data{"stuff": stuff})
	}

	stuff3 := stuff2[0].(map[string]interface{})

	credentials := stuff3["credentials"].(map[string]interface{})
	logger.Debug("credentials-parsed", lager.Data{"credentials": credentials})

}

func createServer(logger lager.Logger) ifrit.Runner {
	if isCfPushed() {
		parseVcapServices(logger, &osshim.OsShim{})
	}

	var credhubCACert string
	if *credhubCACertPath != "" {
		b, err := os.ReadFile(*credhubCACertPath)
		if err != nil {
			logger.Fatal("cannot-read-credhub-ca-cert", err, lager.Data{"path": *credhubCACertPath})
		}
		credhubCACert = string(b)
	}

	var uaaCACert string
	if *uaaCACertPath != "" {
		b, err := os.ReadFile(*uaaCACertPath)
		if err != nil {
			logger.Fatal("cannot-read-credhub-ca-cert", err, lager.Data{"path": *uaaCACertPath})
		}
		uaaCACert = string(b)
	}

	store := brokerstore.NewStore(
		logger,
		*credhubURL,
		credhubCACert,
		*uaaClientID,
		*uaaClientSecret,
		uaaCACert,
		*storeID,
	)

	retired, err := IsRetired(store)
	if err != nil {
		logger.Fatal("check-is-retired-failed", err)
	}

	if retired {
		logger.Fatal("retired-store", errors.New("store is retired"))
	}

	cacheOptsValidator := vmo.UserOptsValidationFunc(validateCache)

	configMask, err := vmo.NewMountOptsMask(
		strings.Split(*allowedOptions, ","),
		vmou.ParseOptionStringToMap(*defaultOptions, ":"),
		map[string]string{
			"share": "source",
		},
		[]string{},
		[]string{"source"},
		cacheOptsValidator,
	)
	if err != nil {
		logger.Fatal("creating-config-mask-error", err)
	}

	logger.Debug("nfsbroker-startup-config", lager.Data{"config-mask": configMask})

	services, err := NewServicesFromConfig(*servicesConfig)
	if err != nil {
		logger.Fatal("loading-services-config-error", err)
	}

	serviceBroker := existingvolumebroker.New(
		existingvolumebroker.BrokerTypeNFS,
		logger,
		services,
		&osshim.OsShim{},
		clock.NewClock(),
		store,
		configMask,
	)

	credentials := brokerapi.BrokerCredentials{Username: username, Password: password}
	handler := brokerapi.New(serviceBroker, slog.New(lager.NewHandler(logger.Session("broker-api"))), credentials)

	return http_server.New(*atAddress, handler)
}

func isCfPushed() bool {
	return *cfServiceName != ""
}

func IsRetired(store brokerstore.Store) (bool, error) {
	if retiredStore, ok := store.(RetiredStore); ok {
		return retiredStore.IsRetired()
	}
	return false, nil
}

func validateCache(key string, val string) error {

	if key != "cache" {
		return nil
	}

	_, err := strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("%s is not a valid value for cache", val)
	}

	return nil
}
