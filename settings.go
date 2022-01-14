package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const pthSettings = "settings.json"

/*
type recentItem struct {
	Lbl, Ref, Img string
	Pos, Dur, Lts int64
	Vid           bool
}
*/
type settings struct {
	TorrServer string
	//CheckUpdate int64
	HTML5X map[string]bool
	//Plugins     map[string]string
	//Recent      map[string]recentItem
	//toSave      bool
}
type client struct{ Addr, Platform, Player, Vers string }

var (
	stg     = &settings{HTML5X: make(map[string]bool)}
	clients = make(map[string]client)
)

func init() {
	http.Handle("/settings", stg)
}
func (s *settings) save() (e error) {
	var f *os.File
	mutex.Lock()
	if f, e = os.Create(pthSettings); e == nil {
		j := json.NewEncoder(f)
		j.SetIndent("", "  ")
		if e = j.Encode(s); e != nil {
			e = errors.New("Encoding " + pthSettings + " error: " + e.Error())
		}
		f.Close()
	}
	mutex.Unlock()
	return
}
func (s *settings) load() (e error) {
	var f *os.File
	if f, e = os.Open(pthSettings); e == nil {
		if e = json.NewDecoder(f).Decode(s); e != nil {
			e = errors.New("Decoding " + pthSettings + " error: " + e.Error())
		}
		f.Close()
	} else if os.IsNotExist(e) {
		e = s.save()
	}
	return
}
func (s *settings) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var (
			i struct{ Data interface{} }
			a = "error:Unknown operation!"
			d interface{}
		)
		check(json.NewDecoder(r.Body).Decode(&i))
		switch v := i.Data.(type) {
		case string:
			a, d = s.setTorr(v)
			check(s.save())
		case bool:
			if v {
				a, d = s.inputTorr(r.Host)
			} else {
				a, d = s.switchPlayer(r.URL.Query().Get("id"))
				check(s.save())
			}
		}
		svcAnswer(w, a, d)
	} else {
		m := new(runtime.MemStats)
		runtime.ReadMemStats(m)
		i := struct {
			Name, Version, Up, Platform string
			MemSys, MemAlloc            uint64
			Plugins                     interface{}
			*settings
		}{Name, Vers, (time.Since(started) / time.Second * time.Second).String(), runtime.GOOS + "/" + runtime.GOARCH, m.Sys, m.Alloc, nil, stg}
		if p, e := plugsInfo(); e == nil {
			i.Plugins = p
		} else {
			i.Plugins = e.Error()
		}
		j := json.NewEncoder(w)
		j.SetIndent("", "  ")
		j.Encode(&i)
	}
}
func (s *settings) switchPlayer(id string) (string, interface{}) {
	s.HTML5X[id] = !s.HTML5X[id]
	return "reload:menu", nil
}
func (s *settings) setTorr(t string) (a string, d interface{}) {
	if v, e := checkTorr(t); e != nil {
		a = "error:" + e.Error()
		s.TorrServer = ""
	} else {
		if v == "" {
			v = "success:TorrServer: {dic:label:none|None}"
		} else {
			v = "success:TorrServer " + v + ": " + t
		}
		a, d = "data", map[string][]map[string]string{"actions": {{"action": v}, {"action": "back"}, {"action": "reload:menu"}}}
		s.TorrServer = t
	}
	return
}
func (s *settings) inputTorr(hst string) (string, interface{}) {
	a, t := "execute:http://"+hst+"/msx/input?addr", hst
	if s.TorrServer != "" {
		t = s.TorrServer
	} else if i := strings.LastIndexByte(t, ':'); i > 0 {
		t = t[:i] + ":8090"
		if v, e := checkTorr(t); e == nil && v != "" {
			a = "[" + a + "|info:Torrserver " + v + ": " + t + "]"
		} else if i = strings.LastIndexByte(t, '.'); i > 0 {
			t = t[:i+1]
		}
	} else if i = strings.LastIndexByte(t, '.'); i > 0 {
		t = t[:i]
	}
	return a, map[string]string{
		"action":    "execute:http://" + hst + "/settings",
		"headline":  "{dic:Address|Address of} Torrserver:",
		"extension": "<IP>:8090",
		"value":     t,
	}
}
