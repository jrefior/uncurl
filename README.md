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

### Header Value Lists

`net/http` servers do not split request header values into different elements by comma. This was
suspected during review of standard library source code, and then shown via testing. E.g. a `curl`
with this header:

```
-H 'accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9'
```

hit a server handler with this code:

```
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
