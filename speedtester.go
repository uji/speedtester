package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	speedtest "github.com/ujiprog/speedtest-go/pkg"
	"gopkg.in/alecthomas/kingpin.v2"
)

func setTimeout() {
	if *timeoutOpt != 0 {
		timeout = *timeoutOpt
	}
}

var (
	showList   = kingpin.Flag("list", "Show available speedtest.net servers").Short('l').Bool()
	serverIds  = kingpin.Flag("server", "Select server id to speedtest").Short('s').Ints()
	timeoutOpt = kingpin.Flag("timeout", "Define timeout seconds. Default: 10 sec").Short('t').Int()
	timeout    = 10
)

func cron() {
	kingpin.Version("1.0.3")
	kingpin.Parse()

	setTimeout()

	user := speedtest.FetchUserInfo()
	user.Show()

	list := speedtest.FetchServerList(user)
	if *showList {
		list.Show()
		return
	}

	targets := list.FindServer(*serverIds)
	targets.StartTest()
	targets.ShowResult()

	time := time.Now().Format(time.RFC822)
	speed := targets.GetResult()
	download := strconv.FormatFloat(speed.Download, 'f', 4, 64)
	upload := strconv.FormatFloat(speed.Upload, 'f', 4, 64)

	file, err := os.OpenFile("/tmp/export.csv", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Error:", err)
	}
	defer file.Close()

	record := []string{time, download, upload}
	writer := csv.NewWriter(file)
	err = writer.Write(record)
	if err != nil {
		log.Fatal("Error:", err)
	}
	writer.Flush()
}

func main() {
	go func() {
		for true {
			cron()
			time.Sleep(15 * time.Minute)
		}
	}()
	http.ListenAndServe(":8080", http.FileServer(http.Dir("/tmp")))
}
