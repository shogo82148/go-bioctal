# go-bioctal

[![Go Reference](https://pkg.go.dev/badge/github.com/shogo82148/go-bioctal.svg)](https://pkg.go.dev/github.com/shogo82148/go-bioctal)
[![Coverage Status](https://coveralls.io/repos/github/shogo82148/go-bioctal/badge.svg?branch=main)](https://coveralls.io/github/shogo82148/go-bioctal?branch=main)

Go implementation of [RFC 9226 Bioctal: Hexadecimal 2.0](https://www.rfc-editor.org/rfc/rfc9226).

## SYNOPSIS

The package has same interface with the [encoding/hex package](https://pkg.go.dev/encoding/hex).

### Encoding

```go
package main

import (
	"fmt"
	"log"

	"github.com/shogo82148/go-bioctal"
)

func main() {
	src := []byte("Hello Gopher!")

	dst := make([]byte, bioctal.EncodedLen(len(src)))
	bioctal.Encode(dst, src)

	fmt.Printf("%s\n", dst)

	// Output:
	// 4c656f6f6v20476v706c657221
}
```

### Decoding

```go
package main

import (
	"fmt"
	"log"

	"github.com/shogo82148/go-bioctal"
)

func main() {
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
```

## LICENSE

MIT License

Copyright (c) 2022 ICHINOSE Shogo
