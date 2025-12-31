package utils

import (
	"io"
	"strings"
	"testing"
)

func TestCleanReader(t *testing.T) {
	// Test cases covering:
	// 1. Logic: \r should reset cursor, overwrite content.
	// 2. Logic: \b should move back, overwrite content.
	// 3. Logic: \x1b[K (Clear) should truncate.
	// 4. Logic: \x1b[D (Left) should move back.
	// 5. Logic: Colors should be preserved attached to cells.

	input := "Normal Text\x1b[31mRed Text\x1b[0m\rOverwritten\n" + // "Overwritten" (shorter) overwrites start.
		"Full Line\rPartial\n" + // "Partial" overwrites "Full ". Result "Partialine".
		"Typo\x08\x08Correction\n" + // "TyCorrection"
		"Command\x1b[5Dmn\x1b[K\n" + // "Command" -> Left 5 ("mmand") -> "mn" -> Clear. Result: "Commn".
		"Color\x1b[32mGap\x1b[2D\x1b[31mO\n" // "Color" + Green "Gap" -> Left 2 ("ap") -> Red "O" -> "Color" + Green "G" + Red "O" + Green "p".

	expected := "Overwritten\x1b[31mRed Text\x1b[0m\n" + // "Overwritten" (11) overwrites "Normal Text" (11).
		"Partialne\n" + // "Partial" + "ine"
		"TyCorrection\n" +
		"Comn\n" +
		"Color\x1b[32mG\x1b[31mO\x1b[32mp\x1b[0m\n"

	// explanation for Color case:
	// "Color" (def style)
	// \x1b[32m (Green)
	// 'G' (Green)
	// 'a' (Green)
	// 'p' (Green)
	// Cursor at end.
	// \x1b[2D -> Cursor back 2 (at 'a')
	// \x1b[31m (Red)
	// 'O' (Red) -> Overwrites 'a'.
	// Result: 'G' (Green), 'O' (Red), 'p' (Green).
	// Render: "Color" \x1b[32m "G" \x1b[31m "O" \x1b[32m "p" \x1b[0m.

	r := NewCleanReader(strings.NewReader(input))
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Compare line by line to easier debugging
	gotLines := strings.Split(string(out), "\n")
	expLines := strings.Split(expected, "\n")

	for i := 0; i < len(expLines); i++ {
		if i >= len(gotLines) {
			t.Errorf("Line %d missing. Expected %q", i, expLines[i])
			continue
		}
		if gotLines[i] != expLines[i] {
			t.Errorf("Line %d mismatch:\nExp: %q\nGot: %q", i, expLines[i], gotLines[i])
		}
	}
}
