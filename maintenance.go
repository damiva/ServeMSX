package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

const gitRepo = "damiva/ServeMSX"

type gitrelease struct {
	Tarball_url, Tag_name string
	Assets                []struct{ Browser_download_url, Name string }
}

func init() {
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		if u := r.URL.Query().Get("link"); u != "" {
			check(download(u, mypath+".new", nil))
			check(os.Chmod(mypath+".new", 0777))
			mutex.Lock()
			defer mutex.Unlock()
			check(os.Rename(mypath, mypath+".old"))
			if e := os.Rename(mypath+".new", mypath); e != nil {
				os.Rename(mypath+".old", mypath)
				panic(e)
			}
			svcAnswer(w, "restart", nil)
			go restart()
			return
		} else {
			i, e := gitRelease("", "")
			check(e)
			if i.Tag_name != Vers {
				n := strings.Join([]string{Name, runtime.GOOS, runtime.GOARCH}, "-")
				for _, a := range i.Assets {
					if strings.HasPrefix(a.Name, n) {
						svcAnswer(w, "execute:http://"+r.Host+"/msx/dialog", dialogStg{
							"execute:fetch:http://" + r.Host + "/update?link=" + url.QueryEscape(a.Browser_download_url),
							"{dic:CheckUp|Check updates}:",
							Name,
							"{dic:HasUpdate|There are updates}:{br}" + Name + " {dic:label:version|Version} " + i.Tag_name + "{br}{br}{dix:Update}Would you like to update?",
						})
						return
					}
				}
			}
			svcAnswer(w, "info:{dic:NoUpdate|There are no updates!}", nil)
		}
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

/*
func checkUpdate(r *http.Request) (string, interface{}) {
	var ups [][3]string
	i, e := gitRelease("", "")
	check(e)
	if i.Tag_name != Vers {
		n := strings.Join([]string{Name, runtime.GOOS, runtime.GOARCH}, "-")
		for _, a := range i.Assets {
			if strings.HasPrefix(a.Name, n) {
				ups = append(ups, [3]string{Name, i.Tag_name, a.Browser_download_url})
			}
		}
	}
	for p, t := range stg.Plugins {
		if i, e := gitRelease(p, ""); e != nil {
			log.Println(e)
		} else if i.Tag_name != t {
			ups = append(ups, [3]string{p, i.Tag_name, i.Tarball_url})
		}
	}
	if len(ups) > 0 {
		dat := dialogStg{Action: "execute:fetch:http://" + r.Host + "/update?", Headline: "{dic:HasUpdate|There is an update}:", Extension: "{ico:system-update-alt}"}
		for _, u := range ups {
			dat.Action += "link=" + url.QueryEscape(u[2]) + "&name=" + url.QueryEscape(u[0]) + "&tag=" + url.QueryEscape(u[1]) + "&"
			dat.Value += u[0] + " {dic:label:version|Version}: " + u[1] + "{br}"
		}
		dat.Value += "{br}{dix:Update}Would you like to update (the application will be reloaded)?"
		return "execute:http://" + r.Host + "/msx/dialog", dat
	}
	return "info:{dic:NoUpdate|There is no update}!", nil
}
func update(r *http.Request, url, src, tag string) (string, interface{}) {
	if src == "" {
		check(download(url, mypath+".new", nil))
		mutex.Lock()
		defer mutex.Unlock()
		check(os.Rename(mypath, mypath+".old"))
		if e := os.Rename(mypath+".new", mypath); e != nil {
			os.Rename(mypath+".old", mypath)
			panic(e)
		}
		return "execute:fetch:http://" + r.Host + "/restart", nil
	} else {
		check(download(url, filepath.Join(pthPlugs, pthInstPlug), true))
		mutex.Lock()
		defer mutex.Unlock()
		check(untar(pthInstPlug, ""))
		stg.Plugins[src] = tag
		return "reload:menu", nil
	}
}
func untar(src, dst string) (e error) {
	p := filepath.Join(pthPlugs, dst)
	if e = os.MkdirAll(p, 0777); e == nil {
		var f *os.File
		if f, e = os.Open(src); e == nil {
			var h *tar.Header
			t := tar.NewReader(f)
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
				} else if e == io.EOF {
					e = nil
					break
				}
				if e != nil {
					break
				}
			}
			f.Close()
			if e != nil {
				if er := os.RemoveAll(p); er != nil {
					log.Println(e)
				}
			}
			if er := os.Remove(src); er != nil {
				log.Println(e)
			}
		}
	}
	return
}
*/
