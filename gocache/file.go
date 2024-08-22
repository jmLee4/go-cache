package gocache

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
)

const (
	baseFilePath      = "./gocache/assets/"
	defaultFileSuffix = ".jpg"
)

type StaticFile struct {
	filename  string
	hashValue string
	content   []byte
}

func NewStaticFile(filename string) *StaticFile {
	file, err := os.Open(baseFilePath + filename + defaultFileSuffix)
	if err != nil {
		log.Fatalf("[File] Failed to open file %s: %v\n", filename, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("[File] Failed to read file: %v\n", err)
	}

	hashValue := md5.Sum(content)
	return &StaticFile{
		filename:  filename,
		hashValue: hex.EncodeToString(hashValue[:]),
		content:   content,
	}
}

func (f *StaticFile) GetContent() []byte {
	return f.content
}
