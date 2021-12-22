package tengohttp

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/d5/tengo/v2"
)

func (s *server) read(args ...tengo.Object) (r tengo.Object, e error) {
	switch len(args) {
	case 0:
		if bs, er := ioutil.ReadAll(s.r.Body); er == nil {
			r = &tengo.Bytes{Value: bs}
		} else {
			r = &tengo.Error{Value: &tengo.String{Value: er.Error()}}
		}
	case 1:
		switch a := args[0].(type) {
		case *tengo.String:
			r = &tengo.String{Value: s.r.FormValue(a.Value)}
		case *tengo.Bool:
			if er := s.r.ParseForm(); er != nil {
				r = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
			} else if a.IsFalsy() {
				r = vals2map(s.r.PostForm)
			} else {
				r = vals2map(s.r.Form)
			}
		default:
			e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string/bool", Found: args[0].TypeName()}
		}
	default:
		e = tengo.ErrWrongNumArguments
	}
	return
}
func (s *server) write(args ...tengo.Object) (tengo.Object, error) {
	c := 0
	for n, arg := range args {
		switch a := arg.(type) {
		case *tengo.Map:
			if s.h {
				return nil, tengo.ErrInvalidArgumentType{Name: strconv.Itoa(n) + "-th", Expected: "nor map or int", Found: a.TypeName()}
			}
			for k, vs := range map2vals(a) {
				for _, v := range vs {
					s.w.Header().Add(k, v)
				}
			}
		case *tengo.Int:
			if s.h {
				return nil, tengo.ErrInvalidArgumentType{Name: strconv.Itoa(n) + "-th", Expected: "nor map or int", Found: a.TypeName()}
			}
			s.h, c = true, int(a.Value)
		default:
			var e error
			if c > 0 {
				if v, o := tengo.ToString(a); !o {
					s.w.WriteHeader(c)
				} else if c < 300 {
					s.w.WriteHeader(c)
					_, e = s.w.Write([]byte(v))
				} else if c < 400 {
					http.Redirect(s.w, s.r, v, c)
				} else {
					http.Error(s.w, v, c)
				}
			} else if v, o := a.(*tengo.Bytes); o {
				_, e = s.w.Write(v.Value)
			} else if v, o := tengo.ToString(a); o {
				_, e = s.w.Write([]byte(v))
			}
			if e != nil {
				return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
			} else {
				c, s.h = 0, true
			}
		}
	}
	return nil, nil
}
