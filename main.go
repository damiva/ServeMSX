/*
MIT License

Copyright (c) 2021 damiva

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package main

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	out     = log.New(os.Stdout, "(i) ", log.Flags())
	stg     = new(settings)
	mutex   = new(sync.Mutex)
	started = time.Now()
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	defer log.Fatalln(recover())
	log.SetPrefix("<!> ")
	la := ":8008"
	for _, a := range os.Args[1:] {
		switch a {
		case "-t":
			out.SetFlags(0)
			log.SetFlags(0)
		case "-i":
			log.SetPrefix("")
			out.SetOutput(ioutil.Discard)
		case "-s":
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		case "-d":
			p, e := os.Executable()
			check(e)
			check(os.Chdir(filepath.Dir(p)))
		default:
			la = a
		}
	}
	check(stg.load())
	out.Println(Name, "v.", Vers, "listening at", la)
	check(http.ListenAndServe(la, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				switch v := e.(type) {
				case int:
					http.Error(w, http.StatusText(v), v)
				case string:
					http.Error(w, v, 500)
				case error:
					http.Error(w, v.Error(), 500)
				default:
					http.Error(w, "Unknown error!", 500)
				}
				log.Println(e)
			}
		}()
		out.Println(r.Method, ":", r.RequestURI)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		if r.Method != "OPTIONS" {
			w.Header().Set("Server", Name+"/"+Vers)
			http.DefaultServeMux.ServeHTTP(w, r)
		}
	})))
}
