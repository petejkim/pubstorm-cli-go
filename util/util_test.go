package util_test

import (
	"strings"
	"testing"

	"github.com/nitrous-io/rise-cli-go/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "util")
}

var _ = Describe("Util", func() {
	DescribeTable("ValidationErrorsToString",
		func(j map[string]interface{}, expected []string) {
			actual := util.ValidationErrorsToString(j)

			if expected == nil {
				Expect(actual).To(Equal(""))
			} else {
				actualStrs := strings.Split(actual, ", ")
				Expect(actualStrs).To(ConsistOf(expected))
			}
		},

		Entry("nil", nil, nil),
		Entry("empty map", map[string]interface{}{}, nil),

		Entry("map with empty errors", map[string]interface{}{
			"errors": map[string]interface{}{},
		}, nil),

		Entry("map with one item", map[string]interface{}{
			"errors": map[string]interface{}{
				"foo": "is not bar",
			},
		}, []string{
			"Foo is not bar",
		}),

		Entry("map with many items", map[string]interface{}{
			"errors": map[string]interface{}{
				"foo": "is not bar",
				"bar": "is not foo",
				"baz": "is not qux",
			},
		}, []string{
			"Foo is not bar",
			"Bar is not foo",
			"Baz is not qux",
		}),
	)

	DescribeTable("Capitalize",
		func(str, expected string) {
			Expect(util.Capitalize(str)).To(Equal(expected))
		},

		Entry("empty string", "", ""),
		Entry("capitalize one word", "hello", "Hello"),
		Entry("capitalize only the first word", "hello world", "Hello world"),
		Entry("capitalize the first word, not touching the rest", "foo Bar baz", "Foo Bar baz"),
		Entry("unicode shouldn't break", "유니코드", "유니코드"),
	)
})
