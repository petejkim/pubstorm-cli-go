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
				actualStrs := strings.Split(actual, "\n")
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
			"* foo is not bar",
		}),

		Entry("map with many items", map[string]interface{}{
			"errors": map[string]interface{}{
				"foo": "is not bar",
				"bar": "is not foo",
				"baz": "is not qux",
			},
		}, []string{
			"* foo is not bar",
			"* bar is not foo",
			"* baz is not qux",
		}),
	)
})
