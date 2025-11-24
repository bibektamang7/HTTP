package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/bibektamang7/httpFromScratch/internal/headers"
)

type parseState string

const (
	StateInit    parseState = "init"
	StateHeaders parseState = "headers"
	StateBody    parseState = "body"
	StateDone    parseState = "done"
	StateError   parseState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func getHeaderInt(headers *headers.Headers, name string, defaultValue int) int {
	value, ok := headers.Get(name)
	if !ok {
		return defaultValue
	}

	val, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return val
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
	Body        []byte
}

func (r *Request) done() bool { return r.state == StateDone || r.state == StateError }
func (r *Request) hasBody() bool {
	length := getHeaderInt(&r.Headers, "content-length", 0)
	return length > 0
}

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
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.state {
		case StateError:
			return 0, ERROR_PARSE_ERROR
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
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
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}

			read += n
			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}
		case StateBody:
			length := getHeaderInt(&r.Headers, "content-length", 0)
			if length == 0 {
				panic("body chunked not implemented")
			}
			remaining := min(length-len(r.Body), len(currentData))

			read += remaining
			r.Body = append(r.Body, currentData[:remaining]...)
			if length == len(r.Body) {
				r.state = StateDone
			}
			break outer
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

	//TODO: EDGE CASES
	/*
		content-length is missing, but body is present,
		Length of content-length and body is differnt
	*/
	// length := getHeaderInt(&request.Headers, "content-length", 0)
	// fmt.Println("this len of body", len(request.Body))
	// if length != len(request.Body) {
	// 	return request, fmt.Errorf("mismatched content-length and body length")
	// }
	return request, nil
}
