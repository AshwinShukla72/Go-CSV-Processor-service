package unit

import (
	"testing"

	. "github.com/AshwinShukla72/csv-processor/services"
)

func TestCSVParser_Parse(t *testing.T) {
	parser := NewCSVParser()
	input := []byte("name,email\nJohn Doe,john@example.com\nJane Doe,none\n")
	expected := "name,email,has_email\nJohn Doe,john@example.com,true\nJane Doe,none,false\n"

	out, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if string(out) != expected {
		t.Errorf("unexpected result:\n%s\nexpected:\n%s", out, expected)
	}
}

func TestCSVParser_Parse_Empty(t *testing.T) {
	parser := NewCSVParser()
	input := []byte("")
	// Empty input should produce empty output (or at least not panic / return invalid bytes)
	out, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse failed on empty input: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty output for empty input, got: %q", out)
	}
}

func TestCSVParser_Parse_MultipleEmails(t *testing.T) {
	parser := NewCSVParser()
	input := []byte("name,email,alt\nAlice,alice@example.com,alt@example.org\nBob,none,none\n")
	expected := "name,email,alt,has_email\nAlice,alice@example.com,alt@example.org,true\nBob,none,none,false\n"

	out, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if string(out) != expected {
		t.Errorf("unexpected result:\n%s\nexpected:\n%s", out, expected)
	}
}

func TestCSVParser_Parse_CommaInField(t *testing.T) {
	parser := NewCSVParser()
	// CSV contains a quoted field with a comma — ensure parser handles quoted fields correctly.
	input := []byte("name,email\n\"Smith, John\",john@example.com\n")
	expected := "name,email,has_email\n\"Smith, John\",john@example.com,true\n"

	out, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if string(out) != expected {
		t.Errorf("unexpected result with quoted field:\n%s\nexpected:\n%s", out, expected)
	}
}
