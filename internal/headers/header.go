package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Set(n, value string) {
	h[n] = value
}

var CRLF = []byte("\r\n")

func (h *Headers) Parse(data []byte) (int, bool, error) {
	idx := bytes.Index(data, CRLF)

	if idx == -1 {
		return 0, false, nil
	}

	read := idx + len(CRLF)
	fieldLine := bytes.TrimSpace(data[:idx])
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)

	if len(parts) != 2 {
		return 0, false, nil
	}

	if bytes.HasSuffix(parts[0], []byte(" ")) {
		return 0, false, fmt.Errorf("malformed field name")
	}
	value := bytes.TrimSpace(parts[1])

	h.Set(string(parts[0]), string(value))

	return read, false, nil
}
