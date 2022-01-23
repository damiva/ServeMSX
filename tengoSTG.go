package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/d5/tengo/v2"
)

func tengoSTG(r *http.Request) (m map[string]tengo.Object) {
	id := r.URL.Query().Get("id")
	m = map[string]tengo.Object{
		"dic": &tengo.UserFunction{Name: "dic", Value: func(args ...tengo.Object) (tengo.Object, error) {
			if _, n, _, e := getDic(); e != nil {
				return &tengo.Error{Value: &tengo.String{Value: e.Error()}}, nil
			} else {
				return &tengo.String{Value: n}, nil
			}
		}},
		"torr": &tengo.UserFunction{Name: "torr", Value: func(args ...tengo.Object) (tengo.Object, error) {
			var lnk []string
			switch len(args) {
			case 3:
				if s, _ := tengo.ToString(args[2]); s != "" {
					lnk = append(lnk, "&img=", url.QueryEscape(s))
				}
				fallthrough
			case 2:
				if s, _ := tengo.ToString(args[1]); s != "" {
					lnk = append(lnk, "&ttl=", url.QueryEscape(s))
				}
				fallthrough
			case 1:
				if s, _ := tengo.ToString(args[0]); s != "" {
					lnk = append([]string{"?link=", url.QueryEscape(s)}, lnk...)
				}
				lnk = append([]string{"content:http://", r.Host, "/msx/torr"}, lnk...)
				return &tengo.String{Value: strings.Join(lnk, "")}, nil
			case 0:
				return &tengo.String{Value: stg.TorrServer}, nil
			default:
				return nil, tengo.ErrWrongNumArguments
			}
		}},
		"player": &tengo.UserFunction{Name: "player", Value: func(args ...tengo.Object) (tengo.Object, error) {
			switch len(args) {
			case 0:
				if stg.Clients[id]&cHTML5X == 0 {
					return &tengo.String{Value: clients[id].Player}, nil
				} else {
					return &tengo.String{Value: "html5x"}, nil
				}
			case 1:
				switch a := args[0].(type) {
				case *tengo.Bool:
					ps := &tengo.Map{Value: make(map[string]tengo.Object)}
					for k, p := range playerProp(r.Host, id, !a.IsFalsy()) {
						ps.Value[k] = &tengo.String{Value: p}
					}
					return ps, nil
				case *tengo.String:
					iv := true
					switch a.Value[:6] {
					case "audio:":
						iv = false
						fallthrough
					case "video:":
						return &tengo.String{Value: playerURL(id, a.Value[6:], iv)}, nil
					default:
						return &tengo.String{Value: playerURL(id, a.Value, iv)[6:]}, nil
					}
				default:
					return nil, tengo.ErrInvalidArgumentType{Name: "first", Expected: "bool/string", Found: args[0].TypeName()}
				}
			default:
				return nil, tengo.ErrWrongNumArguments
			}
		}},
	}
	if stg.Clients[id]&cCompressed == 0 {
		m["compressed"] = tengo.FalseValue
	} else {
		m["compressed"] = tengo.TrueValue
	}
	return
}
