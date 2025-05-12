package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bmamha/httpfromtcp/internal/request"
	"github.com/bmamha/httpfromtcp/internal/response"
	"github.com/bmamha/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handlerfunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handlerfunc(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		responseWriter(w, 400, "Your request honestly kinda sucked")
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		responseWriter(w, 500, "Ok, you know what? This one is on me.")
		return
	}

	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		proxyHandler(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/video" {
		videoHandler(w, req)
		return
	}

	responseWriter(w, 200, "Your request was an absolute banger.")
}

func responseWriter(w *response.Writer, s response.StatusCode, message string) {
	err := w.WriteStatusLine(s)
	if err != nil {
		fmt.Printf("unable to write response status line: %v", err)
		return
	}
	w.Headers.Set("Content-Type", "text/html")
	_, err = w.Write([]byte(message))
	if err != nil {
		fmt.Printf("unable to write response: %v", err)
	}
}

func proxyHandler(w *response.Writer, req *request.Request) {
	trimmedPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	resp, err := http.Get("https://httpbin.org/" + trimmedPath)
	if err != nil {
		fmt.Printf("unable to fetch data from http bin: %v", err)
		return
	}
	defer resp.Body.Close()

	w.Headers = response.GetDefaultChunkHeaders()

	w.WriteStatusLine(200)
	w.WriteHeaders()

	buffer := make([]byte, 1024)

	var tracker bytes.Buffer

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			tracker.Write(buffer[:n])
			w.WriteChunkedBody(buffer[:n])
		}

		if err == io.EOF {
			_, err := w.WriteChunkedBodyDone()
			if err != nil {
				fmt.Printf("Chunk ending not written: %v", err)
			}
			trailers := response.GetDefaultTrailers()

			xContentSHA256 := fmt.Sprintf("%x", sha256.Sum256(tracker.Bytes()))
			trailers.Set("X-Content-SHA256", xContentSHA256)

			xContentLength := fmt.Sprintf("%d", len(tracker.Bytes()))
			trailers.Set("X-Content-Length", xContentLength)

			err = w.WriteTrailers(trailers)
			if err != nil {
				fmt.Printf("trailer not written: %v", err)
			}

			break
		}

		if err != nil {
			fmt.Printf("unable to read data from response: %v", err)
			break
		}

	}
}

func videoHandler(w *response.Writer, req *request.Request) {
	videoFile, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		responseWriter(w, 400, "Error processing video file")
		return
	}
	err = w.WriteStatusLine(200)
	if err != nil {
		fmt.Printf("unable to write status line: %v", err)
		return
	}
	w.Headers.Set("Content-Type", "video/mp4")
	w.Headers.Set("Content-Length", fmt.Sprintf("%d", len(videoFile)))
	w.Headers.Set("Accept-Ranges", "bytes")
	err = w.WriteHeaders()
	if err != nil {
		fmt.Printf("Unable to write headers: %v", err)
		return
	}
	_, err = w.Writer.Write(videoFile)
	if err != nil {
		fmt.Printf("unable to write video: %v", err)
	}
	w.State = response.WritingStatusLine
}
