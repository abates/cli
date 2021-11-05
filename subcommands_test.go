package cli

import (
	"testing"
)

func TestSubCommandsGetSet(t *testing.T) {
	sc := []*Command{}

	if subCommands(sc).Len() != 0 {
		t.Errorf("Expected len 0 got %d", subCommands(sc).Len())
	}

	got := subCommands(sc).get("foo")
	if got != nil {
		t.Errorf("Expected nil got %v", got)
	}

	sc = append(sc, &Command{Name: "foo"})
	if subCommands(sc).Len() != 1 {
		t.Errorf("Expected len 1 got %d", subCommands(sc).Len())
	}

	got = subCommands(sc).get("foo")
	if got == nil {
		t.Errorf("Expected non-nil")
	}

	if subCommands(sc).maxLen() != len("foo") {
		t.Errorf("Expected %d got %d", len("foo"), subCommands(sc).maxLen())
	}
}

func TestSubCommandsGet(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		lookup string
		found  bool
	}{
		{"single item", []string{"one"}, "one", true},
		{"two items", []string{"one", "two"}, "one", true},
		{"three items", []string{"one", "two", "three"}, "two", true},
		{"three items again", []string{"one", "two", "three"}, "three", true},
		{"three items finally", []string{"two", "one", "three"}, "one", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands := []*Command{}
			for _, input := range test.input {
				commands = append(commands, &Command{Name: input})
			}

			c := subCommands(commands).get(test.lookup)
			if (c == nil) == test.found {
				t.Errorf("Wanted found to be %v got %v", test.found, c == nil)
			}

			if c != nil && c.Name != test.lookup {
				t.Errorf("Expecting name to be %q got %q", test.lookup, c.Name)
			}
		})
	}
}
