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

/*func TestSubCommandsUsage(t *testing.T) {
	tests := []struct {
		desc string
		cmds []*Command
		want string
	}{
		{"single command", []*Command{{name: "foo"}}, "Commands:\nfoo\n\n"},
		{"single command (description)", []*Command{{name: "foo", description: "bar"}}, "Commands:\nfoo bar\n\n"},
		{"single command (usage)", []*Command{{name: "foo", usageStr: "bar"}}, "Commands:\nfoo bar\n\n"},
		{"single command (usage, description)", []*Command{{name: "foo", usageStr: "bar", description: "foobar"}}, "Commands:\nfoo bar\n    foobar\n\n"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			sc := &subCommands{cmds: make(map[string]*Command)}
			for _, cmd := range test.cmds {
				sc.set(cmd.name, cmd)
			}

			builder := &strings.Builder{}
			sc.usage(&indenter{writer: builder})
			got := builder.String()
			if test.want != got {
				t.Errorf("want string %q got %q", test.want, got)
			}
		})
	}
}*/
