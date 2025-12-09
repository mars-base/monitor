package utils

import (
	"fmt"
	"io"
	"net"
	"os"
)

func ZeroIOFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src file error: %v", err)
	}
	defer srcFile.Close()
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create dest file error: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("copy file error: %v", err)
	}
	return nil
}

func ZeroIOConn(file string, conn net.Conn) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("open file error: %v", err)
	}
	defer f.Close()

	_, err = io.Copy(conn, f)
	if err != nil {
		return fmt.Errorf("copy file to conn error: %v", err)
	}
	return nil
}
