package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
}

func WriteFile(filePath, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

func RunShellCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run command %s: %w", command, err)
	}
	return string(output), nil
}

func FindDirectoryRecursive(rootDir, targetDirectoryName string) (string, error) {
	var foundPath string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() == targetDirectoryName {
			foundPath = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if foundPath == "" {
		return "", fmt.Errorf("file %q not found in the directory %q", targetDirectoryName, rootDir)
	}
	return foundPath, nil
}

func ReadAfterFirstBlankLine(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var contentLines []string
	inContent := false

	for scanner.Scan() {
		line := scanner.Text()

		if !inContent {
			if strings.TrimSpace(line) == "" {
				inContent = true
			}
			continue
		}

		contentLines = append(contentLines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(contentLines, "\n"), nil
}

func ReadSkillIntro(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var introLines []string
	state := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		isDelimiter := strings.HasPrefix(trimmed, "---")

		if state == 0 {
			if isDelimiter {
				state = 1
			}
			continue
		}

		if state == 1 {
			if isDelimiter {
				break
			}
			introLines = append(introLines, line)
		}

	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(introLines, "\n"), nil
}

func ProcessSubdirectories(rootDir, targetFileName string) ([]string, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return []string{}, nil
	}

	strings := []string{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		subDirName := entry.Name()
		targetFilePath := filepath.Join(rootDir, subDirName, targetFileName)

		if _, err := os.Stat(targetFilePath); os.IsNotExist(err) {
			fmt.Printf("[Skipping] %s not found in the directory %q\n", targetFileName, subDirName)
			continue
		}

		content, err := ReadSkillIntro(targetFilePath)
		if err != nil {
			return strings, err
		}

		strings = append(strings, content)
	}

	return strings, nil
}
