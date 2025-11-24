package request

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequest(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)

	fmt.Println("i'm taking here")
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

// func TestChunkRequestLineParse(t *testing.T) {
// 	// Test: Good GET Request line
// 	reader := &chunkReader{
// 		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
// 		numBytesPerRead: 3,
// 	}
// 	r, err := RequestFromReader(reader)
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "GET", r.RequestLine.Method)
// 	assert.Equal(t, "/", r.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

// 	// Test: Good GET Request line with path
// 	reader = &chunkReader{
// 		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
// 		numBytesPerRead: 1,
// 	}
// 	r, err = RequestFromReader(reader)
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "GET", r.RequestLine.Method)
// 	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
// }

// func TestHeaders(t *testing.T) {
// 	// Test: Standard Headers
// 	reader := &chunkReader{
// 		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
// 		numBytesPerRead: 3,
// 	}
// 	r, err := RequestFromReader(reader)
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	hostValue, _ := r.Headers.Get("host")
// 	userAgentValue, _ := r.Headers.Get("user-agent")
// 	acceptValue, _ := r.Headers.Get("accept")
// 	assert.Equal(t, "localhost:42069", hostValue)
// 	assert.Equal(t, "curl/7.81.0", userAgentValue)
// 	assert.Equal(t, "*/*", acceptValue)

// 	// Test: Malformed Header
// 	reader = &chunkReader{
// 		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
// 		numBytesPerRead: 3,
// 	}
// 	r, err = RequestFromReader(reader)
// 	require.Error(t, err)

// 	reader = &chunkReader{
// 		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nHost: localhost:8000\r\n\r\n",
// 		numBytesPerRead: 3,
// 	}
// 	r, err = RequestFromReader(reader)
// 	require.NoError(t, err)
// 	require.NotNil(t, r)

// 	hostValue, _ = r.Headers.Get("host")
// 	assert.Equal(t, "localhost:42069,localhost:8000", hostValue)
// }

// func TestRequestLineParse(t *testing.T) {
// 	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "GET", r.RequestLine.Method)
// 	assert.Equal(t, "/", r.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

// 	r, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "GET", r.RequestLine.Method)
// 	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

// 	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.Error(t, err)

// 	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1Host: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.Error(t, err)

// 	req, err := RequestFromReader(strings.NewReader("/coffee GET HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
// 	require.NoError(t, err)
// 	require.NotNil(t, req)
// 	assert.NotEqual(t, "GET", req.RequestLine.Method)
// 	assert.NotEqual(t, "/coffee", req.RequestLine.RequestTarget)
// 	assert.Equal(t, "1.1", req.RequestLine.HttpVersion)
// }
