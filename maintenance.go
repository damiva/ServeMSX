package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const gitRepo = "damiva/ServeMSX"

type gitrelease struct {
	Tarball_url, Tag_name string
	Assets                []struct{ Browser_download_url, Name string }
}

func init() {
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		svcAnswer(w, "warn:Underconstruction!", nil)
	})

}
func download(src string, dst, opt interface{}) error {
	var (
		r *http.Response
		e error
		g bool
	)
	switch v := opt.(type) {
	case nil:
		r, e = http.Get(src)
	case bool:
		g = v
	case http.Header:
		var q *http.Request
		if q, e = http.NewRequest("GET", src, nil); e == nil {
			q.Header = v
		}
		r, e = http.DefaultClient.Do(q)
	default:
		var b []byte
		if b, e = json.Marshal(opt); e == nil {
			r, e = http.Post(src, "application/json", bytes.NewReader(b))
		}
	}
	if e == nil {
		defer r.Body.Close()
		defer ioutil.ReadAll(r.Body)
		if r.StatusCode != 200 {
			e = errors.New("Downloading " + src + " answered: " + r.Status)
		} else {
			switch v := dst.(type) {
			case string:
				var f *os.File
				if f, e = os.Create(v); e == nil {
					if g {
						var z *gzip.Reader
						if z, e = gzip.NewReader(r.Body); e == nil {
							_, e = io.Copy(f, z)
							z.Close()
						}
					} else {
						_, e = io.Copy(f, r.Body)
					}
					f.Close()
					if e != nil {
						os.Remove(v)
					}
				}
			case *[]byte:
				*v, e = ioutil.ReadAll(r.Body)
			default:
				if e = json.NewDecoder(r.Body).Decode(dst); e != nil {
					e = errors.New("Decoding " + src + " error: " + e.Error())
				}
			}
		}
	}
	return e
}
func gitRelease(repo, tag string) (info *gitrelease, err error) {
	info = new(gitrelease)
	if repo == "" {
		repo = gitRepo
	}
	if tag == "" {
		tag = "latest"
	} else {
		tag = "tags/" + tag
	}
	err = download("https://api.github.com/repos/"+repo+"/releases/"+tag, info, http.Header{"Accept": {"application/vnd.github.v3+json"}})
	return
}
