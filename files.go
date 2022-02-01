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
	extVid, extAud, extPic       = ".avi.mp4.mkv.mpeg.mpg.m4v.mp2.webm.ts.mts.m2ts.mov.wmv.flv", ".mp3.m4a.flac.wav.wma.aac.ogg", ".jpg.jpeg.png"
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
	} else if f.IsDir() {
		fs, e := ioutil.ReadDir(p)
		check(e)
		var (
			l        *plist
			ext, hdr string
			z        = "label"
			ts, ms   []plistObj
			t        byte
		)
		id := r.FormValue("id")
		if strings.IndexRune(p, filepath.Separator) > 0 {
			hdr = "{ico:msx-white-soft:folder-open} " + f.Name()
		}
		switch t = p[0]; t {
		case 'p':
			opt := plistObj{"key": "green", "label": "{dic:Files|List of files} {ico:apps}", "action": "execute:http://" + r.Host + "/settings?id={ID}", "data": cPhoto}
			ext = extPic
			if stg.Clients[id]&cPhoto != 0 {
				opt["label"] = "{dic:Files|List of files} {ico:list}"
				z, l = "title", &plist{
					Type: "list", Head: hdr, Ext: "{ico:msx-white:photo-library}",
					Template: plistObj{"type": "separate", "imageFiller": "smart", "layout": "0,0,4,4"}, Compress: true,
					Options: options(opt, plistObj{"key": "yellow", "label": "{dic:Up|up}", "action": "focus:index:0"}),
				}
			} else {
				l = mediaList(r, hdr, "{ico:msx-white:photo-library}", "image", []plistObj{opt}, false, false, false)
			}
			l.Flag = "photo"
		case 'm':
			l, ext = mediaList(r, hdr, "{ico:msx-white:library-music}", "audiotrack", nil, false, true, true), extAud
		default:
			l, ext = mediaList(r, hdr, "{ico:msx-white:video-library}", "movie", nil, false, false, true), extVid
		}
		for _, f := range fs {
			n := f.Name()
			x, u := strings.ToLower(filepath.Ext(n)), "http://"+r.Host+r.URL.EscapedPath()+url.PathEscape(n)
			switch {
			case f.IsDir():
				l.Items = append(l.Items, plistObj{"icon": "msx-yellow:folder", "label": n, "action": "content:" + u + "/?id={ID}"})
			case x == ".torrent":
				if t != 'p' && stg.TorrServer != "" {
					ts = append(ts, plistObj{"icon": "msx-yellow:offline-bolt", "label": n, "action": "content:http://" + r.Host + "/msx/torr?id={ID}&link=" + url.QueryEscape(u)})
				}
			case t != 'p' && stg.Background != "" && n == stg.Background:
				l.Background = u
			case strings.Contains(ext, x):
				i := plistObj{z: n, "playerLabel": n, "extensionLabel": sizeFormat(f.Size())}
				switch t {
				case 'p':
					i["action"] = "image:" + u
					if stg.Clients[id]&cPhoto != 0 {
						i["image"] = strings.Replace(u, "/msx/", pthFFmpeg, 1)
					}
				case 'm':
					i["cover"] = strings.Replace(u, "/msx/", pthFFmpeg, 1)
					fallthrough
				default:
					i["action"] = playerURL(id, u, t == 'v')
				}
				ms = append(ms, i)
			}
		}
		l.Items = append(l.Items, append(ts, ms...)...)
		l.write(w)
	} else {
		http.ServeFile(w, r, p)
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
