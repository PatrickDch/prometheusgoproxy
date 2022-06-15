package main

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/gorilla/context"
	"github.com/keep94/weblogs"
)

func main() {

	logPath := "/data/promproxy/access.log"

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(logFile)
	log.SetOutput(mw)

	fmt.Printf("listening on %v\n", "4445")
	fmt.Printf("Logging to %v\n", logPath)

	director := func(req *http.Request) {

		ua := req.Header.Get("User-Agent")
		promproxy := req.Header.Get("Host")

		s := promproxy

		se := base64.StdEncoding.EncodeToString([]byte(s))

		req.Header.Add("X-Origin-Host", promproxy)
		req.Header.Add("CheckMe", se)
		req.URL.Scheme = "https"
		req.Header.Add("Host", promproxy)
		req.URL.Path = "/metrics"
		req.Header.Add("User-Agent", ua)
		req.Header.Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}

	proxy := &httputil.ReverseProxy{Director: director}

	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP384, tls.X25519},
		PreferServerCipherSuites: true,
		InsecureSkipVerify:       true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
	}
	loghandler := context.ClearHandler(weblogs.HandlerWithOptions(
		http.DefaultServeMux,
		&weblogs.Options{
			Writer: mw,
			Logger: weblogs.ApacheCombinedLogger(),
		}))

	srv := &http.Server{
		Addr:         ":4445",
		Handler:      loghandler,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)

	})

	log.Fatal(srv.ListenAndServeTLS("/data/cert/server.crt", "/data/cert/server.key"))

}
