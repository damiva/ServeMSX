package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const gitRepo = "damiva/ServeMSX"

func init() {
	http.HandleFunc("/restart", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			svcAnswer(w, "reload", nil)
			startFocus = ">stg>rst"
		} else {
			w.Write([]byte("Restarting..."))
		}
		time.AfterFunc(time.Second, func() { mutex.Lock(); os.Exit(0) })
	})
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		svcAnswer(w, "warn:Underconstruction!", nil)
	})

}
func download(src string, dat ...interface{}) error {
	var (
		r *http.Response
		e error
	)
	if len(dat) > 0 {
		switch v := append(dat, nil)[1].(type) {
		case nil:
			r, e = http.Get(src)
		case http.Header:
			var q *http.Request
			if q, e = http.NewRequest("GET", src, nil); e == nil {
				q.Header = v
			}
			r, e = http.DefaultClient.Do(q)
		default:
			var b []byte
			if b, e = json.Marshal(dat[1]); e == nil {
				r, e = http.Post(src, "application/json", bytes.NewReader(b))
			}
		}
		if e == nil {
			defer r.Body.Close()
			defer ioutil.ReadAll(r.Body)
			if r.StatusCode != 200 {
				e = errors.New("Get " + src + " answered: " + r.Status)
			} else {
				switch v := dat[0].(type) {
				case *[]byte:
					*v, e = ioutil.ReadAll(r.Body)
				case string:
					var f *os.File
					if f, e = os.Create(v); e == nil {
						defer f.Close()
						_, e = io.Copy(f, r.Body)
					}
				default:
					if e = json.NewDecoder(r.Body).Decode(dat[0]); e != nil {
						e = errors.New("Decoding " + src + " error: " + e.Error())
					}
				}
			}
		}
	}
	return e
}
