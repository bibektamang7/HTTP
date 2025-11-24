package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Set(name, value string) {

	name = strings.ToLower(name)
	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s,%s", v, value)
	} else {
		h.headers[name] = value
	}
}
func (h *Headers) Get(name string) (string, bool) {
	n := strings.ToLower(name)
	value, ok := h.headers[n]
	return value, ok
}

func (h *Headers) ForEach(cb func(name, value string)) {
	for n, v := range h.headers {
		cb(n, v)
	}
}

var ERROR_HEADER_PARSE = fmt.Errorf("malformed headers")
var CRLF = []byte("\r\n")

func isToken(token string) bool {
	if len(token) < 1 {
		return false
	}
	for _, ch := range token {
		found := false
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' {
			found = true
		}
		switch ch {
		case '#', '$', '%', '&', '\'', '*', '+', '-', '.', '_', '`', '|', '~':
			found = true
		}
		if !found {
			return false
		}
	}

	return true
}

func parseHeader(data []byte) (string, string, error) {
	parts := bytes.SplitN(data, []byte(":"), 2)

	if len(parts) != 2 {
		return "", "", ERROR_HEADER_PARSE
	}

	name := parts[0]
	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", ERROR_HEADER_PARSE
	}
	value := bytes.TrimSpace(parts[1])

	return string(name), string(value), nil

}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], CRLF)
		if idx == -1 {
			break
		}
		if idx == 0 {
			done = true
			read += len(CRLF)
			break
		}
		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isToken(name) {
			return 0, false, ERROR_HEADER_PARSE
		}
		read += idx + len(CRLF)

		h.Set(name, value)
	}
	return read, done, nil
}
