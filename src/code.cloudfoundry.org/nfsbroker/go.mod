module code.cloudfoundry.org/nfsbroker

go 1.25.7

require (
	code.cloudfoundry.org/brokerapi/v13 v13.0.25
	code.cloudfoundry.org/clock v1.85.0
	code.cloudfoundry.org/debugserver v0.112.0
	code.cloudfoundry.org/existingvolumebroker v0.228.0
	code.cloudfoundry.org/goshims v0.110.0
	code.cloudfoundry.org/lager/v3 v3.84.0
	code.cloudfoundry.org/service-broker-store v0.167.0
	code.cloudfoundry.org/volume-mount-options v0.164.0
	github.com/google/gofuzz v1.2.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.12.2
	github.com/onsi/ginkgo/v2 v2.32.1
	github.com/onsi/gomega v1.43.0
	github.com/tedsuo/ifrit v0.0.0-20260813155221-94822c932811
)

require (
	code.cloudfoundry.org/credhub-cli v0.0.0-20260824191323-ca2ae25cb1f8 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/cloudfoundry/go-socks5 v0.0.0-20250423223041-4ad5fea42851 // indirect
	github.com/cloudfoundry/socks5-proxy v0.2.186 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260830191439-4932ad3515ea // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

// pin ifrit until https://github.com/tedsuo/ifrit/pull/48 is merged
replace github.com/tedsuo/ifrit => github.com/tedsuo/ifrit v0.0.0-20260418191334-846868129986
