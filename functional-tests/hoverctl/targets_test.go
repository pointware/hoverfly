package hoverctl_suite

import (
	"github.com/SpectoLabs/hoverfly/functional-tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("When using the `targets` command", func() {

	Context("viewing targets", func() {
		Context("with no targets", func() {

			It("should fail nicely", func() {
				output := functional_tests.Run(hoverctlBinary, "targets")

				Expect(output).To(ContainSubstring("No targets registered"))
			})
		})

		Context("with targets", func() {

			BeforeEach(func() {
				functional_tests.Run(hoverctlBinary, "targets", "create", "--target", "default", "--admin-port", "1234", "--proxy-port", "8765", "--host", "localhost")
			})

			AfterEach(func() {
				functional_tests.Run(hoverctlBinary, "targets", "delete", "--target", "default")
			})

			It("print targets", func() {
				output := functional_tests.Run(hoverctlBinary, "targets")

				Expect(output).To(ContainSubstring("TARGET NAME"))
				Expect(output).To(ContainSubstring("HOST"))
				Expect(output).To(ContainSubstring("ADMIN PORT"))
				Expect(output).To(ContainSubstring("PROXY PORT"))

				Expect(output).To(ContainSubstring("default"))
				Expect(output).To(ContainSubstring("localhost"))
				Expect(output).To(ContainSubstring("1234"))
				Expect(output).To(ContainSubstring("8765"))
			})
		})
	})

	Context("creating targets", func() {

		It("should create the target and print it", func() {

			output := functional_tests.Run(hoverctlBinary, "targets", "create", "--target", "new-target", "--admin-port", "1234", "--proxy-port", "8765", "--host", "localhost")

			Expect(output).To(ContainSubstring("TARGET NAME"))
			Expect(output).To(ContainSubstring("HOST"))
			Expect(output).To(ContainSubstring("ADMIN PORT"))
			Expect(output).To(ContainSubstring("PROXY PORT"))

			Expect(output).To(ContainSubstring("new-target"))
			Expect(output).To(ContainSubstring("localhost"))
			Expect(output).To(ContainSubstring("1234"))
			Expect(output).To(ContainSubstring("8765"))
		})

		It("should create a default if no target name is provided", func() {
			output := functional_tests.Run(hoverctlBinary, "targets", "create")

			Expect(output).To(ContainSubstring("TARGET NAME"))
			Expect(output).To(ContainSubstring("HOST"))
			Expect(output).To(ContainSubstring("ADMIN PORT"))

			Expect(output).To(ContainSubstring("default"))
			Expect(output).To(ContainSubstring("localhost"))
			Expect(output).To(ContainSubstring("8888"))
		})
	})

	Context("deleting targets", func() {

		BeforeEach(func() {
			functional_tests.Run(hoverctlBinary, "targets", "create", "--target", "default", "--admin-port", "1234")
		})

		AfterEach(func() {
			functional_tests.Run(hoverctlBinary, "targets", "delete", "--target", "default")
		})

		It("should delete targets and print nice empty message", func() {
			output := functional_tests.Run(hoverctlBinary, "targets", "delete", "--target", "default", "--force")

			Expect(output).To(ContainSubstring("No targets registered"))
		})

		It("should fail nicely if no target name is provided", func() {
			output := functional_tests.Run(hoverctlBinary, "targets", "delete")

			Expect(output).To(ContainSubstring("Cannot delete a target without a name"))
		})
	})

})