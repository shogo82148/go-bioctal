package bioctal

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

type encDecTest struct {
	enc string
	dec []byte
}

var encDecTests = []encDecTest{
	{"", []byte{}},
	{"0001020304050607", []byte{0, 1, 2, 3, 4, 5, 6, 7}},
	{"0c0j0z0w0f0s0b0v", []byte{8, 9, 10, 11, 12, 13, 14, 15}},
	{"v0v1v2v3v4v5v6v7", []byte{0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7}},
	{"vcvjvzvwvfvsvbvv", []byte{0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}},
	{"4b4f", []byte{'N', 'L'}}, // No Lizards
}

func TestEncode(t *testing.T) {
	for i, test := range encDecTests {
		dst := make([]byte, EncodedLen(len(test.dec)))
		n := Encode(dst, test.dec)
		if n != len(dst) {
			t.Errorf("#%d: bad return value: got: %d want: %d", i, n, len(dst))
		}
		if string(dst) != test.enc {
			t.Errorf("#%d: got: %#v want: %#v", i, dst, test.enc)
		}
	}
}

func TestDecode(t *testing.T) {
	// Case for decoding uppercase hex characters, since
	// Encode always uses lowercase.
	decTests := append(encDecTests, encDecTest{"VCVJVZVWVFVSVBVV", []byte{0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}})
	for i, test := range decTests {
		dst := make([]byte, DecodedLen(len(test.enc)))
		n, err := Decode(dst, []byte(test.enc))
		if err != nil {
			t.Errorf("#%d: bad return value: got:%d want:%d", i, n, len(dst))
		} else if !bytes.Equal(dst, test.dec) {
			t.Errorf("#%d: got: %#v want: %#v", i, dst, test.dec)
		}
	}
}

func TestEncodeToString(t *testing.T) {
	for i, test := range encDecTests {
		s := EncodeToString(test.dec)
		if s != test.enc {
			t.Errorf("#%d got:%s want:%s", i, s, test.enc)
		}
	}
}

func TestDecodeString(t *testing.T) {
	for i, test := range encDecTests {
		dst, err := DecodeString(test.enc)
		if err != nil {
			t.Errorf("#%d: unexpected err value: %s", i, err)
			continue
		}
		if !bytes.Equal(dst, test.dec) {
			t.Errorf("#%d: got: %#v want: #%v", i, dst, test.dec)
		}
	}
}

var errTests = []struct {
	in  string
	out string
	err error
}{
	{"", "", nil},
	{"0", "", ErrLength},
	{"azzzz", "", InvalidByteError('a')},
	{"zzzza", "\xaa\xaa", InvalidByteError('a')},
	{"30313", "01", ErrLength},
	{"0g", "", InvalidByteError('g')},
	{"00gg", "\x00", InvalidByteError('g')},
	{"0\x01", "", InvalidByteError('\x01')},
	{"vvbbs", "\xff\xee", ErrLength},
}

func TestDecodeErr(t *testing.T) {
	for _, tt := range errTests {
		out := make([]byte, len(tt.in)+10)
		n, err := Decode(out, []byte(tt.in))
		if string(out[:n]) != tt.out || err != tt.err {
			t.Errorf("Decode(%q) = %q, %v, want %q, %v", tt.in, string(out[:n]), err, tt.out, tt.err)
		}
	}
}

func TestDecodeStringErr(t *testing.T) {
	for _, tt := range errTests {
		out, err := DecodeString(tt.in)
		if string(out) != tt.out || err != tt.err {
			t.Errorf("DecodeString(%q) = %q, %v, want %q, %v", tt.in, out, err, tt.out, tt.err)
		}
	}
}

func TestEncoderDecoder(t *testing.T) {
	for _, multiplier := range []int{1, 128, 192} {
		for _, test := range encDecTests {
			input := bytes.Repeat(test.dec, multiplier)
			output := strings.Repeat(test.enc, multiplier)

			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			r := struct{ io.Reader }{bytes.NewReader(input)} // io.Reader only; not io.WriterTo
			if n, err := io.CopyBuffer(enc, r, make([]byte, 7)); n != int64(len(input)) || err != nil {
				t.Errorf("encoder.Write(%q*%d) = (%d, %v), want (%d, nil)", test.dec, multiplier, n, err, len(input))
				continue
			}

			if encDst := buf.String(); encDst != output {
				t.Errorf("buf(%q*%d) = %v, want %v", test.dec, multiplier, encDst, output)
				continue
			}

			dec := NewDecoder(&buf)
			var decBuf bytes.Buffer
			w := struct{ io.Writer }{&decBuf} // io.Writer only; not io.ReaderFrom
			if _, err := io.CopyBuffer(w, dec, make([]byte, 7)); err != nil || decBuf.Len() != len(input) {
				t.Errorf("decoder.Read(%q*%d) = (%d, %v), want (%d, nil)", test.enc, multiplier, decBuf.Len(), err, len(input))
			}

			if !bytes.Equal(decBuf.Bytes(), input) {
				t.Errorf("decBuf(%q*%d) = %v, want %v", test.dec, multiplier, decBuf.Bytes(), input)
				continue
			}
		}
	}
}

func TestDecoderErr(t *testing.T) {
	for _, tt := range errTests {
		dec := NewDecoder(strings.NewReader(tt.in))
		out, err := io.ReadAll(dec)
		wantErr := tt.err
		// Decoder is reading from stream, so it reports io.ErrUnexpectedEOF instead of ErrLength.
		if wantErr == ErrLength {
			wantErr = io.ErrUnexpectedEOF
		}
		if string(out) != tt.out || err != wantErr {
			t.Errorf("NewDecoder(%q) = %q, %v, want %q, %v", tt.in, out, err, tt.out, wantErr)
		}
	}
}

var sink []byte

func BenchmarkEncode(b *testing.B) {
	for _, size := range []int{256, 1024, 4096, 16384} {
		src := bytes.Repeat([]byte{2, 3, 5, 7, 9, 11, 13, 17}, size/8)
		sink = make([]byte, 2*size)

		b.Run(fmt.Sprintf("%v", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			for i := 0; i < b.N; i++ {
				Encode(sink, src)
			}
		})
	}
}

func BenchmarkDecode(b *testing.B) {
	for _, size := range []int{256, 1024, 4096, 16384} {
		src := bytes.Repeat([]byte{'2', 'w', '7', '4', '4', 'v', 'z', 'z'}, size/8)
		sink = make([]byte, size/2)

		b.Run(fmt.Sprintf("%v", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			for i := 0; i < b.N; i++ {
				Decode(sink, src)
			}
		})
	}
}
