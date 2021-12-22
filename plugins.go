package main

import (
	tengofiles "ServeMSX/tengoFiles"
	tengohttp "ServeMSX/tengoHTTP"
	_ "embed"
	"encoding/json"
	"errors"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
)

const pthPlugs, manifest, mainTengo = "plugins", "manifest.json", "main.tengo"

type pluginf struct{ Label, Image, Icon, URL string }
type pluginfo struct {
	pluginf
	Name  string
	Error error `json:",omitempty"`
}

//go:embed index.html.gz
var html []byte
var plugMemory = make(map[string]tengo.Object)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			w.Header().Set("Content-Encoding", "gzip")
			w.Write(html)
		} else if p := strings.SplitN(r.URL.Path[1:], "/", 2); len(p) < 2 {
			panic(http.StatusNotFound)
		} else if i, e := os.Stat(filepath.Join(pthPlugs, p[0])); os.IsExist(e) || e == nil && !i.IsDir() {
			panic(http.StatusNotFound)
		} else if e != nil {
			panic(e)
		} else {
			pth := filepath.Join(pthPlugs, p[0], filepath.Clean(p[1]))
			if i, e := os.Stat(pth); os.IsNotExist(e) {
				tengoRun(w, r, filepath.Join(pthPlugs, p[0], mainTengo), p[0], p[1])
			} else if e != nil {
				panic(e)
			} else if i.IsDir() {
				tengoRun(w, r, filepath.Join(pth, mainTengo), p[0], p[1])
			} else if strings.ToLower(filepath.Ext(pth)) == ".tengo" {
				tengoRun(w, r, pth, p[0], p[1])
			} else {
				http.ServeFile(w, r, pth)
			}
		}
	})
}
func plugInfo(n string) (p pluginf, e error) {
	var f *os.File
	if f, e = os.Open(filepath.Join(pthPlugs, n, manifest)); e == nil {
		if e = json.NewDecoder(f).Decode(&p); e != nil {
			e = errors.New("Decoding " + manifest + " of " + n + " error: " + e.Error())
		}
		f.Close()
	}
	return
}
func plugsInfo() (ps []pluginfo, er error) {
	var d []fs.FileInfo
	if d, er = ioutil.ReadDir(pthPlugs); er == nil {
		for _, f := range d {
			if n := f.Name(); f.IsDir() {
				p, e := plugInfo(n)
				ps = append(ps, pluginfo{p, n, e})
			}
		}
	} else if os.IsNotExist(er) {
		er = nil
	}
	return
}
func tengoRun(w http.ResponseWriter, r *http.Request, script, plug, path string) {
	var t *tengo.Script
	if b, e := ioutil.ReadFile(script); os.IsNotExist(e) {
		panic(404)
	} else if e != nil {
		panic(e)
	} else {
		t = tengo.NewScript(b)
		check(t.Add("panic", &tengo.BuiltinFunction{Name: "panic", Value: tengoPanic}))
	}
	plugpath := filepath.Join(pthPlugs, plug)
	t.EnableFileImport(true)
	t.SetImportDir(plugpath)
	mm := stdlib.GetModuleMap("math", "text", "times", "rand", "json", "base64", "hex")
	mm.AddBuiltinModule("files", (&tengofiles.FS{Dir: plugpath}).GetModuleMap())
	mm.AddBuiltinModule("server", tengohttp.GetModuleMAP(w, r, nil, map[string]tengo.Object{
		"version":    &tengo.String{Value: Vers},
		"script":     &tengo.String{Value: script},
		"plugin":     &tengo.String{Value: plug},
		"path":       &tengo.String{Value: path},
		"torrserver": &tengo.String{Value: stg.TorrServer},
		"base_url":   &tengo.String{Value: "http://" + r.Host + "/" + plug + "/"},
		"memory":     &tengo.UserFunction{Name: "memory", Value: tengoMem(plug)},
		"log_err":    &tengo.UserFunction{Name: "log_err", Value: tengoLog(plug, log.Default())},
		"log_inf":    &tengo.UserFunction{Name: "log_inf", Value: tengoLog(plug, out)},
	}))
	t.SetImports(mm)
	_, e := t.Run()
	check(e)
}
func tengoLog(plg string, log *log.Logger) func(...tengo.Object) (tengo.Object, error) {
	return func(args ...tengo.Object) (tengo.Object, error) {
		vs := []interface{}{plg + ":"}
		for _, a := range args {
			v, _ := tengo.ToString(a)
			vs = append(vs, v)
		}
		log.Println(vs...)
		return nil, nil
	}
}
func tengoMem(plug string) func(...tengo.Object) (tengo.Object, error) {
	return func(args ...tengo.Object) (r tengo.Object, e error) {
		switch len(args) {
		case 0:
			r = plugMemory[plug]
		case 1:
			plugMemory[plug] = args[0]
		default:
			e = tengo.ErrWrongNumArguments
		}
		return
	}
}
func tengoPanic(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	} else if e, o := args[0].(*tengo.Int); o {
		panic(int(e.Value))
	} else if e, o := tengo.ToString(args[0]); o {
		panic(e)
	}
	return nil, nil
}
