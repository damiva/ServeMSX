package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func init() {
	http.HandleFunc("/msx/recent", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.Method == "POST" {
			var v struct {
				Video struct {
					Info struct {
						Index           int
						ID, Label, Type string
					}
					Data struct{ Position, Duration int64 }
				}
			}
			check(json.NewDecoder(r.Body).Decode(&v))
			a := "[]"
			if l := q.Get("link"); l != "" {
				a = recentAdd(r, v.Video.Info.ID, v.Video.Info.Index, recentItem{
					v.Video.Info.Label,
					l, q.Get("image"),
					v.Video.Data.Position, v.Video.Data.Duration,
					time.Now().Unix(), v.Video.Info.Type == "video",
				})
			} else if h := q.Get("hash"); h != "" {
				a = recentUp(h, v.Video.Data.Position, v.Video.Data.Duration)
			}
			svcAnswer(w, a, nil)
		} else if d := q.Get("del"); d != "" {
			delete(stg.Recent, d)
			stg.toSave = true
			svcAnswer(w, "reload:content", nil)
		} else {
			recentList(w, r)
		}
	})
}
func recentAdd(r *http.Request, id string, ix int, it recentItem) string {
	act := "["
	if id != "" {
		it.Ref += ">" + id
	} else {
		it.Ref += ">index:" + strconv.Itoa(ix)
	}
	it.Ref += ">execute"
	h := sha1.New()
	h.Write([]byte(it.Ref))
	id = hex.EncodeToString(h.Sum(nil))
	if i, o := stg.Recent[id]; o && i.Dur == it.Dur {
		act += "resume:position:" + strconv.FormatInt(i.Pos, 10) + "|"
	}
	stg.Recent[id] = it
	stg.toSave = true
	return act + "player:ticking:restart|trigger:10t:execute:service:video:http://" + r.Host + "/msx/recent?hash=" + id + "]"
}
func recentUp(h string, p, d int64) string {
	if i := stg.Recent[h]; i.Ref == "" {
		panic(404)
	} else if d-p > 12 {
		i.Pos, i.Dur, i.Lts = p, d, time.Now().Unix()
		stg.Recent[h] = i
	} else {
		delete(stg.Recent, h)
	}
	stg.toSave = true
	return "player:ticking:restart"
}
func recentList(w http.ResponseWriter, r *http.Request) {
	p := &plist{Type: "List", Template: map[string]interface{}{"type": "control", "layout": "0,0,12,1"}}
	p.opts(r, true, false)
	for k, v := range stg.Recent {
		i := plistObj{"id": k, "label": v.Lbl, "progress": float64(v.Pos) / float64(v.Dur), "action": "[" + v.Ref + "|resume:position:" + strconv.FormatInt(v.Pos, 10) + "]", "timestamp": v.Lts}
		if v.Vid {
			i["extensionIcon"] = "movie"
		} else {
			i["extensionIcon"] = "audiotrack"
		}
		if strings.HasPrefix(v.Img, "ico:") {
			i["icon"] = v.Img[4:]
		} else if v.Img != "" {
			i["image"] = v.Img
		}
		p.Items = append(p.Items, i)
	}
	sort.Slice(p.Items, func(i, j int) bool { return p.Items[i]["timestamp"].(int64) > p.Items[j]["timestamp"].(int64) })
	p.write(w)
}
