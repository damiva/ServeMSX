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
	"strings"
	"sync"
	"time"
)

var (
	out     = log.New(ioutil.Discard, "(i) ", log.Flags())
	mutexR  = new(sync.Mutex)
	mutexF  = new(sync.Mutex)
	mypath  string
	started = time.Now()
)

func restart() {
	log.Println("Restarting...")
	clearFFmpeg()
	mutexR.Lock()
	os.Exit(0)
}
func check(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	var e error
	defer log.Fatalln(recover())
	mypath, e = os.Executable()
	check(e)
	for _, a := range os.Args[1:] {
		switch a {
		case "-t":
			out.SetFlags(0)
			log.SetFlags(0)
		case "+i":
			log.SetPrefix("<!> ")
			out.SetOutput(os.Stdout)
		case "+f":
			logFFmpeg = true
		case "-s":
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		case "-d":
			check(os.Chdir(filepath.Dir(mypath)))
		default:
			if a[0] != '-' {
				server.Addr = a
			}
		}
	}
	check(stg.load())
	if e = os.Remove(mypath + ".old"); !os.IsNotExist(e) {
		out.Println(Name, "has been updated!")
		if e != nil {
			log.Println(e)
		}
		if _, n, _, e := getDic(); n != "" {
			if i, e := gitRelease("", ""); e == nil {
				for _, a := range i.Assets {
					if strings.HasSuffix(a.Name, ".json") && strings.HasPrefix(a.Name, n) {
						check(download(a.Browser_download_url, pthDic, nil))
						break
					}
				}
			} else {
				log.Println(e)
			}
		} else if e != nil {
			log.Println(e)
		}
	}
	if e = checkFFmpeg(); e != nil {
		log.Println(e)
	}
	out.Println(Name, "v.", Vers, "listening at", server.Addr)
	check(server.ListenAndServe())
}
