package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const pthFFmpeg = "/msx/ffmpeg/"

var ffmpegLog bool
var ffmpegRun *exec.Cmd

func init() {
	http.HandleFunc(pthFFmpeg, func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Clean(strings.TrimPrefix(r.URL.Path, pthFFmpeg))
		if d, n := filepath.Split(p); d == "" || d == "." || d == "/" {
			p = filepath.Join(tempDir, n)
			http.ServeFile(w, r, p)
			if e := os.Remove(p); e != nil {
				log.Println(e)
			}
		} else if d = strings.ToLower(filepath.Ext(n)); strings.Contains(extPic, d) {
			d = filepath.Join(tempDir, n) + ".jpeg"
			if c := ffmpegCmd(false, "-hide_banner", "-i", p, "-y", "-vf", "scale=408:-1", d); c != nil && c.Run() == nil {
				http.ServeFile(w, r, d)
				if e := os.Remove(d); e != nil {
					log.Println(e)
				}
				return
			}
			http.ServeFile(w, r, p)
		} else if strings.Contains(extAud, d) {
			var a []string
			n += ".jpeg"
			d = filepath.Join(tempDir, n)
			if c := ffmpegCmd(false, "-hide_banner", "-i", p, "-y", "-an", "-vf", "scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2", d); c != nil && c.Run() == nil {
				a = append(a, "player:background:http://"+r.Host+pthFFmpeg+url.PathEscape(n))
			}
			if c := ffmpegCmd(true, "-v", "error", "-print_format", "json", "-show_entries", "format_tags=artist,title,album,genre", p); c != nil {
				if b, e := c.Output(); e == nil {
					var i struct {
						Format struct {
							Tags struct{ Title, Artist, Album, Genre string }
						}
					}
					if e = json.Unmarshal(b, &i); e != nil {
						log.Println(stg.FFprobe, "decoding result error:", e)
					} else {
						var t []string
						n = filepath.Base(p)
						if i.Format.Tags.Artist != "" {
							t = append(t, "{ico:person} "+i.Format.Tags.Artist)
						}
						if i.Format.Tags.Album != "" {
							t = append(t, "{ico:album} "+i.Format.Tags.Album)
						}
						if i.Format.Tags.Genre != "" {
							t = append(t, "{ico:music-note} "+i.Format.Tags.Genre)
						}
						if i.Format.Tags.Title != "" {
							a = append(a, "player:label:content:"+i.Format.Tags.Title)
						} else {
							a = append(a, "player:label:content:"+n[:strings.LastIndexByte(n, '.')])
						}
						t = append(t, "{col:msx-white-soft}{ico:attach-file} "+n)
						if d = filepath.Dir(p); strings.IndexRune(d, filepath.Separator) > 0 {
							t = append(t, "{ico:folder-open} "+filepath.Base(d))
						}
						a = append(a, "player:info:text:{col:msx-white}"+strings.Join(t, "{br}"))
					}
				} else {
					log.Println(e)
				}
			}
			svcAnswer(w, "["+strings.Join(a, "|")+"]", nil)
		} else {
			panic(400)
		}
	})
}
func ffmpegCmd(probe bool, args ...string) (c *exec.Cmd) {
	n := stg.FFmpeg
	if probe {
		n = stg.FFprobe
	}
	if n != "" {
		c = exec.Command(n, args...)
		c.Stderr = os.Stderr
		if !probe && ffmpegLog {
			c.Stdout = os.Stdout
		}
		out.Println(n, strings.Join(args, " "))
	}
	return
}
func ffmpeg(r *http.Request, src string, args [2][]string) (addr string, err error) {
	if stg.FFstream != "" {
		addr = "http://" + r.Host[:strings.LastIndexByte(r.Host, ':')+1] + stg.FFstream
		args[0] = append(append(args[0], "-hide_banner", "-i", src), append(args[1], "-f", "mpegts", "-listen", "1", addr)...)
		clearFFmpeg()
		mutexF.Lock()
		if ffmpegRun = ffmpegCmd(false, args[0]...); ffmpegRun != nil {
			if err = ffmpegRun.Start(); err == nil {
				go waitFFmpeg()
				time.Sleep(time.Second * time.Duration(performSecs))
			} else {
				mutexF.Unlock()
			}
		} else {
			mutexF.Unlock()
		}
	}
	return
}
func checkFFmpeg(probe bool) (e error) {
	c := stg.FFmpeg
	if probe {
		c = stg.FFprobe
	}
	if c != "" {
		c, e = exec.LookPath(c)
		if probe {
			stg.FFprobe = c
		} else {
			stg.FFmpeg = c
		}
	}
	return
}
func clearFFmpeg() {
	if ffmpegRun != nil {
		if e := ffmpegRun.Process.Kill(); e != nil {
			log.Println("kill ffmpeg error:", e)
		}
	}
}
func waitFFmpeg() {
	if e := ffmpegRun.Wait(); e != nil {
		log.Println("ffmpeg:", e)
	}
	ffmpegRun = nil
	mutexF.Unlock()
}
