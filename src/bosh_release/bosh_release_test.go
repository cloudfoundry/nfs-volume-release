package bosh_release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BoshReleaseTest", func() {

	It("should have a nfsv3driver process running", func() {
		state := findProcessState("nfsv3driver")

		Expect(state).To(Equal("running"))
	})

})
