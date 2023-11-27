package metadata

import (
	"bytes"
	"io"
	"net/http"
)

func RequestBody(req *http.Request) (bs []byte, err error) {
	var (
		buf bytes.Buffer
	)
	if _, err = buf.ReadFrom(req.Body); err != nil {
		return nil, err
	}
	if err = req.Body.Close(); err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
	return buf.Bytes(), nil
}
