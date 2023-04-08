package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"time"
)

type customRoundTripper struct {
	base http.RoundTripper
}

func (c customRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// 認証処理
	req.Header.Set("X-Forwarded-Proto", "test")

	start := time.Now()
	resp, err := c.base.RoundTrip(req)
	// log
	log.Printf("%s %s %d %s, duration: %d", req.Method, req.URL.String(), resp.StatusCode, http.StatusText(resp.StatusCode), time.Since(start))
	return resp, err
}

func main() {
	// test 用のバックエンドサーバ
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, false)
		if err != nil {
			fmt.Fprintln(w, err)
		}
		fmt.Fprintln(w, string(dump))
	}))
	defer backendServer.Close()

	rpURL, err := url.Parse(backendServer.URL)
	if err != nil {
		log.Fatal(err)
	}

	director := func(req *http.Request) {
		req.URL.Scheme = rpURL.Scheme
		req.URL.Host = rpURL.Host
	}

	// リクエストを受け取り、リバースプロキシを適用するハンドラーを定義する
	proxy := &httputil.ReverseProxy{
		Transport: customRoundTripper{base: http.DefaultTransport},
		Director:  director,
	}
	http.HandleFunc("/", proxy.ServeHTTP)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
