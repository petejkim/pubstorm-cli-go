package pathmatch

import "testing"

func TestPathMatch(t *testing.T) {
	testPathMatch := func(path, pattern string, match bool) {
		if v := PathMatch(path, pattern); v != match {
			if match {
				t.Errorf("expected PathMatch to find %s in %s", pattern, path)
			} else {
				t.Errorf("expected PathMatch not to find %s in %s", pattern, path)
			}
		}
	}
	testPathMatch("/foo/bar/123/.unison/foo/bar", ".unison", true)
	testPathMatch("/foo/bar/123/.unison", ".unison", true)
	testPathMatch(".unison/foo/bar/baz", ".unison", true)

	testPathMatch("/foo/bar/123/.unisonn/foo/bar", ".unison", false)
	testPathMatch("/foo/bar/123/.unisonn", ".unison", false)
	testPathMatch(".unisonn/foo/bar/baz", ".unison", false)

	testPathMatch("/foo/bar/123/..unison/foo/bar", ".unison", false)
	testPathMatch("/foo/bar/123/..unison", ".unison", false)
	testPathMatch("..unison/foo/bar/baz", ".unison", false)

	testPathMatch("/foo/bar/123/tmp/sessions/foo/bar", "tmp/sessions", true)
	testPathMatch("/foo/bar/123/tmp/sessions", "tmp/sessions", true)
	testPathMatch("tmp/sessions/foo/bar/baz", "tmp/sessions", true)

	testPathMatch("/foo/bar/123/tmp/sessionss/foo/bar", "tmp/sessions", false)
	testPathMatch("/foo/bar/123/tmp/sessionss", "tmp/sessions", false)
	testPathMatch("tmp/sessionss/foo/bar/baz", "tmp/sessions", false)

	testPathMatch("/foo/bar/123/ttmp/sessions/foo/bar", "tmp/sessions", false)
	testPathMatch("/foo/bar/123/ttmp/sessions", "tmp/sessions", false)
	testPathMatch("ttmp/sessions/foo/bar/baz", "tmp/sessions", false)
}

func TestPathMatchAny(t *testing.T) {
	testPathMatchAny := func(path string, patterns []string, match bool) {
		if v := PathMatchAny(path, patterns...); v != match {
			if match {
				t.Errorf("expected PathMatchAny to find %+v in %s", patterns, path)
			} else {
				t.Errorf("expected PathMatchAny not to find %+v in %s", patterns, path)
			}
		}
	}

	testPathMatchAny("/foo/bar/123/.unison/foo/bar", []string{".unison"}, true)
	testPathMatchAny("/foo/bar/123/.unison", []string{".unison", "foo/bar"}, true)
	testPathMatchAny(".unison/foo/bar/baz", []string{".unison", "bar"}, true)
	testPathMatchAny(".unison/foo/bar/baz", []string{".unison", "bar/baz"}, true)
	testPathMatchAny(".unison/foo/bar/baz", []string{".funison", "bar/baz"}, true)

	testPathMatchAny(".unison/foo/bar/baz", []string{".funison", "bar/baw"}, false)
}
