package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"runtime"
	"time"
)

const pthSettings = "settings.json"

type settings struct {
	TorrServer  string
	Russian     bool
	HistoryURLs [][]string
}

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
func (s *settings) historyURL(url string, dat ...string) (e error) {
	f, c, a := -1, false, len(dat) > 0
	for i, h := range s.HistoryURLs {
		if url == h[0] {
			f = i
			break
		}
	}
	switch {
	case f == 0:
		if c = !a; c {
			s.HistoryURLs = s.HistoryURLs[1:]
		}
	case f > 0:
		s.HistoryURLs, c = append(s.HistoryURLs[:f], s.HistoryURLs[f+1:]...), true
		fallthrough
	default:
		if a {
			s.HistoryURLs, c = append([][]string{append([]string{url}, dat...)}, s.HistoryURLs...), true
			if len(s.HistoryURLs) > 24 {
				s.HistoryURLs = s.HistoryURLs[:24]
			}
		}
	}
	if c {
		e = s.save()
	}
	return e
}
func (s *settings) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var (
			s struct{ Data interface{} }
			a string
		)
		if r.Method != "POST" {
			panic(400)
		} else if e := json.NewDecoder(r.Body).Decode(&s); e != nil {
			panic(e)
		}
		msx := r.URL.Query().Has("v")
		switch v := s.Data.(type) {
		case string:
			if t, e := checkTorr(v); e == nil {
				if t == "" {
					if stg.Russian {
						t = "не задан"
					} else {
						t = "is not set"
					}
				}
				a, stg.TorrServer = "[success:TorrServer "+t+"|back|reload:menu]", v
			} else {
				a, stg.TorrServer = "error:"+e.Error(), ""
			}
			check(stg.save())
		case bool:
			stg.Russian = v
			check(stg.save())
			if msx {
				a, startFocus = "reload", ">stg>dic"
			}
		}
		if msx {
			svcAnswer(w, a, nil)
		}
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
