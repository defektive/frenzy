package server

import (
	"io"
	"net/http"
	"reflect"
	"testing"
)

func setupHTTPRequest(method string, reqURL string, body io.Reader) *http.Request {

	req, err := http.NewRequest(method, reqURL, body)
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
				req: setupHTTPRequest("GET", "http://defektive.github.io", nil),
			},
			wantHost: "defektivse.github.io",
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
