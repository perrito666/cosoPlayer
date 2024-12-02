package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

type noopCloser struct {
	io.Reader
}

func (*noopCloser) Close() error { return nil }

type Skin struct {
	files map[string][]byte
}

func (s *Skin) Open(name string) (io.ReadCloser, error) {
	f, ok := s.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &noopCloser{bytes.NewReader(f)}, nil
}

// we can afford to load this into memory, they are tiny
func skinFromPath(skinPath string) (*Skin, error) {
	f, err := os.Open(skinPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open skin file: %w", err)
	}
	defer f.Close()
	fInfo, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat skin file: %w", err)
	}
	zf, err := zip.NewReader(f, fInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to open skin file: %w", err)
	}
	s := &Skin{
		files: make(map[string][]byte),
	}
	for _, file := range zf.File {
		r, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open skin file: %w", err)
		}
		s.files[strings.ToLower(file.Name)], err = io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("failed to read skin file: %w", err)
		}
		r.Close() // this is likely not that important as f closes upon exit.
	}
	return s, nil
}
