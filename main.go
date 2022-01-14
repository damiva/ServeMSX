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
	mutex   = new(sync.Mutex)
	mypath  string
	started = time.Now()
	//signals = make(chan os.Signal, 1)
)

/*
func init() {
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		s := <-signals
		if stg.toSave {
			if e := stg.save(); e != nil {
				log.Println("Savings settings error:", e)
			}
		}
		mutex.Lock()
		if s == syscall.SIGABRT {
			log.Println("Restarting...")
			os.Exit(0)
		}
		log.Println("Closing server, because of OS signal catched:", s)
		if e := server.Close(); e != nil {
			log.Fatalln("Closing error:", e)
		}
	}()
}
*/
func restart() {
	mutex.Lock()
	log.Println("Restarting...")
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
	log.SetPrefix("<!> ")
	mypath, e = os.Executable()
	check(e)
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
			check(os.Chdir(filepath.Dir(mypath)))
		default:
			if a[0] != '-' {
				server.Addr = a
			}
		}
	}
	check(stg.load())
	if e = os.Remove(mypath + ".old"); e == nil {
		out.Println(Name, "has been updated!")
	} else if !os.IsNotExist(e) {
		log.Println(e)
	}
	out.Println(Name, "v.", Vers, "listening at", server.Addr)
	check(server.ListenAndServe())
}
