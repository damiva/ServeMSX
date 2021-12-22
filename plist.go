package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type plist struct {
	Type     string                                `json:"type,omitempty"`
	Reuse    bool                                  `json:"reuse"`
	Cache    bool                                  `json:"cache"`
	Restore  bool                                  `json:"restore"`
	Compress bool                                  `json:"compress,omitempty"`
	Head     string                                `json:"headline,omitempty"`
	Ext      string                                `json:"extension,omitempty"`
	Logo     string                                `json:"logo,omitempty"`
	Template map[string]interface{}                `json:"template,omitempty"`
	Items    []map[string]string                   `json:"items,omitempty"`
	Pages    []map[string][]map[string]interface{} `json:"pages,omitempty"`
	Menu     []map[string]interface{}              `json:"menu,omitempty"`
}

func (p *plist) opts(r *http.Request, hst bool) {
	cap := "{dic:opt}:{tb}{ico:msx-red:stop} {dic:rmv}"
	its := []map[string]string{{"key": "red", "icon": "msx-red:stop", "label": "{dic:rmv}", "action": "execute:http://" + r.Host + r.URL.Path + "?del={context:id}"}}
	if hst {
		cap += "{tb}{ico:msx-green:stop} {dic:goto}"
		its = append(its, map[string]string{"key": "green", "icon": "msx-green:stop", "label": "{dic:goto}", "action": "execute:http://" + r.Host + r.URL.Path + "?goto={context:id}"})
	} else {
		cap += "{tb}{ico:msx-yellow:stop} {dic:drop}"
		its = append(its, map[string]string{"key": "yellow", "icon": "msx-yellow:stop", "label": "{dic:drop}", "action": "execute:http://" + r.Host + r.URL.Path + "?drop={context:id}"})
	}
	if p.Template == nil {
		p.Template = make(map[string]interface{})
	}
	p.Template["options"] = map[string]interface{}{
		"headline": "{dic:opt}",
		"caption":  cap,
		"template": map[string]interface{}{"enumerate": false, "type": "control", "layout": "0,0,8,1"},
		"items":    its,
	}
}
func (p *plist) mediaList(r *http.Request, ico string, addhst bool) {
	p.Type = "list"
	ps := map[string]string{"resume:key": "url"}
	if addhst {
		ps["trigger:start"] = "execute:video:info:http://" + r.Host + "/msx/history?src=" + url.QueryEscape("content:http://"+r.Host) + r.RequestURI
	}
	if p.Template == nil {
		p.Template = make(map[string]interface{})
	}
	p.Template["type"], p.Template["layout"], p.Template["progress"], p.Template["properties"] = "control", "0,0,12,1", -1, ps
	p.Template["live"] = map[string]string{"type": "playback", "action": "player:show"}
	if ico != "" {
		p.Template["icon"] = ico
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
