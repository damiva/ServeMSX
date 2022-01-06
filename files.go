package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const pthVideo, pthMusic, extVid, extAud = "video", "music", ".avi.mp4.mkv.mpeg.mpg.m4v.mp2.webm.ts.mts.m2ts.mov.wmv.flv", ".mp3.m4a.flac.wav.wma.aac"

func init() {
	http.HandleFunc("/msx/video/", files)
	http.HandleFunc("/msx/music/", files)
}

func files(w http.ResponseWriter, r *http.Request) {
	p := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/msx/"))
	v := p[0] == 'v'
	if f, e := os.Stat(p); os.IsNotExist(e) {
		panic(404)
	} else if e != nil {
		panic(e)
	} else if !f.IsDir() {
		http.ServeFile(w, r, p)
	} else {
		fs, e := ioutil.ReadDir(p)
		check(e)
		var (
			l       *plist
			ext     string
			ts, ms  []plistObj
			id, pre = r.FormValue("id"), f.Name()
		)
		if v {
			l, ext = plistMedia(r, filepath.Base(p), "msx-white-soft:movie"), extVid
		} else {
			l, ext = plistMedia(r, filepath.Base(p), "msx-white-soft:audiotrack"), extAud
		}
		if pre == pthMusic || pre == pthVideo {
			pre = ""
		} else {
			pre = "{col:msx-white-soft}" + pre + ": {col:msx-white}"
		}
		for _, f := range fs {
			n := f.Name()
			x, u := strings.ToLower(filepath.Ext(n)), "http://"+r.Host+r.URL.EscapedPath()+url.PathEscape(n)
			switch {
			case f.IsDir():
				l.Items = append(l.Items, plistObj{"icon": "msx-yellow:folder", "label": n, "action": "content:" + u + "/"})
			case x == ".torrent":
				if stg.TorrServer != "" {
					ts = append(ts, plistObj{"icon": "msx-yellow:offline-bolt", "label": n, "action": "content:http://" + r.Host + "/msx/torr?id={ID}&link=" + url.QueryEscape(u)})
				}
			case strings.Contains(ext, x):
				ms = append(ms, plistObj{"label": n, "extensionLabel": sizeFormat(f.Size()), "playerLabel": pre + n, "action": playerURL(id, u, v)})
			}
		}
		l.Items = append(l.Items, append(ts, ms...)...)
		l.write(w)
	}
}
func sizeFormat(i int64) string {
	f, b, p := float64(i), "", 0
	for _, b = range []string{" B", " KB", " MB", " GB", " TB"} {
		if f < 1000 {
			break
		} else if b != " TB" {
			f = f / 1024
		}
	}
	switch {
	case f < 10:
		p++
		fallthrough
	case f < 100:
		p++
	}
	return strconv.FormatFloat(f, 'f', p, 64) + b
}
