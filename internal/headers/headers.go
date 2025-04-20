package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	endIndex := bytes.Index(data, []byte("\r\n"))
	if endIndex == -1 {
		return 0, false, nil
	}
	if endIndex == 0 {
		return 0, true, nil
	}

	field := strings.TrimSpace(string(data[:endIndex]))
	fieldArray := strings.SplitN(field, ":", 2)

	if len(fieldArray) != 2 {
		return 0, false, errors.New("header field should contain a colon")
	}

	fieldKey, fieldValue := fieldArray[0], fieldArray[1]

	if fieldKey[len(fieldKey)-1] == ' ' {
		return 0, false, errors.New("white space between field key and colon")
	}

	h[fieldKey] = strings.TrimSpace(fieldValue)

	return endIndex + 2, done, err
}

func NewHeaders() Headers {
	return make(Headers)
}
