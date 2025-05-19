package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/base32"
	"fmt"
	"github.com/defektive/frenzy/pkg/common"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var ProxyFlag bool

type Config struct {
	Serve Serve  `yaml:"serve"`
	Proxy Proxy  `yaml:"proxy"`
	Rule  []Rule `yaml:"rule"`
}
type Serve struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}
type Proxy struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}
type Rule struct {
	Name    string `yaml:"Name"`
	Search  string `yaml:"Search"`
	Replace string `yaml:"Replace"`
}

var config Config

func init() {
	config.LoadConfig("config.yaml")
}

func (c *Config) LoadConfig(file string) {
	fileIn, err := os.ReadFile(file)
	common.EnsureNotError(err)

	err = yaml.Unmarshal(fileIn, c)
	common.EnsureNotError(err)
}

func handler(w http.ResponseWriter, r *http.Request) {
	slog.Debug("new request", slog.String("method", r.Method), slog.String("url", r.URL.String()))

	req := r.Clone(context.Background())
	req.RequestURI = ""

	for _, rule := range config.Rule {
		rule.ReplaceRequest(req)
	}

	httpClient := &http.Client{}
	if config.Proxy.Enabled {
		proxyURL, _ := url.Parse("http://127.0.0.1:8081")
		proxy := http.ProxyURL(proxyURL)
		transport := &http.Transport{
			Proxy: proxy,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		httpClient = &http.Client{Transport: transport}
	}

	slog.Debug("send request", slog.String("method", req.Method), slog.String("url", req.URL.String()))
	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Error("failed to send request", "err", err, "url", req.URL.String())
		w.WriteHeader(509)
		return
	}

	for _, rule := range config.Rule {
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

func ServeWebPages() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}

func Start() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	fmt.Printf("proxy enabled: %v\n", config.Proxy.Enabled)
	ServeWebPages()
	http.HandleFunc("/", handler)
	fmt.Printf("Server listening on  %s:%v\n", config.Serve.Address, config.Serve.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%v", config.Serve.Address, config.Serve.Port), nil))
}

func (r *Rule) ReplaceRequest(req *http.Request) {
	//urlStr := fmt.Sprintf("%s://%s%s", "http", req.Host, req.URL.String())
	//urlStr = strings.ReplaceAll(urlStr, r.Search, r.Replace)

	if r.Name == "Change Host" {
		fmt.Println("[+] Evaluating Host String")
		fmt.Println("Incoming req.host: ", req.Host)
		fmt.Println("r.search: ", r.Search)
		fmt.Println("r.replace: ", r.Replace)
		req.Host = strings.ReplaceAll(req.Host, r.Search, r.Replace)
		fmt.Println("Outgoing req.host: ", req.Host)
	} else if r.Name == "Change Path" {
		fmt.Println("[+] Evaluating Path String")
		fmt.Println("Incoming req.URL.Path: ", req.URL.Path)
		fmt.Println("r.search: ", r.Search)
		fmt.Println("r.replace: ", r.Replace)
		req.URL.Path = strings.ReplaceAll(req.URL.Path, r.Search, r.Replace)
		fmt.Println("Outgoing req.URL.Path: ", req.URL.Path)
	}

	urlStr := fmt.Sprintf("%s://%s%s", "https", req.Host, req.URL.Path)
	fmt.Println("urlStr: ", urlStr)

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
	//slog.Debug("match response", "resp", string(respBody))
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
