// This example outputs a single file from a remote zip file without
// downloading the whole archive. If the server does not support HTTP Range
// Requests, the whole file is downloaded to a backing store as a fallback.
package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/guilhem/httpreaderat"
)

func catZipFile(url, fileName string) (err error) {
	// create http.Request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// a backing store in case the server does not support range requests
	bs := httpreaderat.NewDefaultStore()
	defer bs.Close()

	// make a HTTPReaderAt client
	client := &http.Client{Timeout: 30 * time.Second}
	htrdr, err := httpreaderat.New(client, req, bs)
	if err != nil {
		return err
	}

	// make a ZIP file reader
	zrdr, err := zip.NewReader(htrdr, htrdr.Size())
	if err != nil {
		return err
	}
	// go through the ZIP file contents until desired file is found
	for _, f := range zrdr.File {
		if f.Name == fileName {
			fr, err := f.Open()
			if err != nil {
				return err
			}
			defer fr.Close()
			_, err = io.Copy(os.Stdout, fr)
			return err
		}
	}
	return os.ErrNotExist
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: remotezip <archive-url> <file-name>")
		os.Exit(1)
	}
	err := catZipFile(os.Args[1], os.Args[2])

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
