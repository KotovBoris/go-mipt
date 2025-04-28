//go:build !solution

package otp

import (
	"io"
)

type cipherReader struct {
	r    io.Reader
	prng io.Reader
}

func (cr *cipherReader) Read(p []byte) (n int, err error) {
	n, err = cr.r.Read(p)
	if err != nil && err != io.EOF {
		return n, err
	}

	prngBytes := make([]byte, n)
	_, _ = cr.prng.Read(prngBytes)

	for i := 0; i < n; i++ {
		p[i] = p[i] ^ prngBytes[i]
	}

	return n, err
}

type cipherWriter struct {
	w    io.Writer
	prng io.Reader
}

func (cw *cipherWriter) Write(p []byte) (n int, err error) {
	prngBytes := make([]byte, len(p))
	_, _ = cw.prng.Read(prngBytes)

	cipherBytes := make([]byte, len(p))
	for i := 0; i < len(p); i++ {
		cipherBytes[i] = p[i] ^ prngBytes[i]
	}

	n, err = cw.w.Write(cipherBytes)
	return n, err
}

func NewReader(r io.Reader, prng io.Reader) io.Reader {
	return &cipherReader{r: r, prng: prng}
}

func NewWriter(w io.Writer, prng io.Reader) io.Writer {
	return &cipherWriter{w: w, prng: prng}
}
