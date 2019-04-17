package main

import (
	"fmt"
	"image/jpeg"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"runtime"
	"strconv"
	"time"

	"net/http/pprof"
	_ "net/http/pprof"
)

func handleScreenshot(responseWriter http.ResponseWriter, request *http.Request) {
	log.Printf("Start request %s", request.URL)

	mimeWriter := multipart.NewWriter(responseWriter)

	log.Printf("Boundary: %s", mimeWriter.Boundary())

	contentType := fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", mimeWriter.Boundary())
	responseWriter.Header().Add("Content-Type", contentType)

	disp, _ := strconv.Atoi(request.URL.Query().Get("display"))
	i := 0
	info, _ := getMonitors()
	capturer, err := NewCapturer(info[disp])
	if err != nil {
		panic(err)
	}
	for {
		frameStartTime := time.Now()
		partHeader := make(textproto.MIMEHeader)
		partHeader.Add("Content-Type", "image/jpeg")

		partWriter, partErr := mimeWriter.CreatePart(partHeader)
		if nil != partErr {
			log.Printf(partErr.Error())
			break
		}

		img, err := capturer.Capture()
		if err != nil {
			panic(err)
		}
		log.Printf("Capture time: %s", time.Now().Sub(frameStartTime))
		jpeg.Encode(partWriter, img, &jpeg.Options{
			Quality: 100,
		})
		frameDuration := time.Now().Sub(frameStartTime)
		fps := float64(time.Second) / float64(frameDuration)
		log.Printf("Frame time: %s (%.2f)", frameDuration, fps)
		i++
		if i%20 == 0 {
			runtime.GC()
		}
	}

}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/mjpeg", handleScreenshot)
	mux.HandleFunc("/displays", func(responseWriter http.ResponseWriter, request *http.Request) {
		info, _ := getMonitors()
		fmt.Fprintf(responseWriter, "%d", len(info))
	})
	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	http.ListenAndServe("0.0.0.0:4000", mux)
}
