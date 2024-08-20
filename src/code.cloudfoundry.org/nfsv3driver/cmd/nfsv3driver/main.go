package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"code.cloudfoundry.org/tlsconfig"

	cf_debug_server "code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/goshims/bufioshim"
	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/ioutilshim"
	"code.cloudfoundry.org/goshims/ldapshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/goshims/syscallshim"
	"code.cloudfoundry.org/goshims/timeshim"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagerflags"
	"code.cloudfoundry.org/nfsv3driver"
	"code.cloudfoundry.org/nfsv3driver/driveradmin/driveradminhttp"
	"code.cloudfoundry.org/nfsv3driver/driveradmin/driveradminlocal"
	"code.cloudfoundry.org/volumedriver"
	"code.cloudfoundry.org/volumedriver/invoker"
	"code.cloudfoundry.org/volumedriver/mountchecker"
	"code.cloudfoundry.org/volumedriver/oshelper"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var atAddress = flag.String(
	"listenAddr",
	"127.0.0.1:7589",
	"host:port to serve volume management functions",
)

var adminAddress = flag.String(
	"adminAddr",
	"127.0.0.1:7590",
	"host:port to serve process admin functions",
)

var driversPath = flag.String(
	"driversPath",
	"",
	"Path to directory where drivers are installed",
)

var transport = flag.String(
	"transport",
	"tcp",
	"Transport protocol to transmit HTTP over",
)

var mapfsPath = flag.String(
	"mapfsPath",
	"/var/vcap/packages/mapfs/bin/mapfs",
	"Path to the mapfs binary",
)

var mountDir = flag.String(
	"mountDir",
	"/tmp/volumes",
	"Path to directory where NFS v3 volumes are created",
)

var requireSSL = flag.Bool(
	"requireSSL",
	false,
	"whether the NFS v3 driver should require ssl-secured communication",
)

var caFile = flag.String(
	"caFile",
	"",
	"the certificate authority public key file to use with ssl authentication",
)

var certFile = flag.String(
	"certFile",
	"",
	"the public key file to use with ssl authentication",
)

var keyFile = flag.String(
	"keyFile",
	"",
	"the private key file to use with ssl authentication",
)
var clientCertFile = flag.String(
	"clientCertFile",
	"",
	"the public key file to use with client ssl authentication",
)

var clientKeyFile = flag.String(
	"clientKeyFile",
	"",
	"the private key file to use with client ssl authentication",
)

var insecureSkipVerify = flag.Bool(
	"insecureSkipVerify",
	false,
	"whether SSL communication should skip verification of server IP addresses in the certificate",
)

const fsType = "nfs"
const mountOptions = "rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2,actimeo=0"

var (
	ldapSvcUser  string
	ldapSvcPass  string
	ldapUserFqdn string
	ldapHost     string
	ldapPort     int
	ldapCACert   string
	ldapProto    string
	ldapTimeout  int
)

func main() {
	parseCommandLine()
	parseEnvironment()

	var nfsDriverServer ifrit.Runner
	var idResolver nfsv3driver.IdResolver
	var mounter volumedriver.Mounter

	logger, logSink := newLogger()
	logger.Info("start")
	defer logger.Info("end")

	if ldapHost != "" {
		idResolver = nfsv3driver.NewLdapIdResolver(
			ldapSvcUser,
			ldapSvcPass,
			ldapHost,
			ldapPort,
			ldapProto,
			ldapUserFqdn,
			ldapCACert,
			&ldapshim.LdapShim{},
			time.Duration(ldapTimeout)*time.Second,
		)
	}

	mask, err := nfsv3driver.NewMapFsVolumeMountMask()
	if err != nil {
		exitOnFailure(logger, err)
	}

	processGroupInvoker := invoker.NewProcessGroupInvoker()
	mounter = nfsv3driver.NewMapfsMounter(
		processGroupInvoker,
		&osshim.OsShim{},
		&syscallshim.SyscallShim{},
		&ioutilshim.IoutilShim{},
		mountchecker.NewChecker(&bufioshim.BufioShim{}, &osshim.OsShim{}),
		fsType,
		mountOptions,
		idResolver,
		mask,
		*mapfsPath,
	)

	client := volumedriver.NewVolumeDriver(
		logger,
		&osshim.OsShim{},
		&filepathshim.FilepathShim{},
		&ioutilshim.IoutilShim{},
		&timeshim.TimeShim{},
		mountchecker.NewChecker(&bufioshim.BufioShim{}, &osshim.OsShim{}),
		*mountDir,
		mounter,
		oshelper.NewOsHelper(),
	)

	if *transport == "tcp" {
		nfsDriverServer = createNfsDriverServer(logger, client, *atAddress, *driversPath, false)
	} else if *transport == "tcp-json" {
		nfsDriverServer = createNfsDriverServer(logger, client, *atAddress, *driversPath, true)
	} else {
		nfsDriverServer = createNfsDriverUnixServer(logger, client, *atAddress)
	}

	servers := grouper.Members{
		{Name: "nfsdriver-server", Runner: nfsDriverServer},
	}

	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		servers = append(grouper.Members{
			{Name: "debug-server", Runner: cf_debug_server.Runner(dbgAddr, logSink)},
		}, servers...)
	}

	adminClient := driveradminlocal.NewDriverAdminLocal()
	adminHandler, _ := driveradminhttp.NewHandler(logger, adminClient)
	adminServer := http_server.New(*adminAddress, adminHandler)

	servers = append(grouper.Members{
		{Name: "driveradmin", Runner: adminServer},
	}, servers...)

	process := ifrit.Invoke(processRunnerFor(servers))
	logger.Info("started")

	adminClient.SetServerProc(process)
	adminClient.RegisterDrainable(client)

	untilTerminated(logger, process)
}

