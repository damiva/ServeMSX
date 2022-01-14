package main

import (
	_ "embed"
	"log"
	"net/http"
	"strings"
)

type embedFile struct {
	CT string
	GZ bool
	BS []byte
}

var (
	//go:embed assets/logo.png
	logo []byte
	//go:embed assets/logotype.png
	logotype []byte
	//go:embed assets/index.html.gz
	html   []byte
	server = &http.Server{Addr: ":8008", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				switch v := e.(type) {
				case int:
					http.Error(w, http.StatusText(v), v)
				case string:
					http.Error(w, v, 500)
				case error:
					http.Error(w, v.Error(), 500)
				default:
					http.Error(w, "Unknown error!", 500)
				}
				log.Println(e)
			}
		}()
		out.Println(r.Method, ":", r.RequestURI)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		if r.Method != "OPTIONS" {
			w.Header().Set("Server", Name+"/"+Vers)
			http.DefaultServeMux.ServeHTTP(w, r)
		}
	})}
)

func (f embedFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f.CT != "" {
		w.Header().Set("Content-Type", f.CT)
	}
	if f.GZ {
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.Write(f.BS)
}
func init() {
	http.Handle("/logo.png", embedFile{"image/png", false, logo})
	http.Handle("/logotype.png", embedFile{"image/png", false, logotype})
	http.HandleFunc("/restart", func(w http.ResponseWriter, r *http.Request) {
		svcAnswer(w, "reload", nil)
		go restart()
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			embedFile{"text/html", true, html}.ServeHTTP(w, r)
		} else if p := strings.SplitN(r.URL.Path[1:], "/", 2); len(p) > 1 {
			servePlugin(w, r, p[0], p[1])
		} else {
			panic(404)
		}
	})
}
