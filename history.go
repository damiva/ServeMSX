package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	http.HandleFunc("/msx/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			p := new(plist)
			p.mediaList(r, "", false)
			p.opts(r, true)
			for i, f := range stg.HistoryURLs {
				ico := "movie"
				if strings.HasPrefix(f[0], "audio:") {
					ico = "adiotrack"
				}
				if f[2] == "" {
					f[2] = "info:{dic:label:content_not_available|No content!}"
				}
				p.Items = append(p.Items, map[string]string{
					"id":     strconv.Itoa(i),
					"icon":   ico,
					"label":  f[1],
					"action": f[0],
					"data":   f[2],
				})
			}
			p.write(w)
		} else if q := r.URL.Query(); q.Get("del") != "" {
			if i, _ := strconv.Atoi(q.Get("del")); i == 0 {
				stg.HistoryURLs = stg.HistoryURLs[1:]
			} else {
				stg.HistoryURLs = append(stg.HistoryURLs[:i], stg.HistoryURLs[i+1:]...)
			}
			check(stg.save())
			svcAnswer(w, "[success:{dic:rmvd}|reload:content]", nil)
		} else if i := q.Get("goto"); i != "" {
			i, _ := strconv.Atoi(i)
			svcAnswer(w, stg.HistoryURLs[i][3], nil)
		} else {
			var j struct {
				Video struct {
					Info struct {
						URL, Label, Type string
						Index            int
					}
				}
			}
			check(json.NewDecoder(r.Body).Decode(&j))
			u := j.Video.Info.Type + ":" + j.Video.Info.URL
			if q.Has("del") {
				check(stg.historyURL(u))
				svcAnswer(w, `notification:Deleted from history!`, nil)
			} else {
				c := q.Get("src")
				if c != "" {
					c += ">index:" + strconv.Itoa(j.Video.Info.Index)
				}
				check(stg.historyURL(u, j.Video.Info.Label, c))
				svcAnswer(w, `trigger:end:execute:video:info:http://`+r.Host+`/msx/history?del`, nil)
			}
		}
	})
}
