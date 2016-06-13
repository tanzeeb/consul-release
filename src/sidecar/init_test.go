package main_test

import (
	"testing"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSidecar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "sidecar")
}

var (
	pathToSidecar string
)

var _ = BeforeSuite(func() {
	var err error
	pathToSidecar, err = gexec.Build("github.com/cloudfoundry-incubator/consul-release/src/sidecar")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
