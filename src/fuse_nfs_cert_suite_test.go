package fuse_nfs_certs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSrc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FuseNFSCerts Suite")
}
