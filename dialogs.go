package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	keyboards = map[bool]string{false: "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ@-:.,/?+();!&\"", true: "1234567890АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ,.:?!-+"}
	boxTxt    = make(map[string][4]string)
	boxRus    = make(map[string]bool)
)

type dialogSets struct {
	Data struct{ Action, Headline, Extension, Value string }
}

func init() {
	http.HandleFunc("/msx/input", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		if id == "" {
			id = "*"
		}
		if r.Method == "POST" {
			var key struct{ Data string }
			var stg dialogSets
			b, e := ioutil.ReadAll(r.Body)
			check(e)
			if json.Unmarshal(b, &key) == nil {
				inputKey(w, id, key.Data)
			} else if json.Unmarshal(b, &stg) == nil {
				boxTxt[id] = [4]string{stg.Data.Action, stg.Data.Headline, stg.Data.Value, stg.Data.Extension}
				if q.Has("addr") {
					id += "&addr"
				}
				svcAnswer(w, "panel:http://"+r.Host+r.URL.Path+"?id="+id, nil)
			}
		} else {
			inputKbd(w, r, id, q.Has("addr"))
		}
	})
	http.HandleFunc("/msx/dialog", func(w http.ResponseWriter, r *http.Request) {
		var dat dialogSets
		if r.Method != "POST" {
			panic(400)
		} else if e := json.NewDecoder(r.Body).Decode(&dat); e != nil {
			panic("Parsing dialog data error: " + e.Error())
		}
		svcAnswer(w, "panel:data", &plist{Head: dat.Data.Headline, Ext: dat.Data.Extension, Pages: []map[string][]map[string]interface{}{{"items": {
			{"type": "space", "text": dat.Data.Value, "layout": "0,0,8,6"},
			{"type": "button", "icon": "done", "layout": "4,5,2,1", "action": dat.Data.Action, "display": dat.Data.Action != ""},
			{"type": "button", "icon": "close", "layout": "6,5,2,1", "action": "back"},
		}}}})
	})
}
func inputKbd(w http.ResponseWriter, r *http.Request, id string, ka bool) {
	bx, at := boxTxt[id], "execute:service:http://"+r.Host+r.URL.Path+"?id="+id
	ks := []map[string]interface{}{
		{"id": "txt", "type": "space", "label": bx[2], "color": "msx-black-soft", "layout": "0,0,10,1"},
	}
	i := 0
	for _, k := range keyboards[boxRus[id] && !ka] {
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
		{"space-bar", "space|yellow|insert", " ", "fast-forward"},
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
	}
	(&plist{Head: bx[1], Ext: bx[3], Compress: true, Pages: []map[string][]map[string]interface{}{{"items": ks}}}).write(w)
}
func inputKey(w http.ResponseWriter, id, key string) {
	b := boxTxt[id]
	switch key {
	case "\n":
		svcAnswer(w, b[0], b[2])
	case "<lng>":
		boxRus[id] = !boxRus[id]
		svcAnswer(w, "reload:panel", nil)
	case "<clr>":
		b[2] = ""
		fallthrough
	case "<del>":
		key = ""
		if l := len(b[2]); l > 0 {
			b[2] = b[2][:l-1]
		}
		fallthrough
	default:
		b[2] += key
		boxTxt[id] = b
		if strings.HasSuffix(b[2], " ") {
			b[2] = strings.TrimSuffix(b[2], " ") + "{txt:msx-white-soft:_}"
		}
		svcAnswer(w, "update:panel:txt", map[string]string{"label": b[2]})
	}
}
