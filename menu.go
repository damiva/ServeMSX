package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

var startFocus string

func init() {
	http.HandleFunc("/msx/start.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"` + Name + `","version":"` + Vers + `","parameter": "menu:http://{SERVER}/msx/menu.json"}`))
	})
	http.HandleFunc("/msx/menu.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"reuse":false,"cache":false,"restore":false,"name":"` + Name + `","version":"` + Vers +
			`","reference":"http://` + r.Host + `/msx/menu?player={PLAYER}&platform={PLATFORM}&id={ID}` + startFocus))
		startFocus = ""
		if _, e := os.Stat(pthDic); e == nil {
			w.Write([]byte(`","dictionary":"http://` + r.Host + "/msx/" + pthDic))
		} else if !os.IsNotExist(e) {
			log.Println(e)
		}
		w.Write([]byte(`"}`))
	})
	http.HandleFunc("/msx/menu", func(w http.ResponseWriter, r *http.Request) {
		pls, id, u := strings.SplitN(r.FormValue("player"), "/", 2)[0], r.FormValue("id"), "http://"+r.Host
		clients[id] = client{r.RemoteAddr, r.FormValue("platform"), pls, r.FormValue("v")}
		if stg.HTML5X[id] {
			pls += " {col:msx-white}{ico:toggle-on} html5x"
		} else {
			pls = "{txt:msx-white:" + pls + "} {ico:msx-white:toggle-off} html5x"
		}
		l := &plist{Logo: u + "/logotype.png", Menu: []plistObj{ /*{"icon": "history", "label": "{dic:Recent|Continue}...", "data": u + "/msx/recent?id={ID}"}*/ }}
		for _, f := range [3][2]string{{pthVideo, "video-library"}, {pthMusic, "library-music"}, {pthPhoto, "photo-library"}} {
			if _, e := os.Stat(f[0]); !os.IsNotExist(e) {
				l.Menu = append(l.Menu, plistObj{"icon": f[1], "extensionIcon": "folder-open", "label": "{dix:" + f[0] + "}My " + f[0], "data": u + "/msx/" + f[0] + "/?id={ID}"})
			}
		}
		ta := "{dic:label:none|None}"
		if stg.TorrServer != "" {
			ta = "{col:msx-white}" + stg.TorrServer
			l.Menu = append(l.Menu, plistObj{"image": "http://" + stg.TorrServer + "/apple-touch-icon.png", "extensionIcon": "bolt", "label": "{dic:Torrents|My torrents}", "data": u + "/msx/torr?id={ID}"})
		}
		l.Menu = append(l.Menu, plistObj{"type": "separator"})
		if ps, e := plugsInfo(); e == nil {
			ml := len(l.Menu)
			for _, p := range ps {
				if p.Error == "" {
					if !p.Torrent || stg.TorrServer != "" {
						m := plistObj{"label": p.Label}
						if p.Torrent {
							m["extensionIcon"] = "bolt"
						} else {
							m["extensionIcon"] = "http"
						}
						if p.Label == "" {
							m["label"] = p.Name
						}
						if p.URL == "" {
							m["data"] = u + "/" + p.Name + "/"
						} else {
							m["data"] = strings.ReplaceAll(p.URL, "{BASE_URL}", u+"/"+p.Name+"/")
						}
						if p.Image != "" {
							m["image"] = strings.ReplaceAll(p.Image, "{BASE_URL}", u+"/"+p.Name+"/")
						} else if p.Icon != "" {
							m["icon"] = p.Icon
						}
						l.Menu = append(l.Menu, m)
					}
				}
			}
			if len(l.Menu) > ml {
				l.Menu = append(l.Menu, plistObj{"type": "separator"})
			}
		}
		ts := started.UnixMilli()
		lst := "{txt:msx-white:dic:Default|default} {ico:msx-white:toggle-off} {dic:Compress|compressed}"
		if stg.Compress[id] {
			lst = "{dic:Default|default} {col:msx-white}{ico:toggle-on} {dic:Compress|compressed}"
		}
		l.Menu = append(l.Menu, plistObj{"id": "stg", "icon": "settings", "label": "{dic:label:settings|Settings}",
			"data": plistObj{"extension": "{ico:msx-white:settings}", "pages": []map[string][]plistObj{{"items": {
				{"type": "space", "layout": "0,0,12,2", "image": u + "/logotype.png", "imageFiller": "height", "imageWidth": 7, "imagePreload": true,
					"headline":    "{txt:msx-white-soft:dic:label:version|Version} " + Vers,
					"titleHeader": "", "titleFooter": "{ico:http}{tb}{txt:msx-white:" + r.Host +
						"}{br}{ico:hardware}{tb}{txt:msx-white:" + runtime.GOOS + "/" + runtime.GOARCH +
						"}{br}{ico:web}{tb}{txt:msx-white:https://github.com/" + gitRepo + "}",
					"live": plistObj{"type": "schedule", "from": ts, "to": ts, "titleHeader": "{ico:timer}{tb}{txt:msx-white:overflow:text:dhms}"}},
				{"type": "space", "layout": "0,2,12,1", "text": "{txt:msx-white:" + Name + "} {dix:About}is a software for playing user's content and developing user's plugins.{br}It does not provide any video/audio content by itself!", "alignment": "center"},
				{"id": "update", "type": "control", "layout": "0,3,6,1", "label": "{dic:CheckUp|Check updates}", "icon": "system-update-alt", "action": "execute:fetch:" + u + "/update"},
				{"type": "control", "layout": "0,4,6,1", "icon": "smart-display", "label": "{dic:label:player|Player}:", "extensionLabel": pls, "action": "execute:" + u + "/settings?id={ID}", "data": 1},
				{"type": "control", "layout": "0,5,6,1", "label": "{dic:Files|List of files}:", "icon": "format-list-bulleted", "extensionLabel": lst, "action": "execute:" + u + "/settings?id={ID}", "data": 2},
				{"id": "dic", "type": "control", "layout": "6,3,6,1", "icon": "language", "label": "{dic:Language|Language}:", "extensionLabel": "default", "action": "panel:" + u + "/msx/dictionary", "live": map[string]string{"type": "setup", "action": "execute:service:info:dictionary:" + u + "/msx/dictionary"}},
				{"type": "control", "layout": "6,4,6,1", "icon": "bolt", "label": "TorrServer:", "extensionLabel": ta, "action": "execute:" + u + "/settings", "data": nil},
				{"type": "control", "layout": "6,5,6,1", "label": "{dic:label:application|Application}", "icon": "monitor", "extensionIcon": "menu-open", "action": "dialog:application"},
			}}}}})
		l.write(w)
	})
}
