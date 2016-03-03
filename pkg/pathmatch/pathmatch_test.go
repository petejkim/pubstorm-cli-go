package pathmatch_test

import (
	"testing"

	"github.com/nitrous-io/rise-cli-go/pkg/pathmatch"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pathmatch")
}

var _ = Describe("Pathmatch", func() {
	DescribeTable("PathMatch",
		func(path, pattern string, match bool) {
			v := pathmatch.PathMatch(path, pattern)
			Expect(v).To(Equal(match))
		},

		Entry("matches", "/foo/bar/123/.unison/foo/bar", ".unison", true),
		Entry("matches", "/foo/bar/123/.unison", ".unison", true),
		Entry("matches", ".unison/foo/bar/baz", ".unison", true),

		Entry("matches", "/foo/bar/123/tmp/sessions/foo/bar", "tmp/sessions", true),
		Entry("matches", "/foo/bar/123/tmp/sessions", "tmp/sessions", true),
		Entry("matches", "tmp/sessions/foo/bar/baz", "tmp/sessions", true),

		Entry("does not match", "/foo/bar/123/.unisonn/foo/bar", ".unison", false),
		Entry("does not match", "/foo/bar/123/.unisonn", ".unison", false),
		Entry("does not match", ".unisonn/foo/bar/baz", ".unison", false),

		Entry("does not match", "/foo/bar/123/..unison/foo/bar", ".unison", false),
		Entry("does not match", "/foo/bar/123/..unison", ".unison", false),
		Entry("does not match", "..unison/foo/bar/baz", ".unison", false),

		Entry("does not match", "/foo/bar/123/tmp/sessionss/foo/bar", "tmp/sessions", false),
		Entry("does not match", "/foo/bar/123/tmp/sessionss", "tmp/sessions", false),
		Entry("does not match", "tmp/sessionss/foo/bar/baz", "tmp/sessions", false),

		Entry("does not match", "/foo/bar/123/ttmp/sessions/foo/bar", "tmp/sessions", false),
		Entry("does not match", "/foo/bar/123/ttmp/sessions", "tmp/sessions", false),
		Entry("does not match", "ttmp/sessions/foo/bar/baz", "tmp/sessions", false),
	)

	DescribeTable("PathMatchAny",
		func(path string, patterns []string, match bool) {
			v := pathmatch.PathMatchAny(path, patterns...)
			Expect(v).To(Equal(match))
		},

		Entry("matches", "/foo/bar/123/.unison/foo/bar", []string{".unison"}, true),
		Entry("matches", "/foo/bar/123/.unison", []string{".unison", "foo/bar"}, true),
		Entry("matches", ".unison/foo/bar/baz", []string{".unison", "bar"}, true),
		Entry("matches", ".unison/foo/bar/baz", []string{".unison", "bar/baz"}, true),
		Entry("matches", ".unison/foo/bar/baz", []string{".funison", "bar/baz"}, true),

		Entry("does not match", ".unison/foo/bar/baz", []string{".funison", "bar/baw"}, false),
	)
})
