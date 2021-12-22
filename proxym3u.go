package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func init() {
	http.HandleFunc("/proxy.m3u", func(w http.ResponseWriter, r *http.Request) {
		check(r.ParseForm())
		u := r.FormValue("link")
		if u == "" {
			panic(400)
		}
		ul, e := url.Parse(u)
		check(e)
		q, e := http.NewRequest("GET", u, nil)
		check(e)
		if len(r.Form["header"]) > 0 {
			for _, h := range r.Form["header"] {
				if v := strings.SplitN(h, ":", 2); len(v) > 1 {
					q.Header.Add(v[0], v[1])
				}
			}
		}
		a, e := http.DefaultClient.Do(q)
		check(e)
		defer a.Body.Close()
		for h, vs := range a.Header {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
		if a.StatusCode != 200 {
			log.Println(a.StatusCode, "ON /proxy.m3u8:", u, " header:", r.Form["header"])
		}
		w.WriteHeader(a.StatusCode)
		b := bufio.NewReader(a.Body)
		for e == nil {
			var s string
			if s, e = b.ReadString('\n'); e == nil || e == io.EOF {
				if len(s) == 0 {
					continue
				} else if s[0] != '#' {
					if us, e := url.Parse(strings.TrimSpace(s)); e == nil {
						/*
							if us := ul.ResolveReference(us); strings.HasSuffix(us.Path, ".m3u8") {
								s = []byte("http://" + r.Host + "/proxy.m3u8?link=" + url.QueryEscape(us.String()))
								for _, h := range r.Form["header"] {
									s = append(s, "&header="...)
									s = append(s, []byte(url.QueryEscape(h))...)
								}
								s = append(s, '\r', '\n')
							} else {
						*/
						s = ul.ResolveReference(us).String() + "\r\n"
						//}
					}
				} else if strings.HasPrefix(s, "#EXT-X-MEDIA") {
					ss := strings.Split(s, ",")
					if l := len(ss); l > 1 && strings.HasPrefix(ss[l-1], "URI=\"") {
						sss := ss[l-1][5:]
						if us, e := url.Parse(sss[:strings.IndexByte(sss, '"')]); e == nil {
							ss[l-1] = "URI=\"" + ul.ResolveReference(us).String() + "\"\r\n"
							s = strings.Join(ss, ",")
						}
					}
				}
				w.Write([]byte(s))
			}
		}
	})
}
