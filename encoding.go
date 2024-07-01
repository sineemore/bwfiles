package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"unicode/utf8"
)

const maxSizeForPlainText = 8000

func encode(b []byte) (string, bool, error) {
	if len(b) <= maxSizeForPlainText && utf8.Valid(b) {
		return string(b), false, nil
	}

	var buf bytes.Buffer
	b64w := base64.NewEncoder(base64.StdEncoding, &buf)
	gw := gzip.NewWriter(b64w)

	_, err := gw.Write(b)
	if err != nil {
		return "", false, err
	}

	if err = gw.Close(); err != nil {
		return "", false, err
	}

	if err = b64w.Close(); err != nil {
		return "", false, err
	}

	return buf.String(), true, nil
}

func decode(content string, encoded bool) ([]byte, error) {
	if !encoded {
		return []byte(content), nil
	}

	buf := bytes.NewBufferString(content)
	b64r := base64.NewDecoder(base64.StdEncoding, buf)
	gr, err := gzip.NewReader(b64r)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	b, err := io.ReadAll(gr)
	if err != nil {
		return nil, err
	}

	if err = gr.Close(); err != nil {
		return nil, err
	}

	return b, nil
}
