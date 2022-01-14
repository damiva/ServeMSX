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
	Header   plistObj                `json:"header,omitempty"`
	Ext      string                  `json:"extension,omitempty"`
	Logo     string                  `json:"logo,omitempty"`
	Template plistObj                `json:"template,omitempty"`
	Items    []plistObj              `json:"items,omitempty"`
	Pages    []map[string][]plistObj `json:"pages,omitempty"`
	Menu     []plistObj              `json:"menu,omitempty"`
}

func (p *plist) write(w io.Writer) error {
	j := json.NewEncoder(w)
	j.SetIndent("", "  ")
	return j.Encode(p)
}
func mediaList(r *http.Request, hdr, ext string, vid bool) *plist {
	i, ps := "movie", playerProp(r.Host, r.FormValue("id"), false)
	if !vid {
		i = "audiotrack"
	}
	ps["resume:key"] = "url"
	return &plist{
		Type: "list", Head: hdr, Ext: ext,
		Template: plistObj{
			"icon":       "msx-white-soft:" + i,
			"type":       "control",
			"layout":     "0,0,12,1",
			"progress":   -1,
			"live":       map[string]string{"type": "playback", "action": "player:show"},
			"properties": ps,
		},
	}
}
func svcAnswer(w http.ResponseWriter, act string, dat interface{}) {
	json.NewEncoder(w).Encode(map[string]map[string]interface{}{"response": {"status": 200, "data": map[string]interface{}{"action": act, "data": dat}}})
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
func playerProp(host, id string, live bool) (ps map[string]string) {
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
	return
}
