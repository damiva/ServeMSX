package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var logFFmpeg bool
var ffmpegRun *exec.Cmd

func ffmpeg(r *http.Request, src string, args [2][]string) (addr string, err error) {
	if stg.FFmpegCMD != "" || stg.FFmpegPORT != "" {
		clearFFmpeg()
		addr = "http://" + r.Host[:strings.LastIndexByte(r.Host, ':')+1] + stg.FFmpegPORT
		args[0] = append(append(args[0], "-i", src), append(args[1], "-listen", "1", addr)...)
		mutexF.Lock()
		ffmpegRun = exec.Command(stg.FFmpegCMD, args[0]...)
		if logFFmpeg {
			ffmpegRun.Stdout, ffmpegRun.Stderr = os.Stdout, os.Stderr
		} else {
			ffmpegRun.Stdout, ffmpegRun.Stderr = nil, nil
		}
		out.Println(stg.FFmpegCMD, strings.Join(args[0], " "))
		if err = ffmpegRun.Start(); err == nil {
			go waitFFmpeg()
			time.Sleep(time.Second)
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
