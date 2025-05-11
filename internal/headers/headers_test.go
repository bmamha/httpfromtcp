package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Lower case keys
	headers = NewHeaders()
	data = []byte("Authorization: password\r\n")
	_, done, _ = headers.Parse(data)
	assert.False(t, done)
	assert.Equal(t, "password", headers["authorization"])

	// Test: Invalid characters
	headers = NewHeaders()
	data = []byte("H@st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: duplicate headers keys
	headers = NewHeaders()
	data = []byte("Set-Person: Lemma\r\n")
	_, _, _ = headers.Parse(data)
	assert.Equal(t, "Lemma", headers["set-person"])
	data = []byte("Set-Person: Tony\r\n")
	_, _, _ = headers.Parse(data)
	assert.Equal(t, "Lemma, Tony", headers["set-person"])

	// Test: No ':' between key and field
	headers = NewHeaders()
	data = []byte("Host localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
