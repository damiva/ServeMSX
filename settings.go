package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const pthSettings, cHTML5X, cCompressed, cPhoto, cMarksLIFO = "settings.json", 1, 2, 4, 8

type settings struct {
	TorrServer, FFmpegCMD, FFmpegPORT string
	Clients                           map[string]int
}
type client struct{ Addr, Platform, Player, Vers string }

var (
	stg     = new(settings)
	clients = make(map[string]client)
)

func init() {
	stg.FFmpegPORT = "8009"
	http.Handle("/settings", stg)
}
func (s *settings) save() (e error) {
	var f *os.File
	mutexR.Lock()
	if f, e = os.Create(pthSettings); e == nil {
		j := json.NewEncoder(f)
		j.SetIndent("", "  ")
		if e = j.Encode(s); e != nil {
			e = errors.New("Encoding " + pthSettings + " error: " + e.Error())
		}
		f.Close()
	}
	mutexR.Unlock()
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
	if stg.Clients == nil {
		stg.Clients = make(map[string]int)
	}
	return
}
func (s *settings) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var (
			i struct {
				Code string
				Data interface{}
			}
			a = "error:Unknown operation!"
			d interface{}
		)
		check(json.NewDecoder(r.Body).Decode(&i))
		if i.Code != "" {
			if i.Code[0] == '#' {
				s.FFmpegCMD = i.Code[1:]
				if s.FFmpegCMD != "" {
					var e error
					s.FFmpegCMD, e = exec.LookPath(i.Code[1:])
					check(e)
				}
			} else {
				s.FFmpegPORT = i.Code
			}
			a = "[back|reload:menu]"
			check(s.save())
		} else {
			switch v := i.Data.(type) {
			case string:
				a, d = s.setTorr(v, !r.URL.Query().Has("v"))
				check(s.save())
			case float64:
				s.Clients[r.URL.Query().Get("id")] ^= int(v)
				check(s.save())
				a = "reload:menu"
			case map[string]interface{}:
				for _, k := range [3]string{pthVideo, pthMusic, pthPhoto} {
					if s, o := v[k].(string); o {
						if e := os.Remove(k); e != nil && !os.IsNotExist(e) {
							panic(e)
						} else if s != "" {
							if i, e := os.Stat(s); e != nil {
								panic(e)
							} else if !i.IsDir() {
								panic(s + " is not a directory!")
							}
							check(os.Symlink(s, k))
						}
					}
				}
				a = "info:OK"
			default:
				a, d = s.inputTorr(r.Host)
			}
		}
		svcAnswer(w, a, d)
	} else {
		m := new(runtime.MemStats)
		runtime.ReadMemStats(m)
		i := struct {
			Name, Version, Up, Platform, Video, Music, Photo string
			MemSys, MemAlloc                                 uint64
			Plugins                                          interface{}
			*settings
		}{Name, Vers, (time.Since(started) / time.Second * time.Second).String(), runtime.GOOS + "/" + runtime.GOARCH, "", "", "", m.Sys, m.Alloc, nil, stg}
		if p, e := plugsInfo(); e == nil {
			i.Plugins = p
		} else {
			i.Plugins = e.Error()
		}
		i.Video, _ = os.Readlink(pthVideo)
		i.Music, _ = os.Readlink(pthMusic)
		i.Photo, _ = os.Readlink(pthPhoto)
		j := json.NewEncoder(w)
		j.SetIndent("", "  ")
		j.Encode(&i)
	}
}
func (s *settings) setTorr(t string, err bool) (a string, d interface{}) {
	if v, e := checkTorr(t); e != nil {
		s.TorrServer = ""
		if err {
			panic(e)
		} else {
			a = "error:" + e.Error()
		}
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
