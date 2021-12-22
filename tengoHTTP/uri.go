package tengohttp

import (
	"net/url"

	"github.com/d5/tengo/v2"
)

func (s *server) parse(args ...tengo.Object) (r tengo.Object, e error) {
	if l := len(args); l == 0 {
		args = append(args, &tengo.String{Value: "http://" + s.r.Host + s.r.RequestURI})
	} else if l > 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	switch a := args[0].(type) {
	case *tengo.String:
		if u, e := url.Parse(a.Value); e != nil {
			r = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
		} else {
			ret := &tengo.Map{Value: map[string]tengo.Object{
				"scheme":       &tengo.String{Value: u.Scheme},
				"opaque":       &tengo.String{Value: u.Opaque},
				"user":         &tengo.String{Value: u.User.Username()},
				"host":         &tengo.String{Value: u.Host},
				"path":         &tengo.String{Value: u.Path},
				"raw_path":     &tengo.String{Value: u.EscapedPath()},
				"query":        &tengo.String{Value: u.RawQuery},
				"fragment":     &tengo.String{Value: u.Fragment},
				"raw_fragment": &tengo.String{Value: u.EscapedFragment()},
			}}
			if p, o := u.User.Password(); o {
				ret.Value["pass"] = &tengo.String{Value: p}
			}
			r = ret
		}
	case *tengo.Map:
		u, au, ap := new(url.URL), "", ""
		for k, v := range a.Value {
			if s, _ := tengo.ToString(v); s != "" {
				switch k {
				case "scheme":
					u.Scheme = s
				case "opaque":
					u.Opaque = s
				case "user":
					au = s
				case "pass":
					ap = s
				case "host":
					u.Host = s
				case "path":
					u.Path = s
				case "query":
					u.RawQuery = s
				case "fragment":
					u.Fragment = s
				}
			}
		}
		if au != "" {
			if ap != "" {
				u.User = url.UserPassword(au, ap)
			} else {
				u.User = url.User(au)
			}
		}
		r = &tengo.String{Value: u.String()}
	default:
		e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string/map", Found: args[0].TypeName()}
	}
	return
}
func (s *server) resolve(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) == 1 {
		args = append(args, &tengo.String{Value: "http://" + s.r.Host + s.r.URL.Path})
	}
	if len(args) != 2 {
		err = tengo.ErrWrongNumArguments
	} else if ro, o := args[0].(*tengo.String); !o {
		err = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string", Found: args[0].TypeName()}
	} else if bo, o := args[1].(*tengo.String); !o {
		err = tengo.ErrInvalidArgumentType{Name: "second", Expected: "string", Found: args[1].TypeName()}
	} else if bu, e := url.Parse(bo.Value); e != nil {
		ret = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
	} else if ru, e := url.Parse(ro.Value); e != nil {
		ret = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
	} else {
		ret = &tengo.String{Value: bu.ResolveReference(ru).String()}
	}
	return
}
func query(args ...tengo.Object) (r tengo.Object, e error) {
	if len(args) != 1 {
		e = tengo.ErrWrongNumArguments
	} else if a, o := args[0].(*tengo.String); o {
		if vs, er := url.ParseQuery(a.Value); er != nil {
			r = &tengo.Error{Value: &tengo.String{Value: e.Error()}}
		} else {
			r = vals2map(vs)
		}
	} else if a, o := args[0].(*tengo.Map); o {
		u := new(url.Values)
		for k, v := range a.Value {
			if ao, o := v.(*tengo.Array); o {
				for _, av := range ao.Value {
					s, _ := tengo.ToString(av)
					u.Add(k, s)
				}
			} else {
				s, _ := tengo.ToString(v)
				u.Set(k, s)
			}
		}
		r = &tengo.String{Value: u.Encode()}
	} else {
		e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string/map", Found: args[0].TypeName()}
	}
	return
}
func encode(args ...tengo.Object) (r tengo.Object, e error) {
	pth := false
	switch len(args) {
	case 2:
		pth = !args[1].IsFalsy()
		fallthrough
	case 1:
		if s, o := args[0].(*tengo.String); !o {
			e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string/map", Found: args[0].TypeName()}
		} else if pth {
			r = &tengo.String{Value: url.PathEscape(s.Value)}
		} else {
			r = &tengo.String{Value: url.QueryEscape(s.Value)}
		}
	default:
		e = tengo.ErrWrongNumArguments
	}
	return
}
func decode(args ...tengo.Object) (r tengo.Object, e error) {
	pth := false
	switch len(args) {
	case 2:
		pth = !args[1].IsFalsy()
		fallthrough
	case 1:
		if s, o := args[0].(*tengo.String); !o {
			e = tengo.ErrInvalidArgumentType{Name: "first", Expected: "string/map", Found: args[0].TypeName()}
		} else {
			var rs string
			if pth {
				rs, e = url.PathUnescape(s.Value)
			} else {
				rs, e = url.QueryUnescape(s.Value)
			}
			if e != nil {
				r, e = &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
			} else {
				r = &tengo.String{Value: rs}
			}
		}
	default:
		e = tengo.ErrWrongNumArguments
	}
	return
}

func vals2map(vals map[string][]string) *tengo.Map {
	r := &tengo.Map{Value: make(map[string]tengo.Object)}
	for k, vs := range vals {
		a := &tengo.Array{}
		for _, v := range vs {
			a.Value = append(a.Value, &tengo.String{Value: v})
		}
		r.Value[k] = a
	}
	return r
}
func map2vals(m *tengo.Map) map[string][]string {
	vals := make(map[string][]string)
	for k, o := range m.Value {
		if a, i := o.(*tengo.Array); i {
			for _, v := range a.Value {
				s, _ := tengo.ToString(v)
				vals[k] = append(vals[k], s)
			}
		} else {
			s, _ := tengo.ToString(o)
			vals[k] = []string{s}
		}
	}
	return vals
}
