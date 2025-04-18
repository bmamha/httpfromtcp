package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	ParserState requestState
}

type requestState int

const bufferSize = 8

const (
	requestStateInitialized requestState = iota
	requestStateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	// control current index of bytes starting from 0
	readToIndex := 0
	// initialize Request struct pointer
	r := &Request{ParserState: requestStateInitialized}
	// initialize receiver buffer - for now 8 bytes
	p := make([]byte, bufferSize)
	// keep reading till state of request parsing is done.
	// done in this context means we have reached '\r\n' bytes of the request from our Reader
	for r.ParserState != requestStateDone {
		// increase buffer size if our read Index exceeds it
		if readToIndex >= len(p) {
			newP := make([]byte, len(p)*2)
			copy(newP, p)
			p = newP
		}

		numBytesRead, err := reader.Read(p[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				r.ParserState = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := r.parse(p[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(p, p[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return r, nil
}

func parseRequestLine(d []byte) (*RequestLine, int, error) {
	endIndex := bytes.Index(d, []byte("\r\n"))
	if endIndex == -1 {
		return nil, 0, nil
	}

	requestLine := string(d[:endIndex])
	fmt.Println(requestLine)
	httpParts := strings.Split(requestLine, " ")
	if len(httpParts) != 3 {
		return nil, 0, errors.New("three parameters required for request line")
	}
	fmt.Println(httpParts)
	method, target, httpVersion := httpParts[0], httpParts[1], httpParts[2]
	fmt.Println(method, target, httpVersion)
	if method != strings.ToUpper(method) {
		return nil, 0, errors.New("method is not in upper case")
	}

	if httpVersion != "HTTP/1.1" {
		return nil, 0, errors.New("HTTP version is not HTTP/1.1")
	}

	version := strings.Split(httpVersion, "/")[1]
	rl := RequestLine{version, target, method}

	return &rl, endIndex, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.ParserState == requestStateInitialized {
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		if n > 0 {
			r.RequestLine = *rl
			r.ParserState = requestStateDone
			return n, nil
		}
	}
	if r.ParserState == requestStateDone {
		return 0, errors.New("trying to read data in a done state")
	}

	return 0, errors.New("unknown state")
}
