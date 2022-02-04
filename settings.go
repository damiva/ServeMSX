package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const pthSettings, cHTML5X, cCompressed, cPhoto, cMarksLIFO, cPhotoScale = "settings.json", 1, 2, 4, 8, 16

type settings struct {
	TorrServer, FFmpeg, FFprobe, FFstream, Background string
	VideoWall                                         int
	Clients                                           map[string]int
}
type client struct{ Addr, Platform, Player, Vers string }

var (
	stg     = &settings{"", "ffmpeg", "ffprobe", "8009", "background.jpg", 0, make(map[string]int)}
	clients = make(map[string]client)
)

func init() {
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
func (s *settings) switcher(r *http.Request, sv int, fv, tv string) string {
	if s.Clients[r.FormValue("id")]&sv == 0 {
		return "{col:msx-white}" + fv + " {ico:toggle-off}{col:msx-white-soft} " + tv
	} else {
		return "{col:msx-white-soft}" + fv + " {col:msx-white}{ico:toggle-on} " + tv
	}
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
			a, d = s.setTorr(v, !r.URL.Query().Has("v"))
			check(s.save())
		case float64:
			if k := int(v); k < 100 {
				s.Clients[r.URL.Query().Get("id")] ^= k
			} else {
				s.VideoWall = k - 100
			}
			check(s.save())
			a = "[cleanup|reload:menu]"
		case map[string]interface{}:
			for k, vv := range v {
				if ss, o := vv.(string); o {
					switch k {
					case pthVideo, pthMusic, pthPhoto:
						if e := os.Remove(k); e != nil && !os.IsNotExist(e) {
							panic(e)
						} else if ss == "" {
							continue
						} else if i, e := os.Stat(ss); e != nil {
							panic(e)
						} else if !i.IsDir() {
							panic(ss + " is not a directory!")
						} else {
							check(os.Symlink(ss, k))
						}
					case "Background":
						s.Background = ss
						check(s.save())
					case "FFmpeg":
						o = false
						fallthrough
					case "FFprobe":
						check(checkFFmpeg(o))
						check(s.save())
					case "FFstream":
						if ss != "" {
							_, e := strconv.Atoi(ss)
							check(e)
						}
						s.FFstream = ss
						check(s.save())
					}
				}
			}
			a = "info:OK"
		default:
			a, d = s.inputTorr(r.Host)
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
