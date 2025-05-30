package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base32"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("new request", slog.String("method", r.Method), slog.String("url", r.URL.String()))

	req := r.Clone(context.Background())
	req.RequestURI = ""

	for _, rule := range rules {
		rule.ReplaceRequest(req)
	}

	httpClient := &http.Client{}

	slog.Debug("send request", slog.String("method", req.Method), slog.String("url", req.URL.String()))
	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Error("failed to send request", "err", err, "url", req.URL.String())
		w.WriteHeader(509)
		return
	}

	for _, rule := range rules {
		rule.ReplaceResponse(resp)
	}

	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read body", "err", err, "url", req.URL.String())
		w.WriteHeader(509)
		return
	}
	w.Write(respBody)

}

func Start() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	http.HandleFunc("/", handler)
	fmt.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

var rules []Rule

func init() {
	rules = append(rules, Rule{
		Name:    "Change host",
		Search:  "http://localhost:8080",
		Replace: "https://defektive.github.io",
	})
}

type Rule struct {
	Name    string `json:"name"`
	Search  string `json:"search"`
	Replace string `json:"replace"`
}

func (r *Rule) ReplaceRequest(req *http.Request) {

	urlStr := fmt.Sprintf("%s://%s%s", "http", req.Host, req.URL.String())
	urlStr = strings.ReplaceAll(urlStr, r.Search, r.Replace)

	parsed, err := url.Parse(urlStr)
	if err != nil {
		log.Println(err)
		return
	}

	req.Host = parsed.Host
	req.Proto = parsed.Scheme
	req.URL = parsed
}

func (r *Rule) ReplaceResponse(resp *http.Response) {

	slog.Debug("replace response", slog.String("url", resp.Request.URL.String()))
	for k, v := range resp.Header {
		for i, vv := range v {
			resp.Header[k][i] = strings.ReplaceAll(vv, r.Search, r.Replace)
		}
	}

	bodyReader := resp.Body

	var err error
	if resp.Header.Get("Content-Encoding") == "gzip" {
		resp.Header.Del("Content-Encoding")
		bodyReader, err = gzip.NewReader(bodyReader)
		if err != nil {
			slog.Error("failed to create gzip reader", "err", err)
		}
		defer bodyReader.Close()
	}

	respBody, err := io.ReadAll(bodyReader)
	if err != nil {
		slog.Error("failed to read body", "err", err)
		return
	}

	respBody = bytes.ReplaceAll(respBody, []byte(r.Search), []byte(r.Replace))

	urlMatch := regexp.MustCompile(`://([^/]+)`)

	matches := urlMatch.FindAll(respBody, -1)
	slog.Debug("match response", "resp", string(respBody))
	for _, match := range matches {
		slog.Debug("match url", "match", string(match), "replace", string(encodeDomain(match)))
		respBody = bytes.ReplaceAll(respBody, match, encodeDomain(match))
	}

	resp.Body = io.NopCloser(strings.NewReader(string(respBody)))

}

func encodeDomain(domain []byte) []byte {
	newDomain := base32.StdEncoding.EncodeToString(domain)
	newDomain = "://localhost:8080/x.x/" + newDomain
	return []byte(newDomain)
}
