package cloudfoundry_test

import (
	"encoding/base32"
	"encoding/hex"
	"fmt"

	"k8s.io/kubernetes/pkg/util/validation"

	"github.com/cf-furnace/pkg/cloudfoundry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TaskGuid", func() {
	var taskGuid cloudfoundry.TaskGuid
	var appGuid, taskId string

	BeforeEach(func() {
		appGuid = "7fe5f9a4-8b38-4f54-931b-df91723c94d7"
		taskId = "dc68c40dce3f468d8ea2398f785efe39"
	})
	Describe("NewTaskGuid", func() {
		var err error

		JustBeforeEach(func() {
			taskGuid, err = cloudfoundry.NewTaskGuid(fmt.Sprintf("%s-%s", appGuid, taskId))
		})

		It("breaks the guid apart", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(taskGuid.AppGuid.String()).To(Equal(appGuid))
			Expect(hex.EncodeToString(taskGuid.TaskId)).To(Equal(taskId))
		})

		Context("when the source guid is too short", func() {
			BeforeEach(func() {
				appGuid = "guid"
				taskId = "version"
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("invalid task guid"))
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
	})

	Describe("ShortenedGuid", func() {
		BeforeEach(func() {
			pg, err := cloudfoundry.NewTaskGuid(fmt.Sprintf("%s-%s", appGuid, taskId))
			Expect(err).NotTo(HaveOccurred())
			taskGuid = pg
		})

		It("returns an encoded version of the task guid", func() {
			Expect(taskGuid.ShortenedGuid()).To(Equal("p7s7tjelhbhvjey336ixepeu24-3rumidooh5di3dvchghxqxx6he"))
		})

		It("is a valid kubernetes dns label", func() {
			Expect(validation.IsDNS1123Label(taskGuid.ShortenedGuid())).To(BeEmpty())
		})
	})

	Describe("String", func() {
		BeforeEach(func() {
			pg, err := cloudfoundry.NewTaskGuid(fmt.Sprintf("%s-%s", appGuid, taskId))
			Expect(err).NotTo(HaveOccurred())
			taskGuid = pg
		})

		It("returns the normal representation of a task guid", func() {
			Expect(taskGuid.String()).To(Equal(appGuid + "-" + taskId))
		})
	})

	Describe("DecodeTaskGuid", func() {
		var expectedGuid cloudfoundry.TaskGuid
		var actualShortenedGuid string
		var actualGuid cloudfoundry.TaskGuid
		var actualErr error

		BeforeEach(func() {
			var err error
			expectedGuid, err = cloudfoundry.NewTaskGuid(fmt.Sprintf("%s-%s", appGuid, taskId))
			Expect(err).NotTo(HaveOccurred())

			actualShortenedGuid = "p7s7tjelhbhvjey336ixepeu24-3rumidooh5di3dvchghxqxx6he"
		})

		JustBeforeEach(func() {
			actualGuid, actualErr = cloudfoundry.DecodeTaskGuid(actualShortenedGuid)
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
				actualShortenedGuid = "h4sy5me5vrdazjbewq76sk7oe4-can"
			})

			It("returns an error", func() {
				Expect(actualErr).To(BeAssignableToTypeOf(base32.CorruptInputError(0)))
			})
		})
	})
})
