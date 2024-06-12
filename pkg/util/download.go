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

func DownloadBytes(url string, output *bytes.Buffer) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return errors.Join(ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Join(ErrResponseNotOk, errors.New(fmt.Sprintf("bad status code: %s", resp.Status)))
	}

	_, err = io.Copy(output, resp.Body)
	if err != nil {
		return errors.Join(ErrFailedToCopyBytes, err)
	}

	return nil
}
