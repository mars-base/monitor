package utils

import (
	"bytes"
	"strings"

	"github.com/google/uuid"
	nanoid "github.com/matoous/go-nanoid/v2"
)

func HasSubStr(s string, sub string) bool {
	return strings.Contains(s, sub)
}

// index of sub string, -1 means not found
func IndexOfSubStr(s string, sub string) int {
	return strings.Index(s, sub)
}

func Split(s string, char string) []string {
	slice := []string{}
	if len(s) > 0 && len(char) > 0 && IndexOfSubStr(s, char) >= 0 {
		slice = strings.Split(s, char)
	}
	return slice
}

func Join(s []string, char string) string {
	return strings.Join(s, char)
}

func FastAppendString(origin string, ss ...string) string {
	var buffer bytes.Buffer
	buffer.WriteString(origin)
	for _, s := range ss {
		buffer.WriteString(s)
	}
	return buffer.String()
}

// UUID is a 128-bit value represented in hexadecimal,
// in the format of b480ef1d-fa18-47c3-b0ed-4e4135f7652e
func UUID() string {
	return uuid.New().String()
}

// default length is 21
func NanoID(length ...int) string {
	defaultLen := 21
	if len(length) > 0 {
		defaultLen = length[0]
	}
	id, err := nanoid.New(defaultLen)
	if err != nil {
		return ""
	}
	return id
}
