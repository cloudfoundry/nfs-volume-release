module code.cloudfoundry.org/mapfs

require (
	code.cloudfoundry.org/goshims v0.121.0
	code.cloudfoundry.org/volumedriver v0.200.0
	github.com/hanwen/go-fuse/v2 v2.12.1
	github.com/maxbrunsfeld/counterfeiter/v6 v6.9.0
	github.com/onsi/ginkgo/v2 v2.34.0
	github.com/onsi/gomega v1.44.0
	github.com/tedsuo/ifrit v0.0.0-20260908181113-dd353a7daa27
	golang.org/x/sys v0.50.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20261111200114-fac431ef4d26 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/mod v0.43.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
)

go 1.27.3

// pin ifrit until https://github.com/tedsuo/ifrit/pull/48 is merged
replace github.com/tedsuo/ifrit => github.com/tedsuo/ifrit v0.0.0-20260418191334-846868129986
