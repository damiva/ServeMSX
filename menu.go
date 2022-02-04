package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const pthVideoWall = "/msx/videowall"

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
		l := &plist{Logo: u + "/logotype.png", Flag: Name, Ready: plistObj{"action": "execute:service:video:data:http://" + r.Host + pthVideoWall}, Menu: []plistObj{}}
		for i, f := range [4][2]string{{pthMarks, "bookmarks"}, {pthVideo, "video-library"}, {pthMusic, "library-music"}, {pthPhoto, "photo-library"}} {
			if _, e := os.Stat(f[0]); !os.IsNotExist(e) {
				c, s := "folder-open", "/?id={ID}"
				if i == 0 {
					c, s = "cloud-queue", "?id={ID}"
				} else if i == 3 {
					s += "&height={HEIGHT}"
				}
				l.Menu = append(l.Menu, plistObj{"icon": f[1], "extensionIcon": c, "label": "{dix:" + f[0] + "}My " + f[0], "data": u + "/msx/" + f[0] + s})
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
							m["extensionIcon"] = "cloud-queue"
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
		hard := runtime.GOOS + "/" + runtime.GOARCH
		if stg.FFmpeg != "" {
			hard += " {txt:msx-yellow:+ ffmpeg}"
		}
		walls := []plistObj{{"offset": "0,0,5,-1", "label": "{dic:label:no|No}", "data": 100, "enumerate": false}, {"type": "space"}}
		for i := 1; i < 7; i++ {
			ii := strconv.Itoa(i)
			walls = append(walls, plistObj{"image": "http://msx.benzac.de/media/thumbs/atmos" + ii + ".jpg", "label": ii, "data": 100 + i})
		}
		wall := "{dic:label:noту|None}"
		if stg.VideoWall > 0 && stg.VideoWall < 7 {
			wall = "{col:msx-white}" + strconv.Itoa(stg.VideoWall)
		}
		sa := "execute:" + u + "/settings?id={ID}"
		l.Menu = append(l.Menu, plistObj{"id": "stg", "icon": "settings", "label": "{dic:label:settings|Settings}",
			"data": plistObj{"extension": "{ico:msx-white:settings}", "compress": true,
				"pages": []map[string][]plistObj{{"items": {
					{"type": "space", "layout": "0,0,16,3", "image": u + "/logotype.png", "imageFiller": "width-top", "imageWidth": 10, "imagePreload": true,
						"headline":    "{txt:msx-white-soft:dic:label:version|Version} " + Vers,
						"titleHeader": "", "titleFooter": "{ico:http}{tb}{col:msx-white}" + r.Host +
							"{br}{ico:msx-white-soft:hardware}{tb}" + hard +
							"{br}{br}{ico:msx-white-soft:web}{tb}https://github.com/" + gitRepo,
						"live": plistObj{"type": "schedule", "from": ts, "to": ts, "titleHeader": "{ico:timer}{tb}{txt:msx-white:overflow:text:dhms}"}},
					{"type": "space", "layout": "0,3,16,1", "text": "{txt:msx-white:" + Name + "} {dix:About}is a software for playing user's content and developing user's plugins.{br}It does not provide any video/audio content by itself!", "alignment": "center"},

					{"type": "control", "layout": "0,4,8,1", "label": "{dic:CheckUp|Check updates}", "icon": "system-update-alt", "action": "execute:fetch:" + u + "/update"},
					{"type": "control", "layout": "0,5,8,1", "icon": "smart-display", "label": "{dic:label:player|Player}:", "extensionLabel": stg.switcher(r, cHTML5X, pls, "html5x"), "action": sa, "data": cHTML5X},
					{"type": "control", "layout": "0,6,8,1", "label": "{dic:Files|List of files}:", "icon": "format-list-bulleted", "extensionLabel": stg.switcher(r, cCompressed, "{dic:Default|default}", "{dic:Compress|compressed}"), "action": sa, "data": cCompressed},
					{"type": "control", "layout": "0,7,8,1", "icon": "photo-library", "label": "{dic:Slide|Photo size}:", "extensionLabel": stg.switcher(r, cPhotoScale, "{dic:Origin|as is}", "{dic:Scale|as screen}"), "action": sa, "data": cPhotoScale, "enable": stg.FFmpeg != ""},
					{"id": "dic", "type": "control", "layout": "8,4,8,1", "icon": "language", "label": "{dic:Language|Language}:", "extensionLabel": "default", "action": "panel:" + u + "/msx/dictionary", "live": map[string]string{"type": "setup", "action": "execute:service:info:dictionary:" + u + "/msx/dictionary"}},
					{"type": "control", "layout": "8,5,8,1", "icon": "bolt", "label": "TorrServer:", "extensionLabel": ta, "action": sa, "data": nil},
					{"type": "control", "layout": "8,6,8,1", "label": "{dic:VideoWall|Video wallpaper}:", "icon": "wallpaper", "extensionLabel": wall, "action": "panel:data", "data": plistObj{
						"type": "list", "headline": "{dic:VideoWall|Video wallpaper}:", "extension": "© Benjamin Zachey", "compress": true, "template": plistObj{"layout": "0,0,5,2", "imageFiller": "cover", "action": "[quiet|" + sa + "]"}, "items": walls,
					}},
					{"type": "control", "layout": "8,7,8,1", "label": "{dic:label:application|Application}", "icon": "monitor", "extensionIcon": "menu-open", "action": "dialog:application"},
				}}}}})
		l.write(w)
	})
	http.HandleFunc(pthVideoWall, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case stg.VideoWall < 1 || stg.VideoWall > 6:
			svcAnswer(w, "[]", nil)
		case r.Method == "POST":
			var i struct {
				Video struct{ Data struct{ State int } }
			}
			if e := json.NewDecoder(r.Body).Decode(&i); e != nil {
				panic(e)
			} else if i.Video.Data.State > 1 {
				svcAnswer(w, "[]", nil)
				return
			}
			fallthrough
		default:
			i := strconv.Itoa(stg.VideoWall)
			svcAnswer(w,
				"video:auto:plugin:http://msx.benzac.de/plugins/background.html?url=http%3A%2F%2Fmsx.benzac.de%2Fmedia%2Fatmos"+i+".mp4",
				plistObj{
					"playerLabel": "Background Video " + i,
					"properties": plistObj{
						"control:type":        "extended",
						"control:reuse":       "play",
						"control:transparent": true,
						"label:duration":      "{VALUE} {ico:repeat}",
					},
				},
			)
		}
	})
}
