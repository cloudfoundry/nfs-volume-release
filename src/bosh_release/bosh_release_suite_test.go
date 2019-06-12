package bosh_release_test

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"os"
	"os/exec"
	"time"

	"testing"
)

func TestBoshReleaseTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BoshReleaseTest Suite")
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Minute)

	if !hasStemcell() {
		uploadStemcell()
	}

	deploy()
})

func deploy(opsfiles ...string) {
	deployCmd := []string {"deploy",
		"-n",
		"-d",
		"bosh_release_test",
		"./nfsv3driver-manifest.yml",
		"-v", fmt.Sprintf("path_to_nfs_volume_release=%s", os.Getenv("NFS_VOLUME_RELEASE_PATH")),
		"-v", fmt.Sprintf("path_to_mapfs_release=%s", os.Getenv("MAPFS_RELEASE_PATH")),
	}

	updatedDeployCmd := make([]string, len(deployCmd))
	copy(updatedDeployCmd, deployCmd)
	for _, optFile := range opsfiles {
		updatedDeployCmd = append(updatedDeployCmd, "-o", optFile)
	}

	boshDeployCmd := exec.Command("bosh", updatedDeployCmd...)
	session, err := gexec.Start(boshDeployCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 60*time.Minute).Should(gexec.Exit(0))
}

func hasStemcell() bool {
	boshStemcellsCmd := exec.Command("bosh", "stemcells", "--json")
	stemcellOutput := gbytes.NewBuffer()
	session, err := gexec.Start(boshStemcellsCmd, stemcellOutput, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 1*time.Minute).Should(gexec.Exit(0))
	boshStemcellsOutput := &BoshStemcellsOutput{}
	Expect(json.Unmarshal(stemcellOutput.Contents(), boshStemcellsOutput)).Should(Succeed())
	return len(boshStemcellsOutput.Tables[0].Rows) > 0
}

func uploadStemcell() {
	boshUsCmd := exec.Command("bosh", "upload-stemcell", "https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-xenial-go_agent")
	session, err := gexec.Start(boshUsCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 20*time.Minute).Should(gexec.Exit(0))
}


func findProcessState(processName string) string {

	boshIsCmd := exec.Command("bosh", "instances", "--ps", "--details", "--json", "--column=process", "--column=process_state")

	boshInstancesOutput := gbytes.NewBuffer()
	session, err := gexec.Start(boshIsCmd, boshInstancesOutput, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	session = session.Wait(1 * time.Minute)
	Expect(session.ExitCode()).To(Equal(0), string(boshInstancesOutput.Contents()))

	instancesOutputJson := &BoshInstancesOutput{}
	err = json.Unmarshal(boshInstancesOutput.Contents(), instancesOutputJson)
	Expect(err).NotTo(HaveOccurred())

	for _, row := range instancesOutputJson.Tables[0].Rows {
		if row.Process == processName {
			return row.ProcessState
		}
	}

	return ""
}

type BoshInstancesOutput struct {
	Tables []struct {
		Content string `json:"Content"`
		Header  struct {
			Process      string `json:"process"`
			ProcessState string `json:"process_state"`
		} `json:"Header"`
		Rows []struct {
			Process      string `json:"process"`
			ProcessState string `json:"process_state"`
		} `json:"Rows"`
		Notes interface{} `json:"Notes"`
	} `json:"Tables"`
	Blocks interface{} `json:"Blocks"`
	Lines  []string    `json:"Lines"`
}


type BoshStemcellsOutput struct {
	Tables []struct {
		Content string `json:"Content"`
		Header  struct {
			Cid     string `json:"cid"`
			Cpi     string `json:"cpi"`
			Name    string `json:"name"`
			Os      string `json:"os"`
			Version string `json:"version"`
		} `json:"Header"`
		Rows []struct {
			Cid     string `json:"cid"`
			Cpi     string `json:"cpi"`
			Name    string `json:"name"`
			Os      string `json:"os"`
			Version string `json:"version"`
		} `json:"Rows"`
		Notes []string `json:"Notes"`
	} `json:"Tables"`
	Blocks interface{} `json:"Blocks"`
	Lines  []string    `json:"Lines"`
}