func exitOnFailure(logger lager.Logger, err error) {
	if err != nil {
		logger.Fatal("fatal-err-aborting", err)
	}
}

func untilTerminated(logger lager.Logger, process ifrit.Process) {
	err := <-process.Wait()
	exitOnFailure(logger, err)
}

func processRunnerFor(servers grouper.Members) ifrit.Runner {
	return sigmon.New(grouper.NewOrdered(os.Interrupt, servers))
}

func createNfsDriverServer(logger lager.Logger, client dockerdriver.Driver, atAddress, driversPath string, jsonSpec bool) ifrit.Runner {
	advertisedUrl := "http://" + atAddress
	logger.Info("writing-spec-file", lager.Data{"location": driversPath, "name": "nfsv3driver", "address": advertisedUrl})
	if jsonSpec {
		driverJsonSpec := dockerdriver.DriverSpec{Name: "nfsv3driver", Address: advertisedUrl, UniqueVolumeIds: true}

		if *requireSSL {
			absCaFile, err := filepath.Abs(*caFile)
			exitOnFailure(logger, err)
			absClientCertFile, err := filepath.Abs(*clientCertFile)
			exitOnFailure(logger, err)
			absClientKeyFile, err := filepath.Abs(*clientKeyFile)
			exitOnFailure(logger, err)
			driverJsonSpec.TLSConfig = &dockerdriver.TLSConfig{InsecureSkipVerify: *insecureSkipVerify, CAFile: absCaFile, CertFile: absClientCertFile, KeyFile: absClientKeyFile}
			driverJsonSpec.Address = "https://" + atAddress
		}

		jsonBytes, err := json.Marshal(driverJsonSpec)

		exitOnFailure(logger, err)
		err = dockerdriver.WriteDriverSpec(logger, driversPath, "nfsv3driver", "json", jsonBytes)
		exitOnFailure(logger, err)
	} else {
		err := dockerdriver.WriteDriverSpec(logger, driversPath, "nfsv3driver", "spec", []byte(advertisedUrl))
		exitOnFailure(logger, err)
	}

	handler, err := driverhttp.NewHandler(logger, client)

	exitOnFailure(logger, err)

	var server ifrit.Runner
	if *requireSSL {
		tlsConfig, err := tlsconfig.
			Build(
				tlsconfig.WithIdentityFromFile(*certFile, *keyFile),
				tlsconfig.WithInternalServiceDefaults(),
			).
			Server(tlsconfig.WithClientAuthenticationFromFile(*caFile))
		if err != nil {
			logger.Fatal("tls-configuration-failed", err)
		}
		server = http_server.NewTLSServer(atAddress, handler, tlsConfig)
	} else {
		server = http_server.New(atAddress, handler)
	}

	return server
}

func createNfsDriverUnixServer(logger lager.Logger, client dockerdriver.Driver, atAddress string) ifrit.Runner {
	handler, err := driverhttp.NewHandler(logger, client)
	exitOnFailure(logger, err)
	return http_server.NewUnixServer(atAddress, handler)
}

func newLogger() (lager.Logger, *lager.ReconfigurableSink) {
	lagerConfig := lagerflags.ConfigFromFlags()
	lagerConfig.RedactSecrets = true

	return lagerflags.NewFromConfig("nfs-driver-server", lagerConfig)
}

func parseCommandLine() {
	lagerflags.AddFlags(flag.CommandLine)
	cf_debug_server.AddFlags(flag.CommandLine)
	flag.Parse()
}

func parseEnvironment() {
	ldapSvcUser, _ = os.LookupEnv("LDAP_SVC_USER")
	ldapSvcPass, _ = os.LookupEnv("LDAP_SVC_PASS")
	ldapUserFqdn, _ = os.LookupEnv("LDAP_USER_FQDN")
	ldapHost, _ = os.LookupEnv("LDAP_HOST")
	port, _ := os.LookupEnv("LDAP_PORT")
	ldapPort, _ = strconv.Atoi(port)
	ldapCACert, _ = os.LookupEnv("LDAP_CA_CERT")
	ldapProto, _ = os.LookupEnv("LDAP_PROTO")
	timeout, _ := os.LookupEnv("LDAP_TIMEOUT")
	ldapTimeout, _ = strconv.Atoi(timeout)

	if ldapProto == "" {
		ldapProto = "tcp"
	}

	if ldapHost != "" && (ldapSvcUser == "" || ldapSvcPass == "" || ldapUserFqdn == "" || ldapPort == 0) {
		panic("LDAP is enabled but required LDAP parameters are not set.")
	}

	if ldapTimeout < 0 {
		panic("LDAP_TIMEOUT is set to negtive value")
	}

	// if ldapTimeout is not set, use default value
	if ldapTimeout == 0 {
		ldapTimeout = 120
	}
}
