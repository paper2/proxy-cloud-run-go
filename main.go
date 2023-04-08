package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"google.golang.org/api/idtoken"
)

type customRoundTripper struct {
	base http.RoundTripper
}

func (c customRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// 認証処理
	audience := "https://" + req.URL.Hostname()
	if err := c.setAuthHeader(req, audience); err != nil {
		log.Print(err)
	}

	start := time.Now()
	resp, err := c.base.RoundTrip(req)

	// log
	log.Printf("%s %s %d %s, duration: %d", req.Method, req.URL.String(), resp.StatusCode, http.StatusText(resp.StatusCode), time.Since(start))
	return resp, err
}

// OIDC TokenをAuth Headerに設定。
func (c customRoundTripper) setAuthHeader(req *http.Request, audience string) error {
	ctx := context.Background()
	ts, err := idtoken.NewTokenSource(ctx, audience)
	if err != nil {
		log.Print("error: NewTokenSource is failed.")
		return err
	}
	token, err := ts.Token()
	if err != nil {
		log.Print("error: failed to get token.")
		return err
	}
	token.SetAuthHeader(req)
	return nil
}

func main() {
	// test 用のバックエンドサーバ
	// backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	dump, err := httputil.DumpRequest(r, false)
	// 	if err != nil {
	// 		fmt.Fprintln(w, err)
	// 	}
	// 	fmt.Fprintln(w, string(dump))
	// }))
	// defer backendServer.Close()

	rpURL, err := url.Parse("https://test-service-sooyb7delq-an.a.run.app")
	if err != nil {
		log.Fatal(err)
	}

	rewrite := func(r *httputil.ProxyRequest) {
		r.SetXForwarded()
		r.SetURL(rpURL)
	}

	// リクエストを受け取り、リバースプロキシを適用するハンドラーを定義する
	proxy := &httputil.ReverseProxy{
		Transport: customRoundTripper{base: http.DefaultTransport},
		Rewrite:   rewrite,
	}
	http.HandleFunc("/", proxy.ServeHTTP)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
