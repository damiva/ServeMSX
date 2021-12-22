package tengohttp

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/d5/tengo/v2"
)

func (s *server) request(args ...tengo.Object) (tengo.Object, error) {
	var (
		ctp string
		met = "GET"
		brd io.Reader
		opt *tengo.Map
		req *http.Request
	)
	switch len(args) {
	case 2:
		switch v := args[1].(type) {
		case *tengo.String:
			met = v.Value
		case *tengo.Bytes:
			brd, met, ctp = bytes.NewReader(v.Value), "POST", "text/plain"
		case *tengo.Map:
			opt = v
			if o, k := opt.Value["body"]; k {
				met, ctp = "POST", "text/plain"
				switch v := o.(type) {
				case *tengo.Map:
					brd, ctp = strings.NewReader(url.Values(map2vals(v)).Encode()), "application/x-www-form-urlencoded"
				case *tengo.Bytes:
					brd = bytes.NewReader(v.Value)
				default:
					s, _ := tengo.ToString(o)
					brd = strings.NewReader(s)
				}
			}
			if o, k := opt.Value["method"]; k {
				met, _ = tengo.ToString(o)
				met = strings.ToUpper(met)
			}
		default:
			return nil, tengo.ErrInvalidArgumentType{Name: "second", Expected: "map/bytes/string", Found: args[1].TypeName()}
		}
		fallthrough
	case 1:
		var e error
		s, _ := tengo.ToString(args[0])
		if req, e = http.NewRequest(met, s, brd); e != nil {
			return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
		} else if ctp != "" {
			req.Header.Set("Content-Type", ctp)
		}
	default:
		return nil, tengo.ErrWrongNumArguments
	}
	if opt != nil {
		if o, k := opt.Value["query"]; k {
			if q, k := o.(*tengo.Map); k {
				req.URL.RawQuery = url.Values(map2vals(q)).Encode()
			} else if s, _ := tengo.ToString(o); s != "" {
				req.URL.RawQuery = s
			}
		}
		if o, k := opt.Value["header"].(*tengo.Map); k {
			for h, hv := range map2vals(o) {
				req.Header[h] = hv
			}
		}
		if o, k := opt.Value["cookies"].(*tengo.Array); k {
			for _, v := range o.Value {
				if m, k := v.(*tengo.Map); k {
					if n, _ := tengo.ToString(m.Value["name"]); n != "" {
						s, _ := tengo.ToString(m.Value["value"])
						req.AddCookie(&http.Cookie{Name: n, Value: s})
					}
				}
			}
		}
		if u, k := opt.Value["user"].(*tengo.String); k && u.Value != "" {
			if p, k := opt.Value["pass"].(*tengo.String); k && p.Value != "" {
				req.SetBasicAuth(u.Value, p.Value)
			}
		}
		if o, k := opt.Value["follow"]; k {
			if o.IsFalsy() {
				s.c.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
			}
		}
		if o, k := opt.Value["timeout"].(*tengo.Int); k && o.Value > 0 {
			s.c.Timeout = time.Duration(o.Value) * time.Second
		}
	}
	rsp, e := s.c.Do(req)
	if e != nil {
		return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
	}
	defer rsp.Body.Close()
	bs, e := ioutil.ReadAll(rsp.Body)
	if e != nil {
		return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
	}
	cs := new(tengo.Array)
	for _, cc := range rsp.Cookies() {
		cs.Value = append(cs.Value, &tengo.Map{Value: map[string]tengo.Object{"name": &tengo.String{Value: cc.Name}, "value": &tengo.String{Value: cc.Value}}})
	}
	au, ap, _ := rsp.Request.BasicAuth()
	return &tengo.Map{Value: map[string]tengo.Object{
		"status":  &tengo.Int{Value: int64(rsp.StatusCode)},
		"user":    &tengo.String{Value: au},
		"pass":    &tengo.String{Value: ap},
		"header":  vals2map(rsp.Header),
		"cookies": cs,
		"body":    &tengo.Bytes{Value: bs},
		"size":    &tengo.Int{Value: rsp.ContentLength},
		"url":     &tengo.String{Value: rsp.Request.URL.String()},
	}}, nil
}
