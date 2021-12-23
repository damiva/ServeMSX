package main

import "net/http"

func init() {
	http.HandleFunc("/logo.svg", logotype)
	http.HandleFunc("/logotype.svg", logotype)
}

func logotype(w http.ResponseWriter, r *http.Request) {
	s, f, c, t := "24", "", "", r.URL.Path[5] == 't'
	if t {
		f, c, s = ` fill="white"`, "lime", "100"
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte(`<?xml version="1.0" standalone="yes"?>
<!-- ` + Name + " v. " + Vers + `  © 2021 damiva -->
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ` + s + ` 24"` + f + `>
	<path d="M22 4v12h-20v-12h20zm2-2h-24v16h24v-16zm-7 18h-10v2h10v-2z"/><polygon points="9,6 17,10 9,14" fill="` + c + `green"/>`))
	if t {
		w.Write([]byte(`
	<text x="30" y="50%" textLength="70" style="font: bold 14px sans-serif; dominant-baseline: middle;"><tspan fill="` + c + `green">Serve</tspan>MSX</text>
	<text x="100" y="24" style="font: 4px sans-serif; dominant-baselin: bottom; text-anchor: end; fill: ` + c + `">© 2021 damiva</text>`))
	}
	w.Write([]byte("\n</svg>"))
}
