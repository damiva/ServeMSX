package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var logFFmpeg bool
var ffmpegRun *exec.Cmd

func init() {
	http.HandleFunc("/msx/temp/", func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(tempDir, filepath.Clean(strings.TrimPrefix(r.URL.Path, "/msx/temp/")))
		http.ServeFile(w, r, p)
		out.Println("Remove", p, "error:", os.Remove(p))
	})
}
func ffmpeg(r *http.Request, src string, args [2][]string) (addr string, err error) {
	if stg.FFmpegCMD != "" || stg.FFmpegPORT != "" {
		clearFFmpeg()
		addr = "http://" + r.Host[:strings.LastIndexByte(r.Host, ':')+1] + stg.FFmpegPORT + "/stream.ts"
		args[0] = append(append(args[0], "-i", src), append(args[1], "-listen", "1", addr)...)
		mutexF.Lock()
		ffmpegRun = exec.Command(stg.FFmpegCMD, args[0]...)
		ffmpegRun.Stderr = os.Stderr
		if logFFmpeg {
			ffmpegRun.Stdout = os.Stdout
		} else {
			ffmpegRun.Stdout = nil
		}
		out.Println(stg.FFmpegCMD, strings.Join(args[0], " "))
		if err = ffmpegRun.Start(); err == nil {
			go waitFFmpeg()
			time.Sleep(time.Second * time.Duration(performSecs))
		}
	}
	return
}
func ffmpegPic(src string) (dst string, err error) {
	if stg.FFmpegCMD != "" {
		dst = filepath.Base(src)
		x, a := strings.ToLower(filepath.Ext(dst)), []string{"-i", src}
		if strings.Contains(extPic, x) {
			a = append(a, []string{"-vf", "scale=480:-1"}...)
		} else if strings.Contains(extAud, x) {
			a = append(a, []string{"-an", "-vf", "scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2"}...)
		} else {
			err = errors.New("wrong file type")
		}
		if err == nil {
			dst += ".jpeg"
			a = append(a, filepath.Join(tempDir, dst))
			out.Println(stg.FFmpegCMD, strings.Join(a, " "))
			os.Remove(a[len(a)-1])
			err = exec.Command(stg.FFmpegCMD, a...).Run()
		}
	}
	return
}
func checkFFmpeg() (e error) {
	c := stg.FFmpegCMD
	if c == "" {
		c = "ffmpeg"
	}
	if c, e = exec.LookPath(c); e == nil {
		stg.FFmpegCMD = c
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
