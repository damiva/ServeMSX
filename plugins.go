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

type pluginf struct{ Label, Image, Icon, URL string }
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
	//	mm.AddBuiltinModule("files", (&tengofiles.FS{Dir: plugpath}).GetModuleMap())
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
		l, id := len(args), ""
		if l == 0 {
			return nil, tengo.ErrWrongNumArguments
		} else if id, _ = tengo.ToString(args[0]); id == "" {
			return nil, nil
		} else if l == 1 {
			r := clients[id].Player
			if stg.HTML5X[id] {
				r = "html5x"
			}
			return &tengo.String{Value: r}, nil
		} else {
			switch v := args[1].(type) {
			case *tengo.String:
				vid, it := true, false
				switch l {
				case 3:
					vid, it = !args[2].IsFalsy(), true
					fallthrough
				case 2:
					rtn := playerURL(id, v.Value, vid)
					if !it {
						rtn = rtn[strings.IndexByte(rtn, ':')+1:]
					}
					return &tengo.String{Value: rtn}, nil
				default:
					return nil, tengo.ErrWrongNumArguments
				}
			case *tengo.Bool:
				var img, ref string
				switch l {
				case 4:
					img, _ = tengo.ToString(args[3])
					fallthrough
				case 3:
					ref, _ = tengo.ToString(args[2])
					fallthrough
				case 2:
					if v.IsFalsy() {
						ref = "*"
					}
					m := &tengo.Map{Value: make(map[string]tengo.Object)}
					for k, v := range playerProp(r.Host, id, ref, img) {
						m.Value[k] = &tengo.String{Value: v}
					}
					return m, nil
				default:
					return nil, tengo.ErrWrongNumArguments
				}
			default:
				return nil, tengo.ErrInvalidArgumentType{Name: "second", Expected: "string/bool", Found: args[1].TypeName()}
			}
		}
	}
}
func tengoFile(pth string) func(...tengo.Object) (tengo.Object, error) {
	return func(args ...tengo.Object) (tengo.Object, error) {
		flg := os.O_CREATE
		if l := len(args); l < 1 || l > 3 {
			return nil, tengo.ErrWrongNumArguments
		} else if n, _ := tengo.ToString(args[0]); n == "" {
			return nil, nil
		} else if pth = filepath.Join(pth, filepath.Clean(n)); l == 1 {
			if b, e := os.ReadFile(pth); e == nil {
				return &tengo.Bytes{Value: b}, nil
			} else if !os.IsNotExist(e) {
				return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
			}
			return nil, nil
		} else if l == 3 && !args[2].IsFalsy() {
			flg |= os.O_APPEND
		} else {
			flg |= os.O_TRUNC
		}
		if b, o := tengo.ToString(args[1]); o {
			f, e := os.OpenFile(pth, flg, 0666)
			if e != nil {
				return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
			}
			defer f.Close()
			n, e := f.WriteString(b)
			if e != nil {
				return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
			}
			return &tengo.Int{Value: int64(n)}, nil
		} else if e := os.Remove(pth); e == nil {
			return tengo.TrueValue, nil
		} else if !os.IsExist(e) {
			return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
		}
		return nil, nil
	}
}
