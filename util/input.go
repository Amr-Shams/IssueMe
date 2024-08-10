package util

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
)

type Input struct {
	scanner *bufio.Scanner
	lines   chan string
}

func FromFile(path string) *Input {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open file %s: %s", path, err)
	}

	return newInputFromReader(f, f)
}

func FromLiteral(input string) *Input {
	return newInputFromReader(strings.NewReader(input), nil)
}

func newInputFromReader(r io.Reader, c io.Closer) *Input {
	result := &Input{
		scanner: bufio.NewScanner(r),
		lines:   make(chan string),
	}

	go func() {
		defer func() {
			if c != nil {
				_ = c.Close()
			}
		}()

		for result.scanner.Scan() {
			result.lines <- result.scanner.Text()
		}

		close(result.lines)
	}()

	return result
}

func (c *Input) Lines() <-chan string {
	return c.lines
}

func (c *Input) LineSlice() (result []string) {
	for line := range c.Lines() {
		result = append(result, line)
	}
	return
}

func (c *Input) Sections() <-chan string {
	result := make(chan string)
	go func() {
		var section string
		for line := range c.Lines() {
			if line == "" {
				if len(section) > 0 {
					result <- section
					section = ""
				}
			} else {
				line += "\n"
				section += line
			}
		}
		if len(section) > 0 {
			result <- section
		}
		close(result)
	}()
	return result
}
