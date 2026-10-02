package main

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCatZipFile(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("../fixtures")))
	defer server.Close()
	output, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	stdout := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = stdout }()
	if err := catZipFile(server.URL+"/file.zip", "test2"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../fixtures/test2")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("extracted %q; want %q", got, want)
	}
	if err := catZipFile(server.URL+"/file.zip", "missing"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing entry: got %v; want os.ErrNotExist", err)
	}
	if err := catZipFile(":", "test2"); err == nil {
		t.Fatal("invalid URL accepted")
	}
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	if err := catZipFile(server.URL+"/file.zip", "test2"); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("closed output: got %v; want os.ErrClosed", err)
	}
}
