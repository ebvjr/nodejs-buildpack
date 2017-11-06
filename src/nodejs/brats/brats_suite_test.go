package brats_test

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"nodejs/brats/helper"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "256M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	return helper.InitBpData().Marshal()
}, func(data []byte) {
	// Run on all nodes
	helper.Data.Unmarshal(data)

	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(helper.Data.Cached, "_buildpack", "", 1))).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(helper.Data.Uncached, "_buildpack", "", 1))).To(Succeed())
	Expect(os.Remove(helper.Data.CachedFile)).To(Succeed())
})

func TestBrats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brats Suite")
}

func CopyBrats(nodejsVersion string) *cutlass.App {
	dir, err := cutlass.CopyFixture(filepath.Join(helper.Data.BpDir, "fixtures", "brats"))
	Expect(err).ToNot(HaveOccurred())

	if nodejsVersion != "" {
		file, err := ioutil.ReadFile(filepath.Join(dir, "package.json"))
		Expect(err).ToNot(HaveOccurred())
		obj := make(map[string]interface{})
		Expect(json.Unmarshal(file, &obj)).To(Succeed())
		engines, ok := obj["engines"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		engines["node"] = nodejsVersion
		file, err = json.Marshal(obj)
		Expect(err).ToNot(HaveOccurred())
		Expect(ioutil.WriteFile(filepath.Join(dir, "package.json"), file, 0644)).To(Succeed())
	}

	return cutlass.New(dir)
}
