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
		if p := r.FormValue("link"); p != "" {
			torrLink(w, r, p)
		} else if p = r.FormValue("del"); p != "" {
			torrDel(w, p, true)
		} else if p = r.FormValue("drop"); p != "" {
			torrDel(w, p, false)
		} else if p = r.FormValue("add"); p != "" {
			torrAdd(w, p, r.FormValue("ttl"), r.FormValue("img"))
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
	ly, cm := "0,0,6,2", stg.Compress[r.FormValue("id")]
	if cm {
		ly = "0,0,8,2"
	}
	check(download("http://"+stg.TorrServer+"/torrents", &d, map[string]string{"action": "list"}))
	sort.Slice(d, func(i, j int) bool { return d[i].Stat < d[j].Stat })
	l := &plist{Type: "list", Ext: "{ico:msx-white:bolt} TorrServer: " + stg.TorrServer, Compress: cm, Template: plistObj{
		"imageWidth": 1.25, "layout": ly, "imageFiller": "height-left", "icon": "msx-glass:bolt",
		"options": plistObj{
			"headline": "{dic:label:menu|Menu}",
			"caption":  "{dic:label:menu|Menu}:{tb}{ico:msx-red:stop} {dic:Remove|Remove}{tb}{ico:msx-yellow:stop} {dic:Drop|Drop}",
			"template": plistObj{"enumerate": false, "type": "control", "layout": "0,0,8,1"},
			"items": []plistObj{
				{"key": "red", "icon": "msx-red:stop", "label": "{dic:Remove|Remove}", "action": "execute:fetch:http://" + r.Host + "/msx/torr?del={context:id}"},
				{"key": "yellow", "icon": "msx-yellow:stop", "label": "{dic:Drop|Drop}", "action": "execute:fetch:http://" + r.Host + "/msx/torr?drop={context:id}"},
			},
		},
	}}
	c := [6]string{"yellow", "yellow", "yellow", "green", "red", "white-soft"}
	for _, i := range d {
		l.Items = append(l.Items, plistObj{
			"id":       i.Hash,
			"headline": i.Title,
			"image":    i.Poster,
			"titleFooter": "{col:msx-" + c[i.Stat] + "}{ico:flag} {dic:Torrent" + strconv.Itoa(i.Stat) + "}" +
				"{br}{ico:attach-file} " + sizeFormat(i.Torrent_size) +
				"{tb}{ico:arrow-upward} " + strconv.Itoa(i.Active_peers) + " / " + strconv.Itoa(i.Total_peers),
			"action": "content:http://" + r.Host + "/msx/torr?noadd&id={ID}&link=" + i.Hash,
		})
	}
	l.write(w)
}
func torrDel(w http.ResponseWriter, i string, del bool) {
	var b []byte
	a := "drop"
	if del {
		a = "rem"
	}
	check(download("http://"+stg.TorrServer+"/torrents", &b, map[string]string{"action": a, "hash": i}))
	svcAnswer(w, "reload:content", nil)
}
func torrAdd(w http.ResponseWriter, l, t, i string) {
	u := "http://" + stg.TorrServer + "/stream/?save&link=" + url.QueryEscape(l)
	if t != "" {
		u += "&title=" + url.QueryEscape(t)
	}
	if i != "" {
		u += "&poster=" + url.QueryEscape(i)
	}
	r, e := http.Get(u)
	check(e)
	if r.StatusCode != 200 {
		panic(r.StatusCode)
	}
	svcAnswer(w, "info:OK", nil)
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
	check(download("http://"+stg.TorrServer+"/stream/?stat&link="+url.QueryEscape(p), &t, nil))
	var as []plistObj
	l, tu, id := mediaList(r, "{ico:msx-white-soft:offline-bolt} "+t.Title, "{ico:msx-white:attach-file} "+sizeFormat(t.Torrent_size), "movie"), "http://"+stg.TorrServer+"/stream/", r.FormValue("id")
	if t.Active_peers > 0 || t.Total_peers > 0 {
		l.Ext += "{tb}{ico:msx-white:arrow-upward} " + strconv.Itoa(t.Active_peers) + "/" + strconv.Itoa(t.Total_peers)
	}
	for _, f := range t.File_stats {
		n, e, u := path.Base(f.Path), sizeFormat(f.Length), tu+url.PathEscape(f.Path)+"?play&link="+url.QueryEscape(p)+"&index="+strconv.Itoa(f.ID)
		if x := strings.ToLower(path.Ext(n)); strings.Contains(extAud, x) {
			as = append(as, plistObj{
				"label":          n,
				"icon":           "msx-white-soft:audiotrack",
				"group":          "{dic:label:audio|Audio}",
				"playerLabel":    n,
				"extensionLabel": e,
				"action":         playerURL(id, u, false),
			})
		} else if strings.Contains(extVid, x) {
			l.Items = append(l.Items, plistObj{
				"label":          n,
				"icon":           "msx-white-soft:movie",
				"group":          "{dic:label:video|Video}",
				"playerLabel":    n,
				"extensionLabel": e,
				"action":         playerURL(id, u, true),
			})
		}
	}
	l.Items = append(l.Items, as...)
	if !r.Form.Has("noadd") {
		l.Template["options"] = plistObj{
			"headline": "{dic:label:menu|Menu}",
			"caption":  "{dic:label:menu|Menu}:{tb}{ico:msx-green:stop} {dic:AddTorr|Add this torrent to My torrents}",
			"template": plistObj{"enumerate": false, "type": "control", "layout": "0,0,8,1"},
			"items": []plistObj{{
				"key":    "green",
				"icon":   "msx-green:stop",
				"label":  "{dic:AddTorr|Add to My torrents}",
				"action": "execute:fetch:http://" + r.Host + "/msx/torr?add=" + url.QueryEscape(p) + "&ttl=" + url.QueryEscape(r.FormValue("ttl")) + "&img=" + url.QueryEscape(r.FormValue("img")),
			}}}
	}
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
