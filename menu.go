package main

import (
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
		w.Write([]byte(`{"name":"` + Name + `","version":"` + Vers + `","reference":"http://` + r.Host + `/msx/menu` + startFocus + `","dictionary":"http://` + r.Host + `/msx/dic.json"}`))
		startFocus = ""
	})
	http.HandleFunc("/msx/menu", func(w http.ResponseWriter, r *http.Request) {
		if stg.TorrServer == "" {
			t := r.Host[:strings.LastIndexByte(r.Host, ':')+1] + "8090"
			if _, e := checkTorr(t); e == nil {
				stg.TorrServer = t
			}
		}
		u := "http://" + r.Host
		l := &plist{Logo: u + "/logotype.svg", Menu: []map[string]interface{}{{"icon": "history", "label": "{dic:history}", "data": u + "/msx/history"}}}
		if _, e := os.Stat(pthVideo); !os.IsNotExist(e) {
			l.Menu = append(l.Menu, map[string]interface{}{"icon": "video-library", "label": "{dic:video}", "data": u + "/msx/" + pthVideo + "/"})
		}
		if _, e := os.Stat(pthMusic); !os.IsNotExist(e) {
			l.Menu = append(l.Menu, map[string]interface{}{"icon": "library-music", "label": "{dic:music}", "data": u + "/msx/" + pthMusic + "/"})
		}
		ri, ta, tv := "msx-white:check-box", "{dic:label:none|None}", r.Host[:strings.LastIndexByte(r.Host, '.')+1]
		if !stg.Russian {
			ri += "-outline-blank"
		}
		if stg.TorrServer != "" {
			ta, tv = "{col:msx-white}"+stg.TorrServer, stg.TorrServer
			l.Menu = append(l.Menu, map[string]interface{}{"icon": "bolt", "label": "{dic:torrents}", "data": u + "/msx/torr"})
		}
		l.Menu = append(l.Menu, map[string]interface{}{"type": "separator"})
		if ps, e := plugsInfo(); e == nil {
			ml := len(l.Menu)
			for _, p := range ps {
				if p.Error == nil {
					m := map[string]interface{}{"label": p.Label}
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
			if len(l.Menu) > ml {
				l.Menu = append(l.Menu, map[string]interface{}{"type": "separator"})
			}
		}
		ts := started.UnixMilli()
		l.Menu = append(l.Menu, map[string]interface{}{"id": "stg", "icon": "settings", "label": "{dic:label:settings|Settings}",
			"data": map[string][]map[string][]map[string]interface{}{"pages": {{"items": {
				{"type": "space", "layout": "0,0,12,2", "image": u + "/logotype.svg", "imageFiller": "height", "imageWidth": 7, "imagePreload": true,
					"headline":    "{txt:msx-white-soft:dic:label:version|Version} " + Vers,
					"titleHeader": "", "titleFooter": "{ico:http}{tb}{txt:msx-white:" + r.Host +
						"}{br}{ico:hardware}{tb}{txt:msx-white:" + runtime.GOOS + "/" + runtime.GOARCH +
						"}{br}{ico:web}{tb}{txt:msx-white:https://github.com/" + gitRepo + "}",
					"live": map[string]interface{}{"type": "schedule", "from": ts, "to": ts, "titleHeader": "{ico:timer}{tb}{txt:msx-white:overflow:text:dhms}"}},
				{"type": "space", "layout": "0,2,12,1", "text": "{dic:about}"},
				{"type": "control", "layout": "0,3,6,1", "icon": "bolt", "label": "TorrServer", "extensionLabel": ta, "action": "execute:" + u + "/msx/input?id={ID}&addr",
					"data": map[string]string{"action": "execute:" + u + "/settings", "headline": "{dic:addrInput} TorrServer", "value": tv, "extension": "<IP>:8090"}},
				{"id": "dic", "type": "control", "layout": "6,3,6,1", "icon": "translate", "label": "{dic:rus}", "extensionIcon": ri, "action": "execute:" + u + "/settings", "data": !stg.Russian},
				{"type": "button", "icon": "system-update-alt", "label": "{dic:update}{br}{br}{br}" + Name, "layout": "0,4,3,2", "action": "execute:" + u + "/update", "enable": false},
				{"id": "rst", "type": "button", "icon": "restart-alt", "label": "{dic:label:restart|Restart}{br}{br}{br}{br}" + Name, "layout": "3,4,3,2", "action": "execute:" + u + "/restart"},
				{"type": "button", "icon": "refresh", "label": "{dic:label:reload|Reload}{br}{br}{br}{br}{dic:label:application|Application}", "layout": "6,4,3,2", "action": "reload"},
				{"type": "button", "icon": "logout", "label": "{dic:label:exit|Exit}{br}{br}{br}{br}{dic:label:application|Application}", "layout": "9,4,3,2", "action": "exit"},
			}}}}})
		l.write(w)
	})
}
