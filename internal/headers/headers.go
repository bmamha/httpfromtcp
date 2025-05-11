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
		return 2, true, nil
	}

	field := strings.TrimSpace(string(data[:endIndex]))
	fieldArray := strings.SplitN(field, ":", 2)

	if len(fieldArray) != 2 {
		return 0, false, errors.New("header field should contain a colon")
	}

	fieldKey, fieldValue := fieldArray[0], fieldArray[1]

	if strings.ContainsRune(fieldKey, ' ') {
		return 0, false, errors.New("white space between field key and colon")
	}
	if !isValidFieldString(fieldKey) {
		return 0, false, errors.New("invalid characters in field key")
	}

	fieldKey = strings.ToLower(fieldKey)

	if _, exists := h[fieldKey]; exists {
		h[fieldKey] += ", " + strings.TrimSpace(fieldValue)
	} else {
		h[fieldKey] = strings.TrimSpace(fieldValue)
	}
	return endIndex + 2, false, err
}

func NewHeaders() Headers {
	return make(Headers)
}

func isValidFieldString(input string) bool {
	for _, r := range input {
		if !((r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			(strings.ContainsRune("!#$%&'*+.^_`|~-", r))) {
			return false
		}
	}
	return len(input) > 0
}

func (h Headers) Get(key string) (string, bool) {
	v, ok := h[strings.ToLower(key)]

	return v, ok
}

func (h Headers) Set(key, value string) {
	h[key] = value
}
