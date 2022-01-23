package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
)

const pthMarks = "webmarks"

func webmarks() [][]byte {
	b, e := os.ReadFile(pthMarks)
	check(e)
	return bytes.Split(b, []byte{'\n'})
}
func webmark(i string, d int, l []byte) plistObj {
	s := bytes.SplitN(l[1:], []byte{'#'}, 2)
	u := string(s[0])
	p := plistObj{"id": strconv.Itoa(d), "label": u}
	if len(s) > 1 {
		p["label"] = string(s[1])
	}
	p["playerLabel"] = p["label"]
	switch l[0] {
	case 'Y':
		p["action"], p["icon"], p["group"] = "video:plugin:http://msx.benzac.de/plugins/youtube.html?id="+u, "msx-red:smart-display", "YouTube"
	case 'H':
		p["icon"], p["group"] = "msx-white-soft:cast", "{dic:Stream|Video-stream}"
		fallthrough
	default:
		p["action"] = playerURL(i, u, true)
	}
	l = nil
	return p
}
func init() {
	http.HandleFunc("/msx/"+pthMarks, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var d struct{ Data interface{} }
			check(json.NewDecoder(r.Body).Decode(&d))
			switch v := d.Data.(type) {
			case string:
				if i, e := strconv.Atoi(v); e != nil {
					if i < 0 {
						check(os.WriteFile(pthMarks, []byte{}, 0666))
					} else if b := webmarks(); i < len(b) {
						b = append(b[:i], b[i+1:]...)
						check(os.WriteFile(pthMarks, append(bytes.Join(b, []byte{'\n'}), '\n'), 0666))
						svcAnswer(w, "reload:content", nil)
					} else {
						svcAnswer(w, "[]", nil)
					}
				} else if len(v) < 2 {
					w.Write([]byte("false"))
				} else {
					f, e := os.OpenFile(pthMarks, os.O_APPEND|os.O_CREATE|os.O_RDONLY, 0666)
					check(e)
					_, e = f.WriteString(v + "\n")
					f.Close()
					check(e)
					w.Write([]byte("true"))
				}
			case bool:
				if v {
					check(os.WriteFile(pthMarks, []byte{}, 0666))
				}
				w.Write(append([]byte{'"'}, append([]byte(stg.TorrServer), '"')...))
			default:
				panic(400)
			}
		} else {
			ls, id, pl := webmarks(), r.FormValue("id"), mediaList(r, "", "{ico:msx-white:bookmarks}", "movie", []plistObj{{"key": "red", "label": "{dic:Remove|Remove}", "action": "execute:http://" + r.Host + r.URL.Path, "data": "{context:id}"}}, true)
			if stg.Clients[id]&cMarksLIFO != 0 {
				for i := len(ls) - 1; i >= 0; i-- {
					pl.Items = append(pl.Items, webmark(id, i, ls[i]))
				}
			} else {
				for i := 0; i < len(ls); i++ {
					pl.Items = append(pl.Items, webmark(id, i, ls[i]))
				}
			}
			pl.Template["group"] = "{dic:label:video|Video}"
			cb, ly := "check-box-outline-blank", []string{"2", "6"}
			if stg.Clients[id]&cMarksLIFO != 0 {
				cb = "msx-white:check-box"
			}
			if stg.Clients[id]&cCompressed != 0 {
				ly = []string{"4", "8"}
			}
			pl.Header = plistObj{"items": []plistObj{
				{"type": "control", "icon": "sort", "label": "{dic:LIFO|Newest first}", "extensionIcon": cb, "layout": ly[0] + ",0,4,1", "action": "execute:http://" + r.Host + "/msx/settings?id={ID}", "data": cMarksLIFO},
				{"type": "control", "icon": "clear", "label": "{dic:label:clear|Clear}", "layout": ly[1] + ",0,4,1", "action": "execute:http://" + r.Host + r.URL.Path, "data": "-1"},
			}}
		}
	})
}
