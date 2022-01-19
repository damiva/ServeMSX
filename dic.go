package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
)

const pthDic = "dictionary.json"

func init() {
	http.HandleFunc("/msx/"+pthDic, func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, pthDic) })
	http.HandleFunc("/msx/dictionary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			type info struct{ Dictionary struct{ Name string } }
			var (
				d struct {
					Data string
					Info *info
				}
				act = "reload"
				dat interface{}
			)
			check(json.NewDecoder(r.Body).Decode(&d))
			startFocus = ">stg>dic"
			if d.Info != nil {
				startFocus = ""
				if d.Info.Dictionary.Name != "" {
					act = "update:content:dic"
					if dn := strings.SplitN(d.Info.Dictionary.Name, "/", 2); len(dn) > 1 {
						d.Info.Dictionary.Name = dn[1]
					}
					dat = map[string]string{"extensionLabel": "{txt:msx-white:" + d.Info.Dictionary.Name + "}"}
				} else {
					act = "[]"
				}
			} else if d.Data != "" {
				check(download(d.Data, pthDic, true))
			} else if e := os.Remove(pthDic); e != nil && !os.IsNotExist(e) {
				panic(e)
			}
			svcAnswer(w, act, dat)
		} else {
			ds := []plistObj{{"label": "default", "data": ""}}
			i, e := gitRelease("", Vers)
			check(e)
			for _, a := range i.Assets {
				if strings.HasSuffix(a.Name, ".json.gz") {
					ds = append(ds, plistObj{"label": strings.TrimSuffix(a.Name, ".json.gz"), "data": a.Browser_download_url})
				}
			}
			(&plist{
				Type:     "list",
				Head:     "{dic:label:language|Language}:",
				Template: plistObj{"type": "button", "layout": "0,0,8,1", "action": "execute:http://" + r.Host + r.URL.Path},
				Items:    ds,
			}).write(w)
		}
	})
}
func getDic() (v, n, k string, e error) {
	var (
		d struct{ Name, Version, Keyboard string }
		f *os.File
	)
	if f, e = os.Open(pthDic); e == nil {
		if e = json.NewDecoder(f).Decode(&d); e == nil {
			v, n, k = d.Version, strings.Split(d.Name, "/")[0], d.Keyboard
		} else {
			e = errors.New("Decoding " + pthDic + " error: " + e.Error())
		}
		f.Close()
	} else if os.IsExist(e) {
		e = nil
	}
	return
}
