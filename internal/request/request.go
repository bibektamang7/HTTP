package request

import (
	"bytes"
	"fmt"
	"io"

	"github.com/bibektamang7/httpFromScratch/internal/headers"
)

type parseState string

const (
	StateInit    parseState = "init"
	StateHeaders parseState = "headers"
	StateDone    parseState = "done"
	StateError   parseState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func NewRequestLine(version, target, method string) *RequestLine {
	return &RequestLine{
		HttpVersion:   version,
		RequestTarget: target,
		Method:        method,
	}
}

type Request struct {
	RequestLine RequestLine
	state       parseState
	Headers     headers.Headers
}

func (r *Request) done() bool { return r.state == StateDone || r.state == StateError }

func NewRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: *headers.NewHeaders(),
	}
}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERROR_PARSE_ERROR = fmt.Errorf("failed to parse chuck request")
var ERROR_PARSE_HEADERS = fmt.Errorf("failed to request headers")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}
	startLine := b[:idx]
	read := idx + len(SEPARATOR)
	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	httpVersion := bytes.Split(parts[2], []byte("/"))

	if len(httpVersion) != 2 || string(httpVersion[0]) != "HTTP" || string(httpVersion[1]) != "1.1" {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	rl := NewRequestLine(string(httpVersion[1]), string(parts[1]), string(parts[0]))

	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		switch r.state {
		case StateError:
			return 0, ERROR_PARSE_ERROR
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}

			read += n
			if done {
				r.state = StateDone
			}
		case StateDone:
			break outer
		default:
			panic("Programmed Poorly")
		}
	}
	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := NewRequest()

	data := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		read, err := reader.Read(data[bufLen:])
		if err != nil {
			fmt.Println("Failed to read from Request")
			return nil, err
		}
		bufLen += read
		n, err := request.parse(data[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(data, data[n:bufLen])
		bufLen -= n
	}

	return request, nil
}
