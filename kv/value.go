package kv

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"time"
)

type Value struct {
	lastModified time.Time
	md5          string
	data         string
}

func compressAndEncode(data []byte) (string, error) {
	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, flate.BestCompression)
	if err != nil {
		return "", NewCompressionError(data, err)
	}

	if _, err = gz.Write(data); err != nil {
		return "", NewCompressionError(data, err)
	}

	if err = gz.Flush(); err != nil {
		return "", NewCompressionError(data, err)
	}

	if err = gz.Close(); err != nil {
		return "", NewCompressionError(data, err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func NewValue(data []byte) (*Value, error) {
	b64data, err := compressAndEncode(data)
	if err != nil {
		return nil, err
	}

	return &Value{
		lastModified: time.Now(),
		md5:          fmt.Sprintf("%x", md5.Sum(data)),
		data:         b64data,
	}, nil
}

func (v *Value) LastModified() time.Time {
	return v.lastModified
}

func (v *Value) MD5() string {
	return v.md5
}

func (v *Value) Data() string {
	return v.data
}

func (v *Value) Text() string {
	gzipped, _ := base64.StdEncoding.DecodeString(v.data)
	r, _ := gzip.NewReader(bytes.NewReader(gzipped))
	raw, _ := ioutil.ReadAll(r)
	return string(raw)
}
