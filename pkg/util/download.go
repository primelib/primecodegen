package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrRequestFailed     = errors.New("request failed")
	ErrResponseNotOk     = errors.New("response not ok")
	ErrFailedToCopyBytes = errors.New("failed to copy bytes")
)

func DownloadBytes(url string) ([]byte, error) {
	buffer := new(bytes.Buffer)
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Join(ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Join(ErrResponseNotOk, errors.New(fmt.Sprintf("bad status code: %s", resp.Status)))
	}

	_, err = io.Copy(buffer, resp.Body)
	if err != nil {
		return nil, errors.Join(ErrFailedToCopyBytes, err)
	}

	return buffer.Bytes(), nil
}
