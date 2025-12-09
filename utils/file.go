package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

func Bytes2StringByDelimiter(b []byte, delimiter string) string {
	var str string
	for i, v := range b {
		if i == len(b)-1 {
			str += fmt.Sprintf("%d", v)
			break
		} else {
			str += fmt.Sprintf("%d", v) + delimiter
		}
	}
	return str
}

func GetFileBytes(filePath string) []byte {
	fileByte, err := os.ReadFile(filePath)
	if err != nil {
		return []byte{}
	}
	return fileByte
}

func SaveToFile(filePath string, content string) bool {
	file, e := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if e != nil {
		return false
	}
	defer file.Close()

	_, err := file.WriteString(content)
	return err == nil
}

func DirExist(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && info.IsDir()
}

func CreateDir(path string) error {
	if DirExist(path) {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

func FileExist(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetFileLines(filePath string) []string {
	var lines []string
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file [%s] error!\n", filePath)
		return lines
	}

	fs := bufio.NewScanner(f)
	fs.Split(bufio.ScanLines)

	for fs.Scan() {
		lines = append(lines, fs.Text())
	}

	f.Close()
	return lines
}

func GetFileAll(filePath string) string {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("open file [%s] error!\n", filePath)
		return ""
	}
	return string(bytes)
}

func ReadFromReader(r io.Reader) string {
	var buffer bytes.Buffer
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		buffer.Write(buf[:n])
	}
	return buffer.String()
}
