package port_pool_test

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/cf-furnace/pkg/port_pool"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("State", func() {
	var (
		tmpDir   string
		filePath string
	)

	BeforeEach(func() {
		var err error

		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		filePath = path.Join(tmpDir, "port_pool.json")
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Describe("NewState", func() {
		It("should parse the provided file", func() {
			Expect(ioutil.WriteFile(filePath, []byte(`{
				"offset": 10
			}`), 0660)).To(Succeed())

			portPoolState, err := port_pool.LoadState(filePath)
			Expect(err).NotTo(HaveOccurred())

			Expect(portPoolState.Offset).To(BeNumerically("==", 10))
		})

		Context("when the file does not exist", func() {
			It("should return a wrapped error", func() {
				_, err := port_pool.LoadState("/path/to/not/existing/banana")
				Expect(err).To(MatchError(ContainSubstring("openning state file")))
			})
		})

		Context("when the file is invalid", func() {
			It("should return a wrapped error", func() {
				Expect(ioutil.WriteFile(filePath, []byte(`{
				"offset": `), 0660)).To(Succeed())

				_, err := port_pool.LoadState(filePath)
				Expect(err).To(MatchError(ContainSubstring("parsing state file")))
			})
		})
	})

	Describe("Save", func() {
		It("should write the file", func() {
			Expect(ioutil.WriteFile(filePath, []byte("{}"), 0660)).To(Succeed())
			state := port_pool.State{
				Offset: 10,
			}

			Expect(port_pool.SaveState(filePath, state)).To(Succeed())

			contents, err := ioutil.ReadFile(filePath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(contents)).To(ContainSubstring("\"offset\":10"))
		})

		Context("when file can not be created", func() {
			It("should return a sensible error", func() {
				state := port_pool.State{
					Offset: 10,
				}

				err := port_pool.SaveState("/path/to/my/basement/", state)
				Expect(err).To(MatchError(ContainSubstring("creating state file")))
			})
		})
	})
})
