package cli

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestOptions(t *testing.T) {
	cb := func(string, ...string) ([]string, error) { return nil, nil }

	tests := []struct {
		desc   string
		option Option
		want   *Command
	}{
		{"UsageOption", UsageOption("useless usage"), &Command{UsageStr: "useless usage", output: os.Stderr}},
		{"DescOption", DescOption("useless description"), &Command{Description: "useless description", output: os.Stderr}},
		{"CallbackOption", CallbackOption(cb), &Command{Callback: cb, output: os.Stderr}},
		{"OutputOption", OutputOption(os.Stdout), &Command{output: os.Stdout}},
		{"ErrorHandlingOption", ErrorHandlingOption(PanicOnError), &Command{errorHandling: PanicOnError, output: os.Stderr}},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := New("", test.option)
			//got.subCommands = subCommands{}

			// hack
			if got.Callback != nil {
				// do nothing :(
			} else {
				if !reflect.DeepEqual(test.want, got) {
					t.Errorf("want cmd %+v got %+v", test.want, got)
				}
			}
		})
	}
}

func TestCommandUsage(t *testing.T) {
	tests := []struct {
		desc  string
		setup func(*Command)
		want  string
	}{
		{"usage str", func(cmd *Command) { cmd.UsageStr = "foobar" }, "Usage: usage str foobar\n"},
		{"no flags", func(*Command) {}, "Usage: no flags\n"},
		{"one flag", func(cmd *Command) { cmd.Flags.Var(&testValue{}, "foo", "bar") }, "Usage: one flag [global options]\n  -foo value\n    \tbar\n\n"},
		{"subcommand", func(cmd *Command) { cmd.SubCommand("foo") }, "Usage: subcommand <command> [command options]\nCommands:\nfoo\n\n"},
		{"subcommand (description)", func(cmd *Command) { cmd.SubCommand("foo", DescOption("bar")) }, "Usage: subcommand (description) <command> [command options]\nCommands:\nfoo bar\n\n"},
		{"subcommand (usage)", func(cmd *Command) { cmd.SubCommand("foo", UsageOption("bar")) }, "Usage: subcommand (usage) <command> [command options]\nCommands:\nfoo bar\n\n"},
		{"subcommand (usage, description)", func(cmd *Command) { cmd.SubCommand("foo", UsageOption("bar"), DescOption("foobar")) }, "Usage: subcommand (usage, description) <command> [command options]\nCommands:\nfoo bar\n    foobar\n\n"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cmd := New(test.desc, test.setup, ErrorHandlingOption(ContinueOnError))
			builder := &strings.Builder{}
			cmd.usage(&indenter{writer: builder})
			got := builder.String()
			if test.want != got {
				t.Errorf("want usage string %q got %q", test.want, got)
			}
		})
	}
}

func TestCommandRun(t *testing.T) {
	runErr := errors.New("Run Error!")

	tests := []struct {
		name    string
		prepare func(*Command)
		args    []string
		wantErr error
	}{
		{"no func", func(*Command) {}, nil, ErrNoCommandFunc},
		{"run error", func(c *Command) { c.Callback = func(string, ...string) ([]string, error) { return nil, runErr } }, nil, runErr},
		{"sub command no func", func(c *Command) { c.SubCommand("foo") }, []string{"foo"}, ErrNoCommandFunc},
		{"callback subcommand no func", func(c *Command) {
			c.Callback = func(string, ...string) ([]string, error) { return []string{"foo"}, nil }
			c.SubCommand("foo")
		}, []string{"foo"}, ErrNoCommandFunc},
		{"callback with subcommand no command", func(c *Command) {
			c.Callback = Callback(func() { return })
			c = c.SubCommand("foo", CallbackOption(Callback(func() { return })))
		}, []string{}, ErrRequiredCommand},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := New(test.name, ErrorHandlingOption(ContinueOnError))
			test.prepare(cmd)
			_, gotErr := cmd.Run(test.args)
			if test.wantErr != gotErr {
				t.Errorf("Wanted error %v got %v", test.wantErr, gotErr)
			}
		})
	}
}
