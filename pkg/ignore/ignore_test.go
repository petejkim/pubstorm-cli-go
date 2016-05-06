package ignore_test

import (
	"testing"

	"github.com/nitrous-io/rise-cli-go/pkg/ignore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ignore")
}

var _ = Describe("Ignore", func() {
	DescribeTable("Parse",
		func(input string, result []string) {
			Expect(ignore.Parse(input)).To(Equal(result))
		},

		Entry("filename", "config.ru", []string{"config.ru"}),
		Entry("filenames", "config.ru\nGemfile.lock", []string{"config.ru", "Gemfile.lock"}),
		Entry("filenames with carriage return", "config.ru\r\nGemfile.lock", []string{"config.ru", "Gemfile.lock"}),
		Entry("filename with #", " # config.ru\r\nGemfile.lock", []string{"Gemfile.lock"}),
		Entry("filenames empty lines", "\nGemfile.lock\n\n", []string{"Gemfile.lock"}),
		Entry("filenames with spaces", "\nGemfile.lock\ncon fig.ru\n", []string{"Gemfile.lock", "con fig.ru"}),
	)
})
