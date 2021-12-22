package tengofiles

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/d5/tengo/v2"
)

type FS struct {
	Dir string
	RO  bool
}

func (f *FS) GetModuleMap() map[string]tengo.Object {
	r := map[string]tengo.Object{"info": &tengo.UserFunction{Name: "info", Value: f.info}, "read": &tengo.UserFunction{Name: "read", Value: f.read}}
	if !f.RO {
		r["write"] = &tengo.UserFunction{Name: "write", Value: f.write}
	}
	return r
}
func (f *FS) inf(a tengo.Object, inf bool) (p string, i fs.FileInfo, e error) {
	p, _ = tengo.ToString(a)
	p = filepath.Join(f.Dir, filepath.Clean(p))
	if inf {
		i, e = os.Stat(p)
	}
	return
}
func (f *FS) info(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
	} else if _, i, e := f.inf(args[0], true); e == nil {
		ret = fi2obj(i)
	} else {
		ret = err2obj(e)
	}
	return
}
func (f *FS) read(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		err = tengo.ErrWrongNumArguments
	} else if p, i, e := f.inf(args[0], true); e != nil {
		ret = err2obj(e)
	} else if !i.IsDir() {
		if b, e := ioutil.ReadFile(p); e == nil {
			ret = &tengo.Bytes{Value: b}
		} else {
			ret = err2obj(e)
		}
	} else if fs, e := ioutil.ReadDir(p); e != nil {
		ret = err2obj(e)
	} else {
		var a []tengo.Object
		for _, i := range fs {
			a = append(a, fi2obj(i))
		}
		ret = &tengo.Array{Value: a}
	}
	return
}
func (f *FS) write(args ...tengo.Object) (ret tengo.Object, err error) {
	switch len(args) {
	case 2:
		if b, o := tengo.ToByteSlice(args[1]); o {
			p, _, _ := f.inf(args[0], false)
			err = ioutil.WriteFile(p, b, 0666)
		} else if !args[1].IsFalsy() {
			ret = tengo.FalseValue
		} else if p, i, e := f.inf(args[0], true); e != nil {
			ret = err2obj(e)
		} else if i.IsDir() {
			err = os.RemoveAll(p)
		} else {
			err = os.Remove(p)
		}
		if err != nil {
			ret, err = err2obj(err), nil
		}
	case 1:
		p, _, _ := f.inf(args[0], false)
		if e := os.Mkdir(p, 0777); e == nil {
			ret = tengo.TrueValue
		} else if os.IsExist(e) {
			ret = tengo.FalseValue
		} else {
			ret = err2obj(e)
		}
	default:
		err = tengo.ErrWrongNumArguments
	}
	return
}
func err2obj(e error) (o tengo.Object) {
	if e != nil && !os.IsNotExist(e) {
		o = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
	}
	return
}
func fi2obj(i fs.FileInfo) tengo.Object {
	r := map[string]tengo.Object{
		"name":     &tengo.String{Value: i.Name()},
		"is_dir":   tengo.FalseValue,
		"size":     &tengo.Int{Value: i.Size()},
		"mod_time": &tengo.Time{Value: i.ModTime()},
	}
	if i.IsDir() {
		r["is_dir"] = tengo.TrueValue
	}
	return &tengo.Map{Value: r}
}
