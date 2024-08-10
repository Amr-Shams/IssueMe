package util

import (
	"os"
	"reflect"
	"strings"
	"testing"

)

func TestFromFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "input.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test content to the file
	testContent := "line1\nline2\nline3"
	if _, err := tempFile.Write([]byte(testContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// get the path of the temp file 
    path := tempFile.Name()

	input := FromFile(path)
	lines := input.LineSlice()

	expectedLines := []string{"line1", "line2", "line3"}
	if !reflect.DeepEqual(lines, expectedLines) {
		t.Errorf("Expected %v, got %v", expectedLines, lines)
	}
}

func TestFromLiteral(t *testing.T) {
	testContent := "line1\nline2\nline3"
	input := FromLiteral(testContent)
	lines := input.LineSlice()

	expectedLines := []string{"line1", "line2", "line3"}
	if !reflect.DeepEqual(lines, expectedLines) {
		t.Errorf("Expected %v, got %v", expectedLines, lines)
	}
}

func TestSections(t *testing.T) {
	testContent := "section1 line1\nsection1 line2\n\nsection2 line1\nsection2 line2\n\nsection3 line1"
	input := FromLiteral(testContent)
	
	sections := []string{}
	for section := range input.Sections() {
		sections = append(sections, strings.TrimSpace(section))
	}

	expectedSections := []string{
		"section1 line1\nsection1 line2",
		"section2 line1\nsection2 line2",
		"section3 line1",
	}

	if !reflect.DeepEqual(sections, expectedSections) {
		t.Errorf("Expected %v, got %v", expectedSections, sections)
	}
}

func TestLineSlice(t *testing.T) {
	testContent := "line1\nline2\nline3"
	input := FromLiteral(testContent)
	
	lines := input.LineSlice()

	expectedLines := []string{"line1", "line2", "line3"}
	if !reflect.DeepEqual(lines, expectedLines) {
		t.Errorf("Expected %v, got %v", expectedLines, lines)
	}
}
