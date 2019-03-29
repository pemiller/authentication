package datastore

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/couchbase/gocb"
)

// DocTypeTranscoder ...
type DocTypeTranscoder struct {
	DefaultTranscoder gocb.Transcoder
}

var typeName = []byte(`,"__type":"`)
var timestampName = []byte(`,"__ts":`)
var endString = []byte(`"`)
var endClass = []byte(`}`)

// Decode sets transcoding behavior
func (t DocTypeTranscoder) Decode(content []byte, formatFlags uint32, out interface{}) error {
	return t.DefaultTranscoder.Decode(content, formatFlags, out)
}

// Encode sets transcoding behavior
func (t DocTypeTranscoder) Encode(value interface{}) ([]byte, uint32, error) {
	enc, flags, err := t.DefaultTranscoder.Encode(value)
	if err != nil {
		return enc, flags, err
	}
	ok, dt := getTypeName(value)
	if !ok {
		return enc, flags, err
	}
	if enc[0] == 123 {
		var buf bytes.Buffer
		buf.Write(enc[:len(enc)-1])
		buf.Write(typeName)
		buf.WriteString(dt)
		buf.Write(endString)
		buf.Write(timestampName)
		buf.WriteString(strconv.FormatInt(time.Now().UTC().UnixNano(), 10))
		buf.Write(endClass)
		return buf.Bytes(), flags, err
	}
	return enc, flags, err
}

func getTypeName(value interface{}) (bool, string) {
	t := reflect.TypeOf(value)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct, camelCase(t.Name())
}

func camelCase(value string) string {
	// don't split invalid utf8
	if !utf8.ValidString(value) {
		return value
	}
	var runes [][]rune
	lastClass := 0
	var class int
	// split into fields based on class of unicode character
	for _, r := range value {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	return combine(runes)
}

func combine(runes [][]rune) string {
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	entries := []string{}
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return strings.ToLower(strings.Join(entries, "_"))
}
