module code.cloudfoundry.org/nfsv3driver

require (
	code.cloudfoundry.org/cf-networking-helpers v0.101.0
	code.cloudfoundry.org/debugserver v0.121.0
	code.cloudfoundry.org/dockerdriver v0.113.0
	code.cloudfoundry.org/goshims v0.119.0
	code.cloudfoundry.org/lager/v3 v3.93.0
	code.cloudfoundry.org/tlsconfig v0.70.0
	code.cloudfoundry.org/volume-mount-options v0.173.0
	code.cloudfoundry.org/volumedriver v0.198.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.9.0
	github.com/onsi/ginkgo/v2 v2.33.1
	github.com/onsi/gomega v1.44.0
	github.com/tedsuo/ifrit v0.0.0-20260908181113-dd353a7daa27
	github.com/tedsuo/rata v1.0.0
	gopkg.in/ldap.v2 v2.5.1
)

require (
	code.cloudfoundry.org/cfhttp/v2 v2.101.0 // indirect
	code.cloudfoundry.org/clock v1.94.0 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/bmizerany/pat v0.0.0-20210406213842-e4b6760bdd6f // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20261026192309-dbd0ed30aa88 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
)

go 1.27.3

// pin ifrit until https://github.com/tedsuo/ifrit/pull/48 is merged
replace github.com/tedsuo/ifrit => github.com/tedsuo/ifrit v0.0.0-20260418191334-846868129986
