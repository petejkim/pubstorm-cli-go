package readline_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/nitrous-io/rise-cli-go/pkg/readline"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "readline")
}

var _ = Describe("Readline", func() {
	var (
		origInput  io.Reader
		origOutput io.Writer

		input  *bytes.Buffer
		output *bytes.Buffer
	)

	BeforeEach(func() {
		input = &bytes.Buffer{}
		origInput = readline.Input
		readline.Input = input

		output = &bytes.Buffer{}
		origOutput = readline.Output
		readline.Output = output
	})

	AfterEach(func() {
		readline.Input = origInput
		readline.Output = origOutput
	})

	writeToInput := func(b []byte) {
		_, err := input.Write(b)
		Expect(err).To(BeNil())
	}

	DescribeTable("Read",
		func(input []byte, retry bool, def, expected string) {
			writeToInput(input)
			result, err := readline.Read("Enter: ", retry, def)
			Expect(err).To(BeNil())

			Expect(result).To(Equal(expected))
		},
		Entry("regular string", []byte("Hello world\n"), false, "", "Hello world"),
		Entry("string that ends with extra CRLF", []byte("Hello world\r\n"), false, "", "Hello world"),
		Entry("empty string followed by new line with no retry", []byte("\nHello world\n"), false, "", ""),
		Entry("empty string followed by new line with retry", []byte("\nHello world\n"), true, "", "Hello world"),
		Entry("empty string with default", []byte("\n"), false, "Hello world", "Hello world"),
	)

	DescribeTable("ReadSecurely",
		func(input []byte, retry bool, def, expected string, expectedErr error) {
			writeToInput(input)
			result, err := readline.ReadSecurely("Enter: ", retry, def)
			if expectedErr != nil {
				Expect(err).To(Equal(expectedErr))
			} else {
				Expect(err).To(BeNil())
			}

			Expect(result).To(Equal(expected))
		},
		Entry("regular string", []byte("Hello world\n"), false, "", "Hello world", nil),
		Entry("string that ends with extra CRLF", []byte("Hello world\r\n"), false, "", "Hello world", nil),
		Entry("empty string followed by new line with no retry", []byte("\nHello world\n"), false, "", "", nil),
		Entry("empty string followed by new line with retry", []byte("\nHello world\n"), true, "", "Hello world", nil),
		Entry("empty string with default", []byte("\n"), false, "Hello world", "Hello world", nil),
		Entry("interrupted by ^C", []byte{'a', 'b', 3 /* ^C */, '\n'}, false, "", "", io.EOF),
		Entry("interrupted by ^D", []byte{'a', 'b', 4 /* ^D */, '\n'}, false, "", "", io.EOF),
		Entry("string followed by BS char", []byte{'a', 'b', 8 /* BS */, 'c', '\n'}, false, "", "ac", nil),
		Entry("string followed by DEL char", []byte{'a', 'b', 127 /* DEL */, 'c', '\n'}, false, "", "ac", nil),
		Entry("non printable chars", []byte{'a', 'b', 128, 'c', 130, 'd', 31, 'e', '\n'}, false, "", "abcde", nil),
	)
})
