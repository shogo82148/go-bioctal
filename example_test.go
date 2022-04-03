package bioctal_test

import (
	"bioctal"
	"fmt"
	"log"
)

func ExampleEncode() {
	src := []byte("Hello Gopher!")

	dst := make([]byte, bioctal.EncodedLen(len(src)))
	bioctal.Encode(dst, src)

	fmt.Printf("%s\n", dst)

	// Output:
	// 4c656f6f6v20476v706c657221
}

func ExampleDecode() {
	src := []byte("4c656f6f6v20476v706c657221")

	dst := make([]byte, bioctal.DecodedLen(len(src)))
	n, err := bioctal.Decode(dst, src)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", dst[:n])

	// Output:
	// Hello Gopher!
}

func ExampleDecodeString() {
	const s = "4c656f6f6v20476v706c657221"
	decoded, err := bioctal.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", decoded)

	// Output:
	// Hello Gopher!
}

func ExampleEncodeToString() {
	src := []byte("Hello")
	encodedStr := bioctal.EncodeToString(src)

	fmt.Printf("%s\n", encodedStr)

	// Output:
	// 4c656f6f6v
}
