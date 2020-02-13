// package uncurl is Go library to consume a Chrome/Chromium browser "Copy as cURL" string and generate one or more Go *http.Request objects from it
//
// In the Chrome or Chromium browser, if you open "Developer tools" and go to the Network tab and
// navigate somewhere with the browser, you will see a list of requests. Right-clicking one these
// requests yields a menu with a Copy submenu. One of those options is "Copy as cURL". It allows you to
// paste a `curl` command to your terminal or editor, one that reproduces the request if run.
//
// This library accepts that text input and turns it into a Go request. Further Go requests can be
// generated with different target URLs while maintaining the same header values.
package uncurl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

const (
	// these patterns match output from Chrome/Chromium
	curlHeaderPattern = `-H\s+'([^:]+?):\s+(.+?)'`
	curlTargetPattern = `^\s*curl\s+'([^']+?)' `
	curlDataPattern   = ` --data '([^']+?)' `

	curlAcceptEncodingPattern = `(?i)^\s*accept-encoding\s*$`
)

var curlHeaderRe, curlTargetRe, curlDataRe, curlAcceptEncodingRe *regexp.Regexp

func init() {
	curlHeaderRe = regexp.MustCompile(curlHeaderPattern)
	curlTargetRe = regexp.MustCompile(curlTargetPattern)
	curlDataRe = regexp.MustCompile(curlDataPattern)
	curlAcceptEncodingRe = regexp.MustCompile(curlAcceptEncodingPattern)
}

// Uncurl is the object from which requests are generated. Create one with NewUncurl
type Uncurl struct {
	// input is the original curl string
	input  []byte
	header http.Header

	// target is the original URL target
	target string

	// method is the original HTTP Method
	method string

	// body is the original body
	body []byte

	// AcceptEncoding is the original `accept-encoding` header value. Including this header on our Go
	// request would signal to the `net/http` package that we do not wish to use DefaultTransport for
	// our request, disabling automatic gzip handling. As that's not usually desired, the value is
	// instead copied here for the user to employ as desired.
	AcceptEncoding string
}

// NewUncurl generates a new Uncurl object from a Chrome/Chromium "Copy as cURL" input as bytes.
// This is useful when you're loading from a file or concerned about efficiency. If you prefer to pass
// string input instead, use NewUncurlString.
func NewUncurl(b []byte) (*Uncurl, error) {
	if b == nil || len(b) == 0 {
		return nil, errors.New("NewUncurl called with empty parameter")
	}
	un := new(Uncurl)
	un.input = b
	un.method = `GET`
	cm := curlTargetRe.FindSubmatch(b)
	if len(cm) < 2 {
		return nil, fmt.Errorf("Failed to find target URL in curl string %s", b)
	}
	un.target = string(cm[1])
	if _, err := url.ParseRequestURI(un.target); err != nil {
		return nil, fmt.Errorf("Target url %s failed to parse: %s", un.target, err)
	}
	h := make(http.Header)
	all := curlHeaderRe.FindAllSubmatch(b, -1)
	for _, m := range all {
		if m[1] == nil {
			continue
		}
		if curlAcceptEncodingRe.Match(m[1]) { // use default Transport
			un.AcceptEncoding = string(m[2])
			continue
		}
		h[string(m[1])] = []string{string(m[2])}
	}
	un.header = h
	dm := curlDataRe.FindSubmatch(b)
	if len(dm) == 2 {
		un.method = `POST`
		un.body = dm[1]
	}
	_, err := http.NewRequest(un.method, un.target, un.bodyReadCloser())
	if err != nil {
		return nil, fmt.Errorf("Unable to create new request from curl: %s", err)
	}
	return un, nil
}

// NewUncurlString generates a new Uncurl object from a Chrome/Chromium "Copy as cURL" string
func NewUncurlString(s string) (*Uncurl, error) {
	return NewUncurl([]byte(s))
}

func (un *Uncurl) bodyReadCloser() io.ReadCloser {
	var bodyBuf io.ReadCloser
	if un.body != nil {
		bodyBuf = ioutil.NopCloser(bytes.NewBuffer(un.body))
	}
	return bodyBuf
}

// Header creates a new http.Header map and copies all headers from the original curl, with the
// exception of Accept-Encoding, to it
func (un *Uncurl) Header() http.Header {
	h := make(http.Header)
	for k, v := range un.header {
		s := make([]string, len(v))
		copy(s, v)
		h[k] = s
	}
	return h
}

// String satisfies the `fmt.Stringer` interface by returning the original curl string
func (un *Uncurl) String() string {
	return string(un.input)
}

// Target returns the URL from the original curl string
func (un *Uncurl) Target() string {
	return un.target
}

// Method returns the HTTP method string from the original curl string
func (un *Uncurl) Method() string {
	return un.method
}

// Body returns a copy of the --data argument from the original curl string. The slice will be empty if
// --data was not present.
func (un *Uncurl) Body() []byte {
	b := make([]byte, len(un.body))
	copy(b, un.body)
	return b
}

// Request returns the Go `*http.Request` version of the curl
func (un *Uncurl) Request() *http.Request {
	r, _ := un.NewRequest(un.method, un.target, un.bodyReadCloser()) // as all relevant variables are private, we can rely on the error check done in NewUncurl
	r.Header = un.Header()
	r.GetBody = func() (io.ReadCloser, error) {
		return un.bodyReadCloser(), nil
	}
	return r
}

// NewRequest is like Request(), but allows the caller to set the method, url, and body; matching the
// function signature of http.NewRequest
func (un *Uncurl) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, body) // as all relevant variables are private, we can rely on the error check done in NewUncurl
	if err != nil {
		return nil, fmt.Errorf("Error building request: %s", err)
	}
	r.Header = un.Header()
	return r, nil
}

// NewRequestWithContext is like NewRequest but allows setting of context as well, matching the
// signature of http.NewRequestWithContext
func (un *Uncurl) NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, method, url, body) // as all relevant variables are private, we can rely on the error check done in NewUncurl
	if err != nil {
		return nil, fmt.Errorf("Error building request: %s", err)
	}
	r.Header = un.Header()
	return r, nil
}
