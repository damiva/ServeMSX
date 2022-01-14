package main

import (
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
	tengohttp "github.com/damiva/TengoHTTP"
)

const pthPlugs, manifest, mainTengo = "plugins", "manifest.json", "main.tengo"

type pluginf struct {
	Label, Image, Icon, URL string
	Torrent                 bool
}
type pluginfo struct {
	pluginf
	Name  string
	Error error `json:",omitempty"`
}

var plugMemory = make(map[string]tengo.Object)

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
func servePlugin(w http.ResponseWriter, r *http.Request, pn, pp string) {
	if i, e := os.Stat(filepath.Join(pthPlugs, pn)); os.IsExist(e) || e == nil && !i.IsDir() {
		panic(http.StatusNotFound)
	} else if e != nil {
		panic(e)
	} else {
		pth := filepath.Join(pthPlugs, pn, filepath.Clean(pp))
		if i, e := os.Stat(pth); os.IsNotExist(e) {
			tengoRun(w, r, filepath.Join(pthPlugs, pn, mainTengo), pn, pp)
		} else if e != nil {
			panic(e)
		} else if i.IsDir() {
			tengoRun(w, r, filepath.Join(pth, mainTengo), pn, pp)
		} else if strings.ToLower(filepath.Ext(pth)) == ".tengo" {
			tengoRun(w, r, pth, pn, pp)
		} else {
			http.ServeFile(w, r, pth)
		}
	}
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
	mm.AddBuiltinModule("server", tengohttp.GetModuleMAP(w, r, nil, map[string]tengo.Object{
		"version":    &tengo.String{Value: Vers},
		"script":     &tengo.String{Value: script},
		"plugin":     &tengo.String{Value: plug},
		"path":       &tengo.String{Value: path},
		"torrserver": &tengo.String{Value: stg.TorrServer},
		"base_url":   &tengo.String{Value: "http://" + r.Host + "/" + plug + "/"},
		"memory":     &tengo.UserFunction{Name: "memory", Value: tengoMem(plug)},
		"file":       &tengo.UserFunction{Name: "file", Value: tengoFile(plugpath)},
		"player":     &tengo.UserFunction{Name: "player", Value: tengoPlayer(r)},
		"log_err":    &tengo.UserFunction{Name: "log_err", Value: tengoLog(plug, log.Default())},
		"log_inf":    &tengo.UserFunction{Name: "log_inf", Value: tengoLog(plug, out)},
		"dictionary": &tengo.UserFunction{Name: "dictionary", Value: tengoDic},
	}))
	t.SetImports(mm)
	_, e := t.Run()
	check(e)
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
func tengoPlayer(r *http.Request) func(...tengo.Object) (tengo.Object, error) {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if l := len(args); l == 1 {
			args = append(args, tengo.FalseValue)
		} else if l != 2 {
			return nil, tengo.ErrWrongNumArguments
		}
		id, _ := tengo.ToString(args[0])
		switch v := args[1].(type) {
		case *tengo.Bool:
			ps := &tengo.Map{Value: make(map[string]tengo.Object)}
			for k, p := range playerProp(r.Host, id, !v.IsFalsy()) {
				ps.Value[k] = &tengo.String{Value: p}
			}
			return ps, nil
		case *tengo.String:
			iv, ul := true, v.Value
			switch ul[:6] {
			case "audio:":
				iv = false
				fallthrough
			case "video:":
				ul = ul[6:]
			}
			return &tengo.String{Value: playerURL(id, ul, iv)}, nil
		default:
			return nil, tengo.ErrInvalidArgumentType{Name: "first", Expected: "bool/string", Found: args[0].TypeName()}
		}
	}
}
func tengoFile(pth string) func(...tengo.Object) (tengo.Object, error) {
	return func(args ...tengo.Object) (rtn tengo.Object, err error) {
		switch len(args) {
		case 1:
			if n, _ := tengo.ToString(args[0]); n == "" {
				err = &tengo.ErrInvalidArgumentType{Name: "first", Expected: "not empty", Found: "empty"}
			} else if b, e := os.ReadFile(filepath.Join(pth, filepath.Clean(n))); e == nil {
				rtn = &tengo.Bytes{Value: b}
			} else if !os.IsNotExist(e) {
				rtn = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
			}
		case 2:
			if n, _ := tengo.ToString(args[0]); n == "" {
				err = &tengo.ErrInvalidArgumentType{Name: "first", Expected: "not empty", Found: "empty"}
			} else {
				var (
					b []byte
					d bool
					e error
				)
				if v, o := args[1].(*tengo.Bytes); o {
					b = v.Value
				} else if v, o := tengo.ToString(args[1]); o {
					b = []byte(v)
				} else {
					d = true
				}
				if d {
					e = os.Remove(n)
				} else {
					e = os.WriteFile(filepath.Join(pth, filepath.Clean(n)), b, 0666)
				}
				if e != nil {
					rtn = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
				}
			}
		default:
			err = tengo.ErrWrongNumArguments
		}
		return
	}
}
func tengoDic(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	} else if n, _, e := getDic(); e != nil {
		return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
	} else {
		return &tengo.String{Value: n}, nil
	}
}
