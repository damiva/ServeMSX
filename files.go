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

const (
	pthVideo, pthMusic, pthPhoto = "video", "music", "photo"
	extVid, extAud, extPic       = ".avi.mp4.mkv.mpeg.mpg.m4v.mp2.webm.ts.mts.m2ts.mov.wmv.flv", ".mp3.m4a.flac.wav.wma.aac", ".jpg.png"
)

func init() {
	for _, f := range [3]string{pthVideo, pthMusic, pthPhoto} {
		http.HandleFunc("/msx/"+f+"/", files)
	}
}

func files(w http.ResponseWriter, r *http.Request) {
	p := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/msx/"))
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
			l      *plist
			ext    string
			z      = "label"
			ts, ms []plistObj
			t      byte
		)
		id := r.FormValue("id")
		switch t = p[0]; t {
		case 'p':
			ext, l, z = extPic, &plist{Type: "list", Head: f.Name(), Ext: "{ico:photo-library}", Template: plistObj{"imageFiller": "smart", "layout": "0,0,3,2"}}, "headline"
		case 'm':
			l, ext = mediaList(r, f.Name(), "{ico:library-music}", false), extAud
		default:
			l, ext = mediaList(r, f.Name(), "{ico:video-library}", true), extVid
		}
		for _, f := range fs {
			n := f.Name()
			x, u := strings.ToLower(filepath.Ext(n)), "http://"+r.Host+r.URL.EscapedPath()+url.PathEscape(n)
			switch {
			case f.IsDir():
				l.Items = append(l.Items, plistObj{"icon": "msx-yellow:folder", z: n, "action": "content:" + u + "/"})
			case x == ".torrent":
				if t != 'p' && stg.TorrServer != "" {
					ts = append(ts, plistObj{"icon": "msx-yellow:offline-bolt", z: n, "action": "content:http://" + r.Host + "/msx/torr?id={ID}&link=" + url.QueryEscape(u)})
				}
			case strings.Contains(ext, x):
				i := plistObj{z: n, "playerLabel": n, "extensionLabel": sizeFormat(f.Size())}
				if t == 'p' {
					i["image"], i["action"] = u, "image:"+u
				} else {
					i["action"] = playerURL(id, u, t == 'v')
				}
				ms = append(ms, i)
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
