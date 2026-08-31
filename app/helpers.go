package main

import (
	"fmt"
	"os"
)

func ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
}

type FileArgs struct {
	FilePath string `json:"file_path"`
}
