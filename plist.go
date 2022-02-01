package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type plistObj map[string]interface{}
type plist struct {
	Type       string                  `json:"type,omitempty"`
	Reuse      bool                    `json:"reuse"`
	Cache      bool                    `json:"cache"`
	Restore    bool                    `json:"restore"`
	Compress   bool                    `json:"compress,omitempty"`
	Background string                  `json:"background,omitempty"`
	Action     string                  `json:"action,omitempty"`
	Flag       string                  `json:"flag,omitempty"`
	Head       string                  `json:"headline,omitempty"`
	Ext        string                  `json:"extension,omitempty"`
	Logo       string                  `json:"logo,omitempty"`
	Header     plistObj                `json:"header,omitempty"`
	Options    plistObj                `json:"options,omitempty"`
	Template   plistObj                `json:"template,omitempty"`
	Items      []plistObj              `json:"items,omitempty"`
	Pages      []map[string][]plistObj `json:"pages,omitempty"`
	Menu       []plistObj              `json:"menu,omitempty"`
}

func (p *plist) write(w io.Writer) error {
	j := json.NewEncoder(w)
	j.SetIndent("", "  ")
	return j.Encode(p)
}
func mediaList(r *http.Request, hdr, ext, ico string, opt []plistObj, optEach, cover, exted bool) *plist {
	id := r.FormValue("id")
	if ico != "" {
		ico = "msx-white-soft:" + ico
	}
	ps, cmp, lay := playerProp(r.Host, id, false, exted), id != "" && stg.Clients[id]&cCompressed != 0, "0,0,12,1"
	if cmp {
		lay = "0,0,16,1"
	}
	ps["resume:key"] = "url"
	if hdr != "" && exted {
		ps["info:text"] = hdr
	}
	if cover {
		ps["trigger:load"] = "execute:fetch:{context:cover}"
	}
	opt = append(opt, nil)
	liv := plistObj{"type": "playback", "action": "player:show"}
	rtn := &plist{
		Type: "list", Head: hdr, Ext: ext, Compress: cmp,
		Template: plistObj{"icon": ico, "type": "control", "layout": lay, "progress": -1, "live": liv, "properties": ps},
	}
	if optEach {
		rtn.Template["options"] = options(opt...)
	} else {
		rtn.Options = options(opt...)
	}
	return rtn
}
func svcAnswer(w http.ResponseWriter, act string, dat interface{}) {
	json.NewEncoder(w).Encode(map[string]map[string]interface{}{"response": {"status": 200, "data": map[string]interface{}{"action": act, "data": dat}}})
}
func playerURL(id, ur string, iv bool) string {
	if stg.Clients[id]&cHTML5X != 0 {
		ur = "plugin:http://msx.benzac.de/plugins/html5x.html?url=" + url.QueryEscape(ur)
	}
	if iv {
		ur = "video:" + ur
	} else {
		ur = "audio:" + ur
	}
	return ur
}
func playerProp(host, id string, live, ext bool) (ps map[string]string) {
	if live {
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
	if ext {
		ps["control:type"] = "extended"
	}
	switch {
	case stg.Clients[id]&cHTML5X != 0:
		ps["button:content:action"] = "panel:request:player:options"
	case clients[id].Player == "tizen":
		ps["button:content:action"] = "content:request:interaction:init@http://msx.benzac.de/interaction/tizen.html"
	case clients[id].Platform == "netcast":
		ps["button:content:action"] = "system:netcast:menu"
	default:
		delete(ps, "button:content:icon")
	}
	return
}
func options(opts ...plistObj) plistObj {
	cap := "{dic:caption:options|Options}:"
	for i := 0; i < len(opts); i++ {
		if opts[i] == nil {
			opts[i] = plistObj{"key": "yellow", "label": "{dic:Up|Up}", "action": "[cleanup|focus:index:0]"}
		}
		cap += "{tb}{ico:msx-" + opts[i]["key"].(string) + ":stop} " + opts[i]["label"].(string)
		opts[i]["icon"] = "msx-" + opts[i]["key"].(string) + ":stop"
	}
	if len(opts) > 0 {
		opts = append(opts, plistObj{"type": "space"})
	}
	opts = append(opts, plistObj{"icon": "menu", "label": "{dic:caption:menu|Menu}", "action": "menu"})
	return plistObj{
		"headline": "{dic:caption:options|Options}", "caption": cap, "items": opts,
		"template": plistObj{"enumerate": false, "type": "control", "layout": "0,0,8,1"},
	}
}
