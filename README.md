# httpreaderat

[![Go Reference](https://pkg.go.dev/badge/github.com/guilhem/httpreaderat.svg)](https://pkg.go.dev/github.com/guilhem/httpreaderat)
[![Go](https://github.com/guilhem/httpreaderat/actions/workflows/go.yml/badge.svg)](https://github.com/guilhem/httpreaderat/actions/workflows/go.yml)

Read a remote file through Go's `io.ReaderAt` interface using HTTP byte range
requests. This lets `archive/zip` inspect an archive and extract individual files
without downloading the whole archive when the server supports ranges.

## Install

Requires Go 1.26 or newer.

```sh
go get github.com/guilhem/httpreaderat
```

## List a remote ZIP archive

Save this as `main.go` in a Go module, install the package, then run
`go run . https://your-server.example/archive.zip`.

```go
package main

import (
    "archive/zip"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/guilhem/httpreaderat"
)

func main() {
    if len(os.Args) != 2 {
        log.Fatal("usage: listzip <archive-url>")
    }
    req, err := http.NewRequest(http.MethodGet, os.Args[1], nil)
    if err != nil {
        log.Fatal(err)
    }
    client := &http.Client{Timeout: 30 * time.Second}
    reader, err := httpreaderat.New(client, req, nil)
    if err != nil {
        log.Fatal(err)
    }
    if reader.Size() < 0 {
        log.Fatal("ZIP reading requires a known archive size")
    }
    archive, err := zip.NewReader(reader, reader.Size())
    if err != nil {
        log.Fatal(err)
    }
    for _, file := range archive.File {
        fmt.Println(file.Name)
    }
}
```

`New` immediately sends a one-byte GET range request to discover support and
metadata. Only GET requests are accepted. The request is a prototype: each read
copies its headers, including authentication headers. Keep the prototype
unchanged while using the reader. Its context applies to the requests; a nil
client uses `http.DefaultClient`, which has no default timeout.

## Fallback and buffering

Passing `nil` as the store requires range support; otherwise `New` returns
`ErrNoRange` when a server responds with the whole file. To allow a full download:

```go
store := httpreaderat.NewDefaultStore()
defer store.Close()
reader, err := httpreaderat.New(client, req, store)
```

The default store uses memory for files below its 1 MiB threshold, then a
temporary file below its 1 GiB threshold. Reaching a limit triggers fallback or
`ErrStoreLimit`; the limits are exclusive. Close the store even when `New` fails.
Custom stores can be built from `NewStoreMemory`, `NewStoreFile`, and
`NewLimitedStore`. The store owns the downloaded data; the HTTP reader does not
close it automatically.

Range mode makes one HTTP request per nonempty `ReadAt` call that reaches the
server. For repeated small reads, wrap it with
[chunkreaderat](https://github.com/guilhem/chunkreaderat), passing `reader.Size()`
as the source size. Choose the chunk size and cache capacity for your workload.
For `io.ReadSeeker`, use `io.NewSectionReader(reader, 0, reader.Size())` with a
known size.

## Behavior and limits

- Reads are safe to call concurrently when the supplied client and store support
  concurrent reads. Do not refill or close a store while it is in use.
- Short reads return the number of bytes delivered and an error; end-of-file
  reads return `io.EOF`.
- Range responses must provide matching `Content-Range` and `Content-Length`
  headers. `Size()` is `-1` when the server reports an unknown total size.
- Changes in total size, ETag, or Last-Modified return `ErrValidationFailed`.
  These checks do not guarantee a snapshot when servers omit validators or
  change content without changing them. Buffered fallback data is a local copy.
- Empty remote files are not supported by the initial one-byte range probe.

## Run the extraction example

The nested example module uses the local checkout. It writes one archive entry
to standard output and enables the full-download fallback:

```sh
cd example
go run . https://go.dev/dl/go1.27.1.windows-amd64.zip go/LICENSE
```

## Development

```sh
go test -race ./...
go vet ./...
(cd example && go build ./... && go test -race ./... && go vet ./...)
```

CI builds and checks both the library and the example module.

## Attribution and license

This is Guilhem Lettron's adaptation of the HTTP ReaderAt work credited to
Janne Snabb (2018). Both authors' copyright notices are preserved in
[LICENSE](LICENSE). Licensed under MIT.
