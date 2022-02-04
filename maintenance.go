package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const gitRepo = "damiva/ServeMSX"

type gitrelease struct {
	Tarball_url, Tag_name string
	Assets                []struct{ Browser_download_url, Name string }
}

func init() {
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		if u := r.FormValue("link"); u != "" {
			if d := r.FormValue("dic"); d != "" {
				if e := download(d, pthDic, true); e != nil {
					log.Println(e)
				}
			}
			check(download(u, mypath+".new", nil))
			check(os.Chmod(mypath+".new", 0777))
			mutexR.Lock()
			defer mutexR.Unlock()
			check(os.Rename(mypath, mypath+".old"))
			if e := os.Rename(mypath+".new", mypath); e != nil {
				os.Rename(mypath+".old", mypath)
				panic(e)
			}
			svcAnswer(w, "delay:installation:"+strconv.Itoa(performSecs)+":restart", nil)
			go restart()
			return
		} else {
			i, e := gitRelease("", "")
			check(e)
			if i.Tag_name != Vers {
				n, as := strings.Join([]string{Name, runtime.GOOS, runtime.GOARCH}, "-"), []string{"", ""}
				_, l, _, _ := getDic()
				for _, a := range i.Assets {
					if strings.HasSuffix(a.Name, ".json.gz") {
						if l != "" && strings.HasPrefix(a.Name, l) {
							as[1] = url.QueryEscape(a.Browser_download_url)
						}
					} else if strings.HasPrefix(a.Name, n) {
						as[0] = url.QueryEscape(a.Browser_download_url)
					}
				}
				if as[0] != "" {
					if r.FormValue("v") == "" {
						w.Write([]byte(`{"link":"` + strings.Join(as, "&dic=") + `","tag":"` + i.Tag_name + `","wait":` + strconv.Itoa(performSecs) + `}`))
					} else {
						svcAnswer(w, "execute:http://"+r.Host+"/msx/dialog", dialogStg{
							"execute:fetch:http://" + r.Host + "/update?link=" + strings.Join(as, "&dic="),
							"{dic:CheckUp|Check updates}:",
							Name,
							"{dic:HasUpdate|There are updates}:{br}" + Name + " {dic:label:version|Version} " + i.Tag_name + "{br}{br}{dix:Update}Would you like to update?",
						})
					}
					return
				}
			}
			svcAnswer(w, "info:{dic:NoUpdate|There are no updates!}", nil)
		}
	})
	http.HandleFunc("/install", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			panic(404)
		} else if p := r.FormValue("remove"); p != "" {
			check(os.RemoveAll(filepath.Join(pthPlugs, p)))
		} else if f, fh, e := r.FormFile("file"); e != nil {
			panic(e)
		} else {
			defer f.Close()
			if filepath.Ext(fh.Filename) != ".tgz" {
				panic(400)
			}
			p = filepath.Join(pthPlugs, strings.SplitN(fh.Filename, ".", 2)[0])
			os.RemoveAll(p)
			check(os.MkdirAll(p, 0777))
			z, e := gzip.NewReader(f)
			check(e)
			defer z.Close()
			t := tar.NewReader(z)
			var h *tar.Header
			for {
				if h, e = t.Next(); e == nil {
					pp := filepath.Join(p, h.Name)
					switch h.Typeflag {
					case tar.TypeDir:
						e = os.MkdirAll(pp, 0777)
					case tar.TypeReg:
						var ff *os.File
						if ff, e = os.Create(pp); e == nil {
							_, e = io.Copy(ff, t)
							ff.Close()
						}
					}
				}
				if e != nil {
					break
				}
			}
			if e != nil && e != io.EOF {
				panic(e)
			}
		}
		w.Write([]byte("true"))
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
		r, e = http.Get(src)
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
