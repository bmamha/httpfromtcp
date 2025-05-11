package response

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net"
	"strconv"

	"github.com/bmamha/httpfromtcp/internal/headers"
)

type StatusCode int

type writerState int

type Writer struct {
	State         writerState
	Writer        net.Conn
	StatusCode    StatusCode
	StatusMessage string
	Headers       headers.Headers
	Protocol      string
	Body          []byte
}

const (
	StatusCodeSuccessful          StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

const (
	WritingStatusLine writerState = iota
	WritingHeaders
	WritingBody
)

const chunkedBufferSize = 32

const Protocol = "HTTP/1.1"

func NewWriter(conn net.Conn) *Writer {
	return &Writer{
		State:    WritingStatusLine,
		Writer:   conn,
		Protocol: Protocol,
		Headers:  GetDefaultHeaders(0),
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var err error
	statusLine := ""
	if w.State == WritingStatusLine {
		w.StatusCode = statusCode
		switch w.StatusCode {
		case StatusCodeSuccessful:
			w.StatusMessage = "OK"
		case StatusCodeBadRequest:
			w.StatusMessage = "Bad Request"
		case StatusCodeInternalServerError:
			w.StatusMessage = "Internal Server Error"
		default:
			w.StatusMessage = ""
		}

		statusLine = fmt.Sprintf("%s %d %s\r\n", w.Protocol, w.StatusCode, w.StatusMessage)
		_, err = w.Writer.Write([]byte(statusLine))
		w.State = WritingHeaders
		return err
	}
	return errors.New("attempting to write status line in a different state")
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := make(headers.Headers)
	headers["Content-Length"] = fmt.Sprintf("%d", contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"
	return headers
}

func GetDefaultChunkHeaders() headers.Headers {
	headers := make(headers.Headers)
	headers["Content-Type"] = "text/plain"
	headers["Transfer-Encoding"] = "chunked"
	return headers
}

func GetDefaultTrailers() headers.Headers {
	headers := make(headers.Headers)
	headers["X-Content-SHA256"] = ""
	headers["X-Content-Length"] = ""
	return headers
}

func (w *Writer) WriteHeaders() error {
	fmt.Printf("Header in state:%d", w.State)
	if w.State == WritingHeaders {
		for k, v := range w.Headers {
			_, err := w.Writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
			if err != nil {
				return err
			}
		}
		_, err := w.Writer.Write([]byte("\r\n"))
		w.State = WritingBody
		return err
	}
	return errors.New("not in Writing Headers state")
}

func (w *Writer) WriteBody(data []byte) (int, error) {
	if w.State == WritingBody {
		n, err := w.Writer.Write(data)
		if err != nil {
			return 0, fmt.Errorf("unable to write body: %v", err)
		}
		w.State = WritingStatusLine
		return n, nil

	}

	return 0, errors.New("not in the right state to write body")
}

func (w *Writer) Write(p []byte) (int, error) {
	var buf bytes.Buffer
	w.Body = p
	t, _ := template.ParseFiles("template/edit.html")
	err := t.Execute(&buf, w)
	if err != nil {
		return 0, fmt.Errorf("unable to execute template to buffer: %v", err)
	}
	content := buf.Bytes()
	contentLength := len(content)

	w.Headers.Set("Content-Length", strconv.Itoa(contentLength))
	err = w.WriteHeaders()
	if err != nil {
		return 0, fmt.Errorf("unable to write headers: %v", err)
	}
	n, err := w.WriteBody(content)
	if err != nil {
		return 0, fmt.Errorf("unable to parse body: %v", err)
	}
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	size := fmt.Sprintf("%x\r\n", len(p))
	w.Writer.Write([]byte(size))
	w.Writer.Write(p)
	w.Writer.Write([]byte("\r\n"))

	return len(p), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.Writer.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for k, v := range h {
		fmt.Println(k, v)
		_, err := w.Writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.Writer.Write([]byte("\r\n"))
	return err
}
