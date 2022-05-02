package oc

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-log/log"
)

const errorBodyCaptureLimit = 512

func safeClose(log log.Logger, name string, c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Logf("failed to close %s: %s", name, err.Error())
	}
}

// ResponseError contains the response struct so that it can be
// inspected by the client.
type ResponseError struct {
	Response *http.Response
	Body     *bytes.Buffer
	message  string
}

func newResponseError(res *http.Response) error {
	var body bytes.Buffer

	_, _ = io.Copy(&body, res.Body)
	_ = res.Body.Close()

	return &ResponseError{
		Response: res,
		Body:     &body,
		message:  printableWithCap(body.Bytes(), errorBodyCaptureLimit),
	}
}

// Error formats an error message.
func (re *ResponseError) Error() string {
	return "server responded with: " + re.Response.Status + ": " + re.message
}

func printableWithCap(data []byte, max int) string {
	var b strings.Builder

	l := len(data)
	if l > max {
		l = max
	}

	b.Grow(l)

	var size int

	for len(data) > 0 {
		r, runeSize := utf8.DecodeRune(data)

		if r == utf8.RuneError || !(unicode.IsPrint(r) || unicode.IsSpace(r)) {
			return b.String()
		}

		if size+runeSize > max {
			return b.String()
		}

		_, _ = b.WriteRune(r)

		data = data[runeSize:]
		size += runeSize
	}

	return b.String()
}
