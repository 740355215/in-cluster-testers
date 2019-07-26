package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	// add help
	var h bool
	flag.BoolVar(&h, "h", false, "this is help")
	var message string
	flag.StringVar(&message, "message", "init message", "the message will be sent to sidecar")
	var exitcode int
	flag.IntVar(&exitcode, "exitcode", 0, "testcase exitcode.(0 -> passed, others -> failed)")
	flag.Parse()

	if h || len(os.Args) <= 1 {
		flag.Usage()
		return
	}

	// get conf from ConfigMap
	fmt.Printf("configmap exmaple is %v\n", os.Getenv("example"))

	fmt.Printf("input exitcode is %v\n", exitcode)
	time.Sleep(time.Duration(3) * time.Second)

	sidecarUrl := "http://127.0.0.1:8099/message"
	resp, err := http.Post(sidecarUrl, "text/plain", strings.NewReader(message))
	if err != nil {
		fmt.Printf("failed to send message to sidecar: %v\n", err)
		os.Exit(20)
		return
	}
	defer resp.Body.Close()

	os.Exit(exitcode)
	return
}
