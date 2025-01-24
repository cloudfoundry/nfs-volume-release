package perf_test

import (
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/gmeasure"
)

const HIGH_LATENCY_MILLIS = "80ms"
const LOW_LATENCY_MILLIS = "10ms"
const numWrites = "count=100"

var _ = Describe("Perf", func() {
	var nativeDirectory string
	var mapfsDirectory string

	BeforeEach(func() {
		var err error

		nativeDirectory, err = os.MkdirTemp("", "native")
		Expect(err).NotTo(HaveOccurred())
		mapfsDirectory, err = os.MkdirTemp("", "mapfs")
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		unshapeTraffic()

		cmd := exec.Command("umount", "-l", nativeDirectory)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
	})

	It("MapFs should not perform writes *much* slower than writing to a native mount", Serial, func() {
		exp := gmeasure.NewExperiment("Mounting via nfs")
		AddReportEntry(exp.Name, exp)

		By("Natively mounting via nfs", func() {
			url := strings.TrimPrefix(config.NfsServer, "nfs://")
			cmd := exec.Command("mount", "-t", "nfs", "-o", "rsize=1048576,wsize=1048576,timeo=600,retrans=2,actimeo=0", url, nativeDirectory)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0), string(session.Out.Contents()))
		})

		exp.MeasureDuration("native-writing-without-throttling", func() { writeDataToMountedDirectory(nativeDirectory) })

		exp.SampleDuration("native-writing-with-throttling-and-high-latencty-network", func(_ int) {
			shapeTraffic(HIGH_LATENCY_MILLIS)
			writeDataToMountedDirectory(nativeDirectory)
		}, gmeasure.SamplingConfig{N: 20})

		exp.SampleDuration("native-writing-with-throttling-and-low-latencty-network", func(_ int) {
			shapeTraffic(LOW_LATENCY_MILLIS)
			writeDataToMountedDirectory(nativeDirectory)
		}, gmeasure.SamplingConfig{N: 20})

		By("Starting MapFS process", func() {
			cmd := exec.Command(config.Driver, "-uid", "2000", "-gid", "2000", "-debug", mapfsDirectory, nativeDirectory)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("Mounted!"))
		})
		exp.SampleDuration("mapfs-writing-with-throttling-and-high-latencty-network", func(_ int) {
			shapeTraffic(HIGH_LATENCY_MILLIS)
			writeDataToMountedDirectory(mapfsDirectory)
		}, gmeasure.SamplingConfig{N: 20})
		exp.SampleDuration("mapfs-writing-with-throttling-and-low-latencty-network", func(_ int) {
			shapeTraffic(LOW_LATENCY_MILLIS)
			writeDataToMountedDirectory(mapfsDirectory)
		}, gmeasure.SamplingConfig{N: 20})

		Expect(exp.GetStats("mapfs-writing-with-throttling-and-low-latencty-network").FloatFor(gmeasure.StatMedian)).To(BeNumerically("<", exp.GetStats("native-writing-with-throttling-and-low-latencty-network").FloatFor(gmeasure.StatMedian)*10.0))
		Expect(exp.GetStats("mapfs-writing-with-throttling-and-high-latencty-network").FloatFor(gmeasure.StatMedian)).To(BeNumerically("<", exp.GetStats("native-writing-with-throttling-and-high-latencty-network").FloatFor(gmeasure.StatMedian)*10.0))
	})
})
