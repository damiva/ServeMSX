package main

import (
	"encoding/json"
	"net/http"
)

var translation = map[string][2]string{
	"opt":       {"Options", "Опции"},
	"rmv":       {"Remove", "Удалить"},
	"goto":      {"Open the source", "Открыть источник"},
	"drop":      {"Drop", "Сбросить"},
	"history":   {"Continue playing", "Продолжить..."},
	"video":     {"My video", "Моё видео"},
	"music":     {"My music", "Моя музыка"},
	"torrents":  {"My torrents", "Мои торренты"},
	"torrent0":  {"Torrent added", "Торрент добавлен"},
	"torrent1":  {"Torrent getting info", "Торрент получает информацию"},
	"torrent2":  {"Torrent preload", "Торрент подгружается"},
	"torrent3":  {"Torrent working", "Торрент работает"},
	"torrent4":  {"Torrent closed", "Торрент закрыт"},
	"torrent5":  {"Torrent in database", "Торрент в базе данных"},
	"rmvd":      {"Deleted!", "Удалено!"},
	"dropd":     {"Dropped!", "Сброшено!"},
	"update":    {"Check{br}updates", "Проверить{br}обновления"},
	"rus":       {"Translate to Russian", "Русский перевод"},
	"addrInput": {"The Address of", "Адрес"},
	"about": {
		"{txt:msx-white:" + Name + "} is the software that allows to view user's content and dvelop user's plugins.{br}{txt:msx-white:" + Name + "} does not provide any video/audio content itself!",
		"{txt:msx-white:" + Name + "} это ПО, позволяющее пользователю просматривать свой контент и разрабатывать свои плагины.{br}{txt:msx-white:" + Name + "} не содержит в себе какого либо видео/аудио контента!",
	},
}

func init() {
	http.HandleFunc("/msx/dic.json", func(w http.ResponseWriter, r *http.Request) {
		dic := struct {
			N string            `json:"name"`
			V string            `json:"version"`
			P map[string]string `json:"properties"`
		}{N: "English", V: Vers, P: make(map[string]string)}
		i := 0
		if stg.Russian {
			i = 1
			download("http://msxplayer.ru/assets/ru.json", &dic, nil)
		}
		for k, v := range translation {
			dic.P[k] = v[i]
		}
		j := json.NewEncoder(w)
		j.SetIndent("", "  ")
		j.Encode(&dic)
	})
}
