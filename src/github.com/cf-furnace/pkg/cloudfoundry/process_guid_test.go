package cloudfoundry_test

import (
	"encoding/base32"
	"fmt"

	"k8s.io/kubernetes/pkg/util/validation"

	"github.com/cf-furnace/pkg/cloudfoundry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProcessGuid", func() {
	var processGuid cloudfoundry.ProcessGuid
	var appGuid, appVersion string

	BeforeEach(func() {
		appGuid = "8d58c09b-b305-4f16-bcfe-b78edcb77100"
		appVersion = "3f258eb0-9dac-460c-a424-b43fe92bee27"
	})

	Describe("NewProcessGuid", func() {
		var err error

		JustBeforeEach(func() {
			processGuid, err = cloudfoundry.NewProcessGuid(fmt.Sprintf("%s-%s", appGuid, appVersion))
		})

		It("breaks the guid apart", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(processGuid.AppGuid.String()).To(Equal(appGuid))
			Expect(processGuid.AppVersion.String()).To(Equal(appVersion))
		})

		Context("when the source guid is too short", func() {
			BeforeEach(func() {
				appGuid = "guid"
				appVersion = "version"
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("invalid process guid"))
			})
		})

		Context("when the source app guid is invalid", func() {
			BeforeEach(func() {
				appGuid = "this-is-an-invalid-app-guid-with-len"
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("Invalid UUID string"))
			})
		})

		Context("when the source app version is invalid", func() {
			BeforeEach(func() {
				appVersion = "this-is-an-invalid-app-version-with-"
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("Invalid UUID string"))
			})
		})
	})

	Describe("ShortenedGuid", func() {
		BeforeEach(func() {
			pg, err := cloudfoundry.NewProcessGuid(fmt.Sprintf("%s-%s", appGuid, appVersion))
			Expect(err).NotTo(HaveOccurred())
			processGuid = pg
		})

		It("returns an encoded version of the process guid", func() {
			Expect(processGuid.ShortenedGuid()).To(Equal("rvmmbg5tavhrnph6w6hnzn3raa-h4sy5me5vrdazjbewq76sk7oe4"))
		})

		It("is a valid kubernetes dns label", func() {
			Expect(validation.IsDNS1123Label(processGuid.ShortenedGuid())).To(BeEmpty())
		})
	})

	Describe("String", func() {
		BeforeEach(func() {
			pg, err := cloudfoundry.NewProcessGuid(fmt.Sprintf("%s-%s", appGuid, appVersion))
			Expect(err).NotTo(HaveOccurred())
			processGuid = pg
		})

		It("returns the normal representation of a process guid", func() {
			Expect(processGuid.String()).To(Equal(appGuid + "-" + appVersion))
		})
	})

	Describe("DecodeProcessGuid", func() {
		var expectedGuid cloudfoundry.ProcessGuid
		var actualShortenedGuid string
		var actualGuid cloudfoundry.ProcessGuid
		var actualErr error

		BeforeEach(func() {
			var err error
			expectedGuid, err = cloudfoundry.NewProcessGuid(fmt.Sprintf("%s-%s", appGuid, appVersion))
			Expect(err).NotTo(HaveOccurred())

			actualShortenedGuid = "rvmmbg5tavhrnph6w6hnzn3raa-h4sy5me5vrdazjbewq76sk7oe4"
		})

		JustBeforeEach(func() {
			actualGuid, actualErr = cloudfoundry.DecodeProcessGuid(actualShortenedGuid)
		})

		It("decodes a shortened GUID", func() {
			Expect(actualErr).NotTo(HaveOccurred())
			Expect(actualGuid).To(Equal(expectedGuid))
		})

		Context("with an invalid shortened guid", func() {
			BeforeEach(func() {
				actualShortenedGuid = "garbage"
			})

			It("returns an error", func() {
				Expect(actualErr).To(MatchError("invalid-shortened-guid"))
			})
		})

		Context("with non-base32 app guid", func() {
			BeforeEach(func() {
				actualShortenedGuid = "can-h4sy5me5vrdazjbewq76sk7oe4"
			})

			It("returns an error", func() {
				Expect(actualErr).To(BeAssignableToTypeOf(base32.CorruptInputError(0)))
			})
		})

		Context("with base32 app non-guid", func() {
			BeforeEach(func() {
				actualShortenedGuid = "mfrgg-h4sy5me5vrdazjbewq76sk7oe4"
			})

			It("returns an error", func() {
				Expect(actualErr.Error()).To(MatchRegexp(".* not valid UUID sequence"))
			})
		})

		Context("with non-base32 task id", func() {
			BeforeEach(func() {
				actualShortenedGuid = "rvmmbg5tavhrnph6w6hnzn3raa-can"
			})

			It("returns an error", func() {
				Expect(actualErr).To(BeAssignableToTypeOf(base32.CorruptInputError(0)))
			})
		})

		Context("with base32 task non-guid", func() {
			BeforeEach(func() {
				actualShortenedGuid = "rvmmbg5tavhrnph6w6hnzn3raa-mfrgg"
			})

			It("returns an error", func() {
				Expect(actualErr.Error()).To(MatchRegexp(".* not valid UUID sequence"))
			})
		})
	})
})
