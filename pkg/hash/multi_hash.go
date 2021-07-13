package hash

import (
	"crypto/sha1"
	"crypto/sha256"
	"io"
)

func Sha1Sha256SumReader(f io.Reader) ([]byte, []byte, error) {
	h1 := sha1.New()
	h2 := sha256.New()
	buf := make([]byte, 64*1024) // read 64k once
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}

		h1.Write(buf[:n])
		h2.Write(buf[:n])

		if err == io.EOF {
			break
		}
	}
	return h1.Sum(nil), h2.Sum(nil), nil
}
