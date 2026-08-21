module code.cloudfoundry.org/nfsv3driver

require (
	code.cloudfoundry.org/cf-networking-helpers v0.96.0
	code.cloudfoundry.org/debugserver v0.110.0
	code.cloudfoundry.org/dockerdriver v0.101.0
	code.cloudfoundry.org/goshims v0.109.0
	code.cloudfoundry.org/lager/v3 v3.82.0
	code.cloudfoundry.org/tlsconfig v0.64.0
	code.cloudfoundry.org/volume-mount-options v0.161.0
	code.cloudfoundry.org/volumedriver v0.184.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.9.0
	github.com/onsi/ginkgo/v2 v2.32.1
	github.com/onsi/gomega v1.42.1
	github.com/tedsuo/ifrit v0.0.0-20260418191334-846868129986
	github.com/tedsuo/rata v1.0.0
	gopkg.in/ldap.v2 v2.5.1
)

require (
	code.cloudfoundry.org/cfhttp/v2 v2.90.0 // indirect
	code.cloudfoundry.org/clock v1.83.0 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/bmizerany/pat v0.0.0-20210406213842-e4b6760bdd6f // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260802141513-ef3492d7dac3 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20181015200546-f715ec2f112d // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)

go 1.25.8

replace (
	// pin ifrit until https://github.com/tedsuo/ifrit/pull/48 is merged
	github.com/tedsuo/ifrit => github.com/tedsuo/ifrit v0.0.0-20260418191334-846868129986
)
