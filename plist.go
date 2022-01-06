package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type plistObj map[string]interface{}
type plist struct {
	Type     string                  `json:"type,omitempty"`
	Reuse    bool                    `json:"reuse"`
	Cache    bool                    `json:"cache"`
	Restore  bool                    `json:"restore"`
	Compress bool                    `json:"compress,omitempty"`
	Action   string                  `json:"action,omitempty"`
	Head     string                  `json:"headline,omitempty"`
	Ext      string                  `json:"extension,omitempty"`
	Logo     string                  `json:"logo,omitempty"`
	Template plistObj                `json:"template,omitempty"`
	Items    []plistObj              `json:"items,omitempty"`
	Pages    []map[string][]plistObj `json:"pages,omitempty"`
	Menu     []plistObj              `json:"menu,omitempty"`
}

func (p *plist) opts(r *http.Request, del, drop bool) {
	var its []plistObj
	cap := "{dic:label:menu|Menu}:"
	if del {
		cap += "{tb}{ico:msx-red:stop} {dic:Remove|Remove}"
		its = append(its, plistObj{"key": "red", "icon": "msx-red:stop", "label": "{dic:Remove|Remove}", "action": "execute:service:fetch:http://" + r.Host + r.URL.Path + "?del={context:id}"})
	}
	if drop {
		cap += "{tb}{ico:msx-yellow:stop} {dic:Drop|Drop}"
		its = append(its, plistObj{"key": "yellow", "icon": "msx-yellow:stop", "label": "{dic:Drop|Drop}", "action": "execute:service:fetch:http://" + r.Host + r.URL.Path + "?drop={context:id}"})
	}
	if len(its) > 0 {
		if p.Template == nil {
			p.Template = make(plistObj)
		}
		p.Template["options"] = map[string]interface{}{
			"headline": "{dic:label:menu|Menu}",
			"caption":  cap,
			"template": map[string]interface{}{"enumerate": false, "type": "control", "layout": "0,0,8,1"},
			"items":    its,
		}
	}
}
func (p *plist) write(w io.Writer) error {
	j := json.NewEncoder(w)
	j.SetIndent("", "  ")
	return j.Encode(p)
}
func svcAnswer(w http.ResponseWriter, act string, dat interface{}) {
	json.NewEncoder(w).Encode(map[string]map[string]interface{}{"response": {"status": 200, "data": map[string]interface{}{"action": act, "data": dat}}})
}
func plistMedia(r *http.Request, hdr, ico string) *plist {
	img, lnk := "", "content:http://"+r.Host+r.URL.EscapedPath()
	switch r.URL.Path[5] {
	case 'v':
		img = "ico:video-library"
	case 'm':
		img = "ico:library-music"
	case 't':
		img = "ico:bolt"
		lnk += "?link=" + url.QueryEscape(r.FormValue("link"))
	}
	ps := playerProp(r.Host, r.URL.Query().Get("id"), lnk, img)
	ps["resume:key"] = "url"
	p := &plist{Type: "list", Head: hdr, Template: map[string]interface{}{"layout": "0,0,12,1", "type": "control", "progress": -1, "properties": ps, "live": map[string]string{"type": "playback", "action": "player:show"}}}
	if ico != "" {
		p.Template["icon"] = ico
	}
	return p
}
func playerURL(id, ur string, iv bool) string {
	if stg.HTML5X[id] {
		ur = "plugin:http://msx.benzac.de/plugins/html5x.html?url=" + url.QueryEscape(ur)
	}
	if iv {
		ur = "video:" + ur
	} else {
		ur = "audio:" + ur
	}
	return ur
}
func playerProp(host, id, ref, img string) (ps map[string]string) {
	if ref == "*" {
		ps = map[string]string{
			"button:play_pause:display": "false",
			"button:forward:display":    "false",
			"button:rewind:display":     "false",
			"button:restart:display":    "false",
			"button:speed:display":      "false",
			"label:position":            "...",
			"label:duration":            "...",
			"button:content:icon":       "settings",
		}
	} else {
		ps = map[string]string{"button:content:icon": "settings"}
	}
	switch {
	case stg.HTML5X[id]:
		ps["button:content:action"] = "panel:request:player:options"
	case clients[id].Player == "tizen":
		ps["button:content:action"] = "content:request:interaction:init@http://msx.benzac.de/interaction/tizen.html"
	case clients[id].Platform == "netcast":
		ps["button:content:action"] = "system:netcast:menu"
	default:
		delete(ps, "button:content:icon")
	}
	if len(ref) > 1 {
		ref = host + "/msx/recent?link=" + url.QueryEscape(ref) + "&image=" + url.QueryEscape(img)
		ps["trigger:start"] = "execute:service:video:http://" + ref
	}
	return
}
