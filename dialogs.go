package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type dialogStg struct{ Action, Headline, Extension, Value string }

var (
	keyboards = map[bool]string{false: "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ@-:.,/?+();!&\"", true: ""}
	boxText   = make(map[string]dialogStg)
	boxLang   = make(map[string]bool)
)

func init() {
	http.HandleFunc("/msx/input", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		if id == "" {
			id = "*"
		}
		if r.Method == "POST" {
			var key struct{ Data string }
			var stg struct{ Data dialogStg }
			b, e := ioutil.ReadAll(r.Body)
			check(e)
			if json.Unmarshal(b, &key) == nil {
				inputKey(w, id, key.Data)
			} else if json.Unmarshal(b, &stg) == nil {
				if _, k, e := getDic(); e == nil {
					keyboards[true] = k
				} else {
					log.Println(e)
				}
				boxText[id] = stg.Data
				if q.Has("addr") {
					id += "&addr"
				}
				svcAnswer(w, "panel:http://"+r.Host+r.URL.Path+"?id="+id, nil)
			} else {
				panic(400)
			}
		} else {
			inputKbd(w, r, id, q.Has("addr"))
		}
	})
	http.HandleFunc("/msx/dialog", func(w http.ResponseWriter, r *http.Request) {
		var dat struct{ Data dialogStg }
		if r.Method != "POST" {
			panic(400)
		} else if e := json.NewDecoder(r.Body).Decode(&dat); e != nil {
			panic("Parsing dialog data error: " + e.Error())
		}
		svcAnswer(w, "panel:data", &plist{Head: dat.Data.Headline, Ext: dat.Data.Extension, Pages: []map[string][]plistObj{{"items": {
			{"type": "space", "headline": dat.Data.Value, "layout": "0,0,8,6", "alignment": "center"},
			{"type": "button", "icon": "done", "layout": "4,5,2,1", "action": dat.Data.Action, "display": dat.Data.Action != ""},
			{"type": "button", "icon": "close", "layout": "6,5,2,1", "action": "back"},
		}}}})
	})
}
func inputKbd(w http.ResponseWriter, r *http.Request, id string, ka bool) {
	bx, at := boxText[id], "execute:service:http://"+r.Host+r.URL.Path+"?id="+id
	ks := []plistObj{{"id": "txt", "type": "space", "label": bx.Value, "color": "msx-black-soft", "layout": "0,0,10,1", "offset": "0,0.3,0,0.3", "compress": false}}
	i := 0
	for _, k := range keyboards[boxLang[id] && !ka] {
		y := i / 10
		x := i - y*10
		uk := string(k)
		lk := strings.ToLower(uk)
		kb := map[string]interface{}{"type": "button", "label": uk, "key": uk + "|" + lk, "action": at, "data": uk, "layout": strconv.Itoa(x) + "," + strconv.Itoa(y+2) + ",1,1"}
		if ka && i > 39 {
			kb["enable"] = false
		}
		ks = append(ks, kb)
		i++
	}
	for i, k := range [][4]string{
		{"backspace", "red|delete", "<del>", "fast-rewind"},
		{"clear", "home", "<clr>", "skip-previous"},
		{"space-bar", "yellow|space|insert", " ", "fast-forward"},
		{"language", "end|tab|caps_lock", "<lng>", "skip-next"},
		{"done", "green", "\n", ""},
	} {
		kb := map[string]interface{}{"type": "button", "icon": k[0], "iconSize": "small", "key": k[1], "action": at, "data": k[2], "layout": strconv.Itoa(i*2) + ",7,2,1"}
		if i%2 == 0 {
			kb["progress"], kb["progressColor"] = "1", strings.SplitN(k[1], "|", 2)[0]
		}
		if k[3] != "" {
			kb["titleFooter"] = "{ico:" + k[3] + "}"
		}
		ks = append(ks, kb)
	}
	if ka {
		ks[40]["progress"], ks[40]["progressColor"], ks[40]["key"] = 1, "msx-yellow", ".|yellow"
		ks[53]["enable"], ks[54]["enable"] = false, false
	} else if keyboards[true] == "" {
		ks[54]["enable"] = false
	}
	(&plist{Head: bx.Headline, Ext: bx.Extension, Compress: true, Pages: []map[string][]plistObj{{"items": ks}}}).write(w)
}
func inputKey(w http.ResponseWriter, id, key string) {
	b := boxText[id]
	switch key {
	case "\n":
		svcAnswer(w, strings.ReplaceAll(b.Action, "{INPUT}", url.QueryEscape(b.Value)), b.Value)
	case "<lng>":
		boxLang[id] = !boxLang[id]
		svcAnswer(w, "reload:panel", nil)
	case "<clr>":
		b.Value = ""
		fallthrough
	case "<del>":
		key = ""
		rs := []rune(b.Value)
		if l := len(rs); l > 0 {
			rs = rs[:len(rs)-1]
			b.Value = string(rs)
		}
		fallthrough
	default:
		b.Value += key
		boxText[id] = b
		if strings.HasSuffix(b.Value, " ") {
			b.Value = strings.TrimSuffix(b.Value, " ") + "{txt:msx-white-soft:_}"
		}
		svcAnswer(w, "update:panel:txt", map[string]string{"label": b.Value})
	}
}
