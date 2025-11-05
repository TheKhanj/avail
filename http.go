package main

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"os"
)

type FileHttp struct {
	Response *http.Response

	fds []io.Closer
}

func OpenHttpResponse() (*FileHttp, error) {
	fhttp := &FileHttp{
		Response: nil,
		fds:      make([]io.Closer, 0),
	}

	filePath, ok := os.LookupEnv("AVAIL_HTTP")
	if !ok {
		return nil, errors.New("AVAIL_HTTP environment variable is not set")
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	fhttp.fds = append(fhttp.fds, file)

	buf := bufio.NewReader(file)
	res, err := http.ReadResponse(buf, nil)
	if err != nil {
		fhttp.Close()
		return nil, err
	}
	fhttp.fds = append(fhttp.fds, res.Body)

	fhttp.Response = res

	return fhttp, nil
}

func (this *FileHttp) Close() error {
	var ret error = nil
	for _, fd := range this.fds {
		err := fd.Close()
		if err != nil {
			ret = err
		}
	}
	return ret
}
