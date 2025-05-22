package server

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func setupHTTPRequest(method string, reqURL string, body []byte) *http.Request {

	parsed, err := url.Parse(reqURL)
	if err != nil {
		panic(err)
	}

	trailer := "\r\n\r\n"
	if len(body) > 0 {
		trailer = fmt.Sprintf("\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	}

	b := bytes.NewReader([]byte(fmt.Sprintf("%s %s HTTP/1.1\r\nHost: %s%s", method, parsed.Path, parsed.Host, trailer)))
	br := bufio.NewReader(b)

	req, err := http.ReadRequest(br)
	if err != nil {
		panic(err)
	}

	return req
}

func TestRule_ReplaceRequest(t *testing.T) {

	type fields struct {
		Name    string
		Search  string
		Replace string
	}
	type args struct {
		req *http.Request
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantHost string
	}{
		{
			name: "Replace host",
			fields: fields{
				Name:    "replace host",
				Search:  "http://localhost:8080",
				Replace: "https://defektive.github.io",
			},
			args: args{
				req: setupHTTPRequest("GET", "http://defektive.github.io/", nil),
			},
			wantHost: "defektive.github.io",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rule{
				Name:    tt.fields.Name,
				Search:  tt.fields.Search,
				Replace: tt.fields.Replace,
			}
			r.ReplaceRequest(tt.args.req)

			if got := tt.args.req.Host; !reflect.DeepEqual(got, tt.wantHost) {
				t.Errorf("replace host = %v, want %v", got, tt.wantHost)
			}
		})
	}
}

func Test_encodeDomain(t *testing.T) {
	type args struct {
		domain []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Replace domain",
			args: args{
				domain: encodeDomain([]byte("defektive.github.io")),
			},
			want: []byte("pizza"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encodeDomain(tt.args.domain); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rewriteRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want *http.Request
	}{
		{
			name: "request gets modified",
			args: args{
				r: setupHTTPRequest("GET", "http://localhost:8080/", nil),
			},
			want: setupHTTPRequest("GET", "https://defektive.github.io/", nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewriteRequest(tt.args.r)
			//if !reflect.DeepEqual(tt.args.r, tt.want) {
			//	t.Errorf("rewriteRequest() = %v, want %v", tt.args.r, tt.want)
			//}

			if !reflect.DeepEqual(tt.args.r.Host, tt.want.Host) {
				t.Errorf("rewriteRequest.Host() = %v, want %v", tt.args.r.Host, tt.want.Host)
			}
		})
	}
}
