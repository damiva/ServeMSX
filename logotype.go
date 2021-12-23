package main

import (
	"net/http"
)

const (
	logoIcon = `
	<path d="M22 4v12h-20v-12h20zm2-2h-24v16h24v-16zm-7 18h-10v2h10v-2z"/><polygon class="green" points="9,6 17,10 9,14"/>`
	logoText = `
	<text x="100%" y="50%" font-size="14px" font-weight="bold" dominant-baseline="middle" textLength="60px" lengthAdjust="spacingAndGlyphs"><tspan class="green">Serve</tspan>MSX</text>
	<text x="100%" y="99%" font-size="3px" dominant-baseline="bottom" class="green">© damiva</text>`
)

func init() {
	http.HandleFunc("/logo.svg", logotype)
	http.HandleFunc("/logotype.svg", logotype)
}

func logotype(w http.ResponseWriter, r *http.Request) {
	s, f, c, t := "24", "", "", r.URL.Path[5] == 't'
	if t {
		f, c, s = ` fill="silver"`, "lime", "90"
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte("<?xml version=\"1.0\" standalone=\"yes\"?>\n"))
	w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ` + s + ` 24"` + f + `>
	<style> .green{fill: ` + c + `green} text{font-family: sans-serif; text-anchor: end;}</style>`))
	w.Write([]byte(logoIcon))
	if t {
		w.Write([]byte(logoText))
	}
	w.Write([]byte("\n</svg>"))
}
