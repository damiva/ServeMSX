package tengohttp

import (
	"net/http"

	"github.com/d5/tengo/v2"
)

type server struct {
	w http.ResponseWriter
	r *http.Request
	c *http.Client
	//	p string
	h bool
}

func GetModuleMAP(w http.ResponseWriter, r *http.Request, c *http.Client, vars map[string]tengo.Object) map[string]tengo.Object {
	if c == nil {
		c = &http.Client{}
	}
	s := &server{w, r, c, false}
	ret := map[string]tengo.Object{
		"proto":       &tengo.String{Value: r.Proto},
		"method":      &tengo.String{Value: r.Method},
		"host":        &tengo.String{Value: r.Host},
		"remote_addr": &tengo.String{Value: r.RemoteAddr},
		"header":      vals2map(r.Header),
		"uri":         &tengo.String{Value: r.RequestURI},
		"raw_query":   &tengo.String{Value: r.URL.RawQuery},
		"read":        &tengo.UserFunction{Name: "read", Value: s.read},
		"write":       &tengo.UserFunction{Name: "write", Value: s.write},
		"request":     &tengo.UserFunction{Name: "request", Value: s.request},
		"uri_encode":  &tengo.UserFunction{Name: "uri_encode", Value: encode},
		"uri_decode":  &tengo.UserFunction{Name: "uri_decode", Value: decode},
		"url_parse":   &tengo.UserFunction{Name: "url_parse", Value: s.parse},
		"url_resolve": &tengo.UserFunction{Name: "url_resolve", Value: s.resolve},
		"query_parse": &tengo.UserFunction{Name: "query_parse", Value: query},
	}
	for k, v := range vars {
		ret[k] = v
	}
	return ret
}
