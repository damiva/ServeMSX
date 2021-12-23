package main

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
)

func init() {
	http.HandleFunc("/msx/torr", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("detect") {
			t := r.Host[:strings.LastIndexByte(r.Host, ':')+1] + "8090"
			if _, e := checkTorr(t); e == nil {
				stg.TorrServer = t
			}
			svcAnswer(w, "[]", nil)
		} else if p := r.FormValue("link"); p != "" {
			torrLink(w, r, p)
		} else if p = r.FormValue("del"); p != "" {
			torrDel(w, p, true)
		} else if p = r.FormValue("drop"); p != "" {
			torrDel(w, p, false)
		} else {
			torrMain(w, r)
		}
	})
}
func torrMain(w http.ResponseWriter, r *http.Request) {
	var d []struct {
		Hash, Poster, Title, Stat_string string
		Stat, Active_peers, Total_peers  int
		Torrent_size                     int64
	}
	check(download("http://"+stg.TorrServer+"/torrents", &d, map[string]string{"action": "list"}))
	sort.Slice(d, func(i, j int) bool { return d[i].Stat < d[j].Stat })
	l := &plist{Type: "list"}
	c := [6]string{"yellow", "yellow", "yellow", "green", "red", "white-soft"}
	for _, i := range d {
		l.Items = append(l.Items, map[string]string{
			"id":       i.Hash,
			"headline": i.Title,
			"image":    i.Poster,
			"titleFooter": "{ico:attach-file}{tb}" + sizeFormat(i.Torrent_size) + "{col:msx-" + c[i.Stat] +
				"}{br}{ico:arrow-upward}{tb}" + strconv.Itoa(i.Active_peers) + " / " + strconv.Itoa(i.Total_peers) +
				"{br}{ico:flag}{tb}{dic:torrent" + strconv.Itoa(i.Stat) + "}",
			"action": "content:http://" + r.Host + "/msx/torr?link=" + i.Hash,
		})
	}
	l.Template = map[string]interface{}{"imageWidth": 1.25, "layout": "0,0,6,2", "imageFiller": "height-left", "icon": "msx-glass:bolt"}
	l.opts(r, false)
	l.write(w)
}
func torrDel(w http.ResponseWriter, i string, del bool) {
	var b []byte
	a := "drop"
	if del {
		a = "rem"
	}
	check(download("http://"+stg.TorrServer+"/torrents", &b, map[string]string{"action": a, "hash": i}))
	svcAnswer(w, "[success:{dic:"+a+"d}|reload:content]", nil)
}
func torrLink(w http.ResponseWriter, r *http.Request, p string) {
	var t struct {
		File_stats []struct {
			ID     int
			Length int64
			Path   string
		}
		Title                     string
		Torrent_size              int64
		Active_peers, Total_peers int
	}
	check(download("http://"+stg.TorrServer+"/stream/?stat&link="+url.QueryEscape(p), &t))
	var as []map[string]string
	l, tu := &plist{Head: t.Title, Ext: sizeFormat(t.Torrent_size)}, "http://"+stg.TorrServer+"/stream/"
	l.mediaList(r, "", true)
	if t.Active_peers > 0 || t.Total_peers > 0 {
		l.Ext += "{br}{ico:arrow-upward} " + strconv.Itoa(t.Active_peers) + "/" + strconv.Itoa(t.Total_peers)
	}
	for _, f := range t.File_stats {
		n, e, u := path.Base(f.Path), sizeFormat(f.Length), tu+url.PathEscape(f.Path)+"?play&link="+url.QueryEscape(p)+"&index="+strconv.Itoa(f.ID)
		t.Title = "{col:msx-white-soft}{ico:bolt}" + t.Title + ": {col:msx-white}"
		if x := strings.ToLower(path.Ext(n)); strings.Contains(extAud, x) {
			as = append(as, map[string]string{
				"label":          n,
				"icon":           "msx-green:audiotrack",
				"group":          "{dic:label:audio|Audio}",
				"playerLabel":    t.Title + n,
				"extensionLabel": e,
				"action":         "audio:" + u,
			})
		} else if strings.Contains(extVid, x) {
			l.Items = append(l.Items, map[string]string{
				"label":          n,
				"icon":           "msx-green:movie",
				"group":          "{dic:label:video|Video}",
				"playerLabel":    t.Title + n,
				"extensionLabel": e,
				"action":         "video:" + u,
			})
		}
	}
	l.Items = append(l.Items, as...)
	l.write(w)
}
func checkTorr(a string) (v string, e error) {
	if a != "" {
		var b []byte
		if e = download("http://"+a+"/echo", &b, nil); e == nil {
			if v = string(b); !strings.HasPrefix(v, "MatriX") {
				e = errors.New("TorrServer v. " + v + " is not supported!")
			}
		}
	}
	return
}
