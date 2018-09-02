package rds

import (
	"bytes"
	"encoding/base64"
	"strconv"
)

func i64(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func str(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strs(s []string) (o []*string) {
	for _, ss := range s {
		o = append(o, &ss)
	}
	return o
}

func bo(b bool) *bool {
	return &b
}

func enc(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	buf := bytes.NewBuffer(nil)
	base64.NewEncoder(base64.RawStdEncoding, buf)
	return buf.Bytes()
}

func strI64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func encStr(s string) []byte {
	src := []byte(s)
	return enc(src)
}

func encI64(i int64) []byte {
	src := []byte(strconv.FormatInt(i, 10))
	return enc(src)
}
