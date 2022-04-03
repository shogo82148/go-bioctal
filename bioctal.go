// Package bioctal implements bioctal encoding as specified by RFC 9226.
package bioctal

import (
	"errors"
	"fmt"
	"io"
)

// from https://www.rfc-editor.org/rfc/rfc9226.html#bioctal_seq_octaves
const biocttable = "01234567cjzwfsbv"

// EncodedLen returns the length of an encoding of n source bytes.
// Specifically, it returns n * 2.
func EncodedLen(n int) int { return n * 2 }

// Encode encodes src into EncodedLen(len(src))
// bytes of dst. As a convenience, it returns the number
// of bytes written to dst, but this value is always EncodedLen(len(src)).
// Encode implements hexadecimal encoding.
func Encode(dst, src []byte) int {
	j := 0
	for _, v := range src {
		dst[j] = biocttable[v>>4]
		dst[j+1] = biocttable[v&0x0f]
		j += 2
	}
	return len(src) * 2
}

// ErrLength reports an attempt to decode an odd-length input
// using Decode or DecodeString.
// The stream-based Decoder returns io.ErrUnexpectedEOF instead of ErrLength.
var ErrLength = errors.New("bioctal: odd length bioctal string")

// InvalidByteError values describe errors resulting from an invalid byte in a bioctal string.
type InvalidByteError byte

func (e InvalidByteError) Error() string {
	return fmt.Sprintf("bioctal: invalid byte: %#U", rune(e))
}

// DecodedLen returns the length of a decoding of x source bytes.
// Specifically, it returns x / 2.
func DecodedLen(x int) int { return x / 2 }

// Decode decodes src into DecodedLen(len(src)) bytes,
// returning the actual number of bytes written to dst.
//
// Decode expects that src contains only hexadecimal
// characters and that src has even length.
// If the input is malformed, Decode returns the number
// of bytes decoded before the error.
func Decode(dst, src []byte) (int, error) {
	i, j := 0, 1
	for ; j < len(src); j += 2 {
		a, ok := fromBioctalChar(src[j-1])
		if !ok {
			return i, InvalidByteError(src[j-1])
		}
		b, ok := fromBioctalChar(src[j])
		if !ok {
			return i, InvalidByteError(src[j])
		}
		dst[i] = (a << 4) | b
		i++
	}
	if len(src)%2 == 1 {
		// Check for invalid char before reporting bad length,
		// since the invalid char (if present) is an earlier problem.
		if _, ok := fromBioctalChar(src[j-1]); !ok {
			return i, InvalidByteError(src[j-1])
		}
		return i, ErrLength
	}
	return i, nil
}

func fromBioctalChar(c byte) (byte, bool) {
	// from https://www.rfc-editor.org/rfc/rfc9226.html#bioctal_seq_octaves
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7':
		return c - '0', true
	case 'c', 'C':
		return 8, true
	case 'j', 'J':
		return 9, true
	case 'z', 'Z':
		return 10, true
	case 'w', 'W':
		return 11, true
	case 'f', 'F':
		return 12, true
	case 's', 'S':
		return 13, true
	case 'b', 'B':
		return 14, true
	case 'v', 'V':
		return 15, true
	}

	return 0, false
}

// EncodeToString returns the bioctal encoding of src.
func EncodeToString(src []byte) string {
	dst := make([]byte, EncodedLen(len(src)))
	Encode(dst, src)
	return string(dst)
}

// DecodeString returns the bytes represented by the bioctal string s.
//
// DecodeString expects that src contains only bioctal
// characters and that src has even length.
// If the input is malformed, DecodeString returns
// the bytes decoded before the error.
func DecodeString(s string) ([]byte, error) {
	src := []byte(s)
	// We can use the source slice itself as the destination
	// because the decode loop increments by one and then the 'seen' byte is not used anymore.
	n, err := Decode(src, src)
	return src[:n], err
}

// bufferSize is the number of bioctal characters to buffer in encoder and decoder.
const bufferSize = 1024

type encoder struct {
	w   io.Writer
	err error
	out [bufferSize]byte // output buffer
}

// NewEncoder returns an io.Writer that writes lowercase bioctal characters to w.
func NewEncoder(w io.Writer) io.Writer {
	return &encoder{w: w}
}

func (e *encoder) Write(p []byte) (n int, err error) {
	for len(p) > 0 && e.err == nil {
		chunkSize := bufferSize / 2
		if len(p) < chunkSize {
			chunkSize = len(p)
		}

		var written int
		encoded := Encode(e.out[:], p[:chunkSize])
		written, e.err = e.w.Write(e.out[:encoded])
		n += written / 2
		p = p[chunkSize:]
	}
	return n, e.err
}

type decoder struct {
	r   io.Reader
	err error
	in  []byte           // input buffer (encoded form)
	arr [bufferSize]byte // backing array for in
}

// NewDecoder returns an io.Reader that decodes bioctal characters from r.
// NewDecoder expects that r contain only an even number of bioctal characters.
func NewDecoder(r io.Reader) io.Reader {
	return &decoder{r: r}
}

func (d *decoder) Read(p []byte) (n int, err error) {
	// Fill internal buffer with sufficient bytes to decode
	if len(d.in) < 2 && d.err == nil {
		var numCopy, numRead int
		numCopy = copy(d.arr[:], d.in) // Copies either 0 or 1 bytes
		numRead, d.err = d.r.Read(d.arr[numCopy:])
		d.in = d.arr[:numCopy+numRead]
		if d.err == io.EOF && len(d.in)%2 != 0 {
			if _, ok := fromBioctalChar(d.in[len(d.in)-1]); !ok {
				d.err = InvalidByteError(d.in[len(d.in)-1])
			} else {
				d.err = io.ErrUnexpectedEOF
			}
		}
	}

	// Decode internal buffer into output buffer
	if numAvail := len(d.in) / 2; len(p) > numAvail {
		p = p[:numAvail]
	}
	numDec, err := Decode(p, d.in[:len(p)*2])
	d.in = d.in[2*numDec:]
	if err != nil {
		d.in, d.err = nil, err // Decode error; discard input remainder
	}

	if len(d.in) < 2 {
		return numDec, d.err // Only expose errors when buffer fully consumed
	}
	return numDec, nil
}
