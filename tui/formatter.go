package tui

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"github.com/Sirupsen/logrus"
	"github.com/mattn/go-isatty"
)

var isTerminal = isatty.IsTerminal(os.Stdout.Fd())

type Formatter struct {
	// The fields are sorted by default for a consistent output. For applications
	// that log extremely frequently and don't use the JSON formatter this may not
	// be desired.
	DisableSorting bool
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var keys []string = make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}

	if !f.DisableSorting {
		sort.Strings(keys)
	}

	b := &bytes.Buffer{}

	prefixFieldClashes(entry.Data)

	if isTerminal {
		f.prettyPrint(b, entry, keys)
	} else {
		f.appendKeyValue(b, "level", entry.Level.String())
		if entry.Message != "" {
			f.appendKeyValue(b, "msg", entry.Message)
		}
		for _, key := range keys {
			f.appendKeyValue(b, key, entry.Data[key])
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *Formatter) prettyPrint(b *bytes.Buffer, entry *logrus.Entry, keys []string) {
	var levelText string

	switch entry.Level {
	case logrus.DebugLevel:
		levelText = Grn("[Debug]")
	case logrus.InfoLevel:
		levelText = Blu("[Info]")
	case logrus.WarnLevel:
		levelText = Ylo("[Warning]")
	case logrus.ErrorLevel:
		levelText = Red("[Error]")
	case logrus.FatalLevel:
		levelText = Red("[Fatal]")
	case logrus.PanicLevel:
		levelText = Red("[Panic]")
	default:
		levelText = ""
	}

	fmt.Fprintf(b, "%s %-44s ", levelText, entry.Message)

	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, " %s=%+v", Bold(k), v)
	}
}

func needsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.') {
			return false
		}
	}
	return true
}

func (f *Formatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {

	b.WriteString(key)
	b.WriteByte('=')

	switch value := value.(type) {
	case string:
		if needsQuoting(value) {
			b.WriteString(value)
		} else {
			fmt.Fprintf(b, "%q", value)
		}
	case error:
		errmsg := value.Error()
		if needsQuoting(errmsg) {
			b.WriteString(errmsg)
		} else {
			fmt.Fprintf(b, "%q", value)
		}
	default:
		fmt.Fprint(b, value)
	}

	b.WriteByte(' ')
}

func prefixFieldClashes(data logrus.Fields) {
	if _, ok := data["msg"]; ok {
		data["fields.msg"] = data["msg"]
	}

	if _, ok := data["level"]; ok {
		data["fields.level"] = data["level"]
	}
}
