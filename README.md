Uncurl
[![Godoc][GodocV2SVG]][GodocV2URL]
=============

A Go library to consume a Chrome/Chromium browser "Copy as cURL" string and generate one or more Go
`*http.Request` objects from it

## Explanation

In the Chrome or Chromium browser, if you open "Developer tools" and go to the Network tab and
navigate somewhere with the browser, you will see a list of requests. Right-clicking one these
requests yields a menu with a Copy submenu. One of those options is "Copy as cURL". It allows you to
paste a `curl` command to your terminal or editor, one that reproduces the request if run.

This library accepts that text input and turns it into a Go request. Further Go requests can be
generated with different target URLs while maintaining the same header values.

## Usage

Initial usage of this package typically follows these steps:

1. Navigate in Chrome/Chromium with dev tools open to network tab, find a request you'd like to
   extract, right-click and select Copy => Copy as cURL
2. In code, call `un, err := uncurl.NewString(str)` on the extracted string
3. Generate new requests via uncurl methods, modify request objects, make requests, and extract new
   URLs from the responses to use in further requests

For a trivial example, let's request the weather in San Diego, and find all the text that looks like a
temperature reading in the returned HTTP body.

### Example 1: Weather in San Diego

```go
package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/jrefior/uncurl"
)

const tempPattern = `-?\d+(?:\.\d+)? ?°`

const weather = `
curl 'https://www.wunderground.com/weather/us/ca/san-diego' -H 'authority: www.wunderground.com' -H 'upgrade-insecure-requests: 1' -H 'user-agent: Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36' -H 'accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9' -H 'sec-fetch-site: none' -H 'sec-fetch-mode: navigate' -H 'accept-encoding: gzip, deflate, br' -H 'accept-language: en-US,en;q=0.9' -H 'cookie: usprivacy=foo; s_fid=bar; s_vi=zip' --compressed`

var tempRe *regexp.Regexp

func init() {
	tempRe = regexp.MustCompile(tempPattern)
}

func main() {
	un, err := uncurl.NewString(weather)
	if err != nil {
		log.Fatalf("Failed to initialize uncurl from weather string: %s", err)
	}
	r := un.Request()
	// modify request here if needed
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatalf("Error making request: %s", err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("Unexpected status code %d", resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %s", err)
	}
	// often would save the body to a file here for later processing, but in this case let's just
	// search it now
	temps := tempRe.FindAll(b, -1)
	if temps == nil {
		log.Fatal("no temps found")
	}
	log.Printf("Temperatures found: %s\n", bytes.Join(temps, []byte(`, `)))
}
```

Output:
```
2020/02/12 07:56:15 Temperatures found: 65°, 48°, 47°, 1.5°, 1.5°, 1.5°, 1.5°, 1.5°, 5°, 41°, 6°, 27°, 81°, 20.9°, 36.4°, 4°, 10°, 5°, 20.9°, 36.4°, 67°, 55°, 67°, 55°, 110°, 80°, 75°, 0°, 30°, 20°, 35°, 58°, 50°, 360°, 360°, 360°, 360°, 360°, 70°, 60°, 36°, 22°, 74°, 70°, 70°, 69°, 3°
```

### Header Value Lists

`net/http` servers do not split request header values into different elements by comma. This was
suspected during review of standard library source code, and then shown via testing. E.g. a `curl`
with this header:

```
-H 'accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9'
```

hit a server handler with this code:

```go
func handler(w http.ResponseWriter, r *http.Request) {    
    for k, v := range r.Header {    
        fmt.Printf("%28s %2d\n", k, len(v))    
    }    
}    
```
which showed the accept header length was 1. A header value slice length of 1 was found for all
headers tested, e.g.:
```
             Accept-Encoding  1
   Upgrade-Insecure-Requests  1
                  User-Agent  1
              Sec-Fetch-Mode  1
             Accept-Language  1
                   Authority  1
                      Accept  1
              Sec-Fetch-Site  1
```

That behavior is reproduced here on the client side -- no effort is made to split curl `-H` arguments
into different elements.

[GodocV2SVG]: https://godoc.org/github.com/jrefior/uncurl?status.svg
[GodocV2URL]: https://godoc.org/github.com/jrefior/uncurl
