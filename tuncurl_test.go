package uncurl

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"
)

func headerEq(a, b http.Header) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, present := b[k]
		if !present {
			return false
		}
		if len(bv) != len(v) {
			return false
		}
		for x := 0; x < len(v); x++ {
			if bv[x] != v[x] {
				return false
			}
		}
	}
	return true
}

func TestNewUncurl(t *testing.T) {
	tests := []struct {
		curl   string
		target string
		header http.Header
		method string
		ae     string
		body   []byte
	}{
		{
			`curl 'https://www.wunderground.com/forecast/us/ma/waltham' -H 'authority: www.wunderground.com' -H 'upgrade-insecure-requests: 1' -H 'user-agent: Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36' -H 'accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9' -H 'sec-fetch-site: none' -H 'sec-fetch-mode: navigate' -H 'accept-encoding: gzip, deflate, br' -H 'accept-language: en-US,en;q=0.9' --compressed`,
			"https://www.wunderground.com/forecast/us/ma/waltham",
			http.Header{
				"authority":                 []string{"www.wunderground.com"},
				"upgrade-insecure-requests": []string{"1"},
				"user-agent":                []string{"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36"},
				"accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
				"sec-fetch-site":            []string{"none"},
				"sec-fetch-mode":            []string{"navigate"},
				"accept-language":           []string{"en-US,en;q=0.9"},
			},
			`GET`,
			"gzip, deflate, br",
			nil,
		},
		{
			`curl 'https://privnote.com/legacy/' -H 'Connection: keep-alive' -H 'Origin: https://privnote.com' -H 'X-Requested-With: XMLHttpRequest' -H 'User-Agent: Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36' -H 'Content-type: application/x-www-form-urlencoded' -H 'Accept: */*' -H 'Sec-Fetch-Site: same-origin' -H 'Sec-Fetch-Mode: cors' -H 'Referer: https://privnote.com/' -H 'Accept-Encoding: gzip, deflate, br' -H 'Accept-Language: en-US,en;q=0.9' --data '&data=U2FsdGVkX1%2BOxTSDTgLVqVwnRWjcvJ8AVWWZJkN456o%3D%0A&has_manual_pass=false&duration_hours=0&dont_ask=false&data_type=T&notify_email=&notify_ref=' --compressed`,
			"https://privnote.com/legacy/",
			http.Header{
				"Connection":       []string{"keep-alive"},
				"Origin":           []string{"https://privnote.com"},
				"X-Requested-With": []string{"XMLHttpRequest"},
				"User-Agent":       []string{"Mozilla/5.0 (X11; Fedora; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36"},
				"Content-type":     []string{"application/x-www-form-urlencoded"},
				"Accept":           []string{"*/*"},
				"Sec-Fetch-Site":   []string{"same-origin"},
				"Sec-Fetch-Mode":   []string{"cors"},
				"Referer":          []string{"https://privnote.com/"},
				"Accept-Language":  []string{"en-US,en;q=0.9"},
			},
			`POST`,
			"gzip, deflate, br",
			[]byte(`&data=U2FsdGVkX1%2BOxTSDTgLVqVwnRWjcvJ8AVWWZJkN456o%3D%0A&has_manual_pass=false&duration_hours=0&dont_ask=false&data_type=T&notify_email=&notify_ref=`),
		},
	}
	for i, test := range tests {
		un, err := NewUncurlString(test.curl)
		if err != nil {
			t.Errorf("Error uncurling %d: %s", i, err)
		}
		if un == nil {
			t.Errorf("un is nil in test %d", i)
		}
		if un.String() != test.curl {
			t.Errorf("curl string mismatch at test %d", i)
		}
		if un.Target() != test.target {
			t.Errorf("Target mismatch in test %d: expected %s, got %s", i, test.target, un.Target())
		}
		if !headerEq(test.header, un.Header()) {
			t.Errorf("Headers not equal in test %d", i)
		}
		if un.method != test.method {
			t.Errorf("Methods not equal in test %d: expected %s, got %s", i, test.method, un.method)
		}
		if un.Method() != test.method {
			t.Errorf("un.Method() mismatch at test %d", i)
		}
		if un.AcceptEncoding != test.ae {
			t.Errorf("accept-encoding mismatch in test %d: expected %s, got %s", i, test.ae, un.AcceptEncoding)
		}
		r := un.Request()
		requestTest(t, i, un, test.header, test.method, test.body, r)
		r, err = un.NewRequest(un.Method(), un.Target(), bytes.NewBuffer(test.body))
		if err != nil {
			t.Errorf("NewRequest error in test %d: %s", i, err)
		}
		requestTest(t, i, un, test.header, test.method, test.body, r)
		r, err = un.NewRequestWithContext(context.Background(), un.Method(), un.Target(), bytes.NewBuffer(test.body))
		if err != nil {
			t.Errorf("NewRequestWithContext error in test %d: %s", i, err)
		}
		requestTest(t, i, un, test.header, test.method, test.body, r)
	}
}

func requestTest(t *testing.T, i int, un *Uncurl, th http.Header, tm string, tb []byte, r *http.Request) {
	if !headerEq(th, r.Header) {
		t.Errorf("r.Header mismatch in test %d", i)
	}
	if r.Method != tm {
		t.Errorf("r.Method mismatch in test %d", i)
	}
	if tb != nil {
		bodyTest(t, i, un, tb, r)
	}
}

func bodyTest(t *testing.T, i int, un *Uncurl, tb []byte, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("error reading body in test %d: %s", i, err)
	}
	if !bytes.Equal(tb, b) {
		t.Errorf("body mismatch at test %d", i)
	}
	if !bytes.Equal(tb, un.Body()) {
		t.Errorf("un.Body() mismatch at test %d", i)
	}
}
