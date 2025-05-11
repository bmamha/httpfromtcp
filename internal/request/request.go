package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/bmamha/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine    RequestLine
	Headers        headers.Headers
	ParserState    requestState
	Body           []byte
	bodyLengthRead int
}

type requestState int

const bufferSize = 124
const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBodyState
	requestStateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	log.Println("initiating request parsing")
	readToIndex := 0
	r := &Request{ParserState: requestStateInitialized, Headers: headers.NewHeaders()}
	buf := make([]byte, bufferSize)
	for r.ParserState != requestStateDone {
		log.Printf("Currently in %d", r.ParserState)
		// increase buffer size if our read Index exceeds it
		if readToIndex >= len(buf) {
			log.Println("increasing buffer index")
			newBuffer := make([]byte, len(buf)*2)
			copy(newBuffer, buf)
			buf = newBuffer
		}
		log.Println("Buffer size is greater than reading index")

		numBytesRead, err := reader.Read(buf[readToIndex:])
		log.Printf("number of bytes read: %d", numBytesRead)
		log.Println(string(buf))
		if err != nil {
			if errors.Is(err, io.EOF) {
				if r.ParserState != requestStateDone && r.Body != nil {
					return nil, errors.New("body shorter than Content-Length")
				}

				break
			}
			return nil, err
		}
		readToIndex += numBytesRead
		numBytesParsed, err := r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
		log.Printf("Number of reading index: %d", readToIndex)
	}
	log.Println("request parsing ended")
	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.ParserState != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func parseRequestLine(d []byte) (*RequestLine, int, error) {
	endIndex := bytes.Index(d, []byte("\r\n"))
	if endIndex == -1 {
		return nil, 0, nil
	}

	requestLine := string(d[:endIndex])
	httpParts := strings.Split(requestLine, " ")
	if len(httpParts) != 3 {
		return nil, 0, errors.New("three parameters required for request line")
	}
	method, target, httpVersion := httpParts[0], httpParts[1], httpParts[2]
	if method != strings.ToUpper(method) {
		return nil, 0, errors.New("method is not in upper case")
	}

	if httpVersion != "HTTP/1.1" {
		return nil, 0, errors.New("HTTP version is not HTTP/1.1")
	}

	version := strings.Split(httpVersion, "/")[1]
	rl := RequestLine{version, target, method}

	return &rl, endIndex + 2, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	if r.ParserState == requestStateInitialized {
		log.Println("Initalized parser state")
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		log.Println("request line parsed")
		r.RequestLine = *rl
		r.ParserState = requestStateParsingHeaders
		log.Printf("request target is: %s", r.RequestLine.RequestTarget)
		log.Println("transitioning to parsing headers")
		return n, nil

	}

	if r.ParserState == requestStateParsingHeaders {
		log.Println("parsing headers state")
		headers := r.Headers
		n, done, err := headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			return 0, nil
		}
		if n > 0 {
			if done {
				r.ParserState = requestStateParsingBodyState
			}
			r.Headers = headers
			return n, nil
		}
	}

	if r.ParserState == requestStateParsingBodyState {
		log.Println("parsing body state")
		contentLength, ok := r.Headers.Get("content-length")

		if !ok {
			r.ParserState = requestStateDone
			return len(data), nil
		}
		contentLengthInt, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, fmt.Errorf("invalid Content-Length: %v", err)
		}

		remaining := contentLengthInt - r.bodyLengthRead
		if remaining <= 0 {
			r.ParserState = requestStateDone
			return 0, nil
		}
		bytesToTake := min(remaining, len(data))
		r.Body = append(r.Body, data[:bytesToTake]...)

		r.bodyLengthRead += bytesToTake
		log.Println(r.bodyLengthRead)
		log.Println(contentLengthInt)

		if r.bodyLengthRead == contentLengthInt {
			r.ParserState = requestStateDone
			fmt.Println("All data has been consumed")
		}

		return bytesToTake, nil
	}

	if r.ParserState == requestStateDone {
		return 0, errors.New("error trying to read data from done state")
	}

	return 0, errors.New("unknown state")
}
