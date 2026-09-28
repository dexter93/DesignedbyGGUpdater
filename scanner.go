package main

import (
	"bufio"
	"io"
	"strings"
)

func scanOutput(r io.Reader, fn func(string)) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(splitOutput)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.ReplaceAll(line, "\033[K", "")

		if line == "" {
			continue
		}

		fn(line)
	}

	return scanner.Err()
}

func splitOutput(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i, b := range data {
		if b != '\r' && b != '\n' {
			continue
		}

		advance = i + 1

		if b == '\r' && advance < len(data) && data[advance] == '\n' {
			advance++
		}

		return advance, data[:i], nil
	}

	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}

	return 0, nil, nil
}