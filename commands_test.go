package cli

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	cb := func(string, ...string) ([]string, error) { return nil, nil }

	tests := []struct {
		desc   string
		option Option
		want   *Command
	}{
		{"UsageOption", UsageOption("useless usage"), &Command{Usage: "useless usage", output: os.Stderr}},
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

func TestArgCallbackOption(t *testing.T) {
	tests := []struct {
		desc    string
		cb      CommandFunc
		input   []string
		wantErr string
	}{
		{"bool", Callback(func(b bool) error { return fmt.Errorf("%v", b) }), []string{"true"}, "true"},
		{"bool (parse error)", Callback(func(b bool) error { return fmt.Errorf("%v", b) }), []string{"yo"}, "parse error"},
		{"duration", Callback(func(d time.Duration) error { return fmt.Errorf("%v", d) }), []string{"1s"}, "1s"},
		{"float64", Callback(func(f float64) error { return fmt.Errorf("%v", f) }), []string{"1.234"}, "1.234"},
		{"int", Callback(func(i int) error { return fmt.Errorf("%v", i) }), []string{"4234"}, "4234"},
		{"int64", Callback(func(i int64) error { return fmt.Errorf("%v", i) }), []string{"934"}, "934"},
		{"string", Callback(func(s string) error { return fmt.Errorf("%v", s) }), []string{"foo"}, "foo"},
		{"uint", Callback(func(u uint) error { return fmt.Errorf("%v", u) }), []string{"1234"}, "1234"},
		{"uint64", Callback(func(u uint64) error { return fmt.Errorf("%v", u) }), []string{"1234"}, "1234"},
		{"value", Callback(func(b *boolValue) error { return fmt.Errorf("%v", b.String()) }), []string{"true"}, "true"},
		{"bad bool", Callback(func(b boolValue) error { return fmt.Errorf("%v", b.String()) }), []string{"true"}, "Type cli.boolValue does not implement Value interface"},
		{"byte slice", Callback(func(b byteslice) error { return fmt.Errorf("%v", b.String()) }), []string{"true"}, "cli.byteslice argument must be a pointer to a type implementing Value"},
		{"no func", Callback("hello world"), []string{"true"}, "Provided callback is not a function"},
		{"int slice", Callback(func(i *intSlice) error { return fmt.Errorf("%v", i.String()) }), []string{"1", "2", "3", "4", "5"}, "1,2,3,4,5"},
		{"two values", Callback(func(a, b int) error { return fmt.Errorf("%d %d", a, b) }), []string{"1", "2"}, "1 2"},
		{"two expected one received", Callback(func(a, b int) error { return fmt.Errorf("%d %d", a, b) }), []string{"1"}, "not enough arguments given"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cmd := New("", ErrorHandlingOption(ContinueOnError))
			cmd.Callback = test.cb
			_, err := cmd.Run(test.input)
			if err == nil {
				t.Errorf("Expected error")
			} else if err.Error() != test.wantErr {
				t.Errorf("Wanted error %s got %s", test.wantErr, err.Error())
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
		{"usage str", func(cmd *Command) { cmd.Usage = "foobar" }, "Usage: usage str foobar\n"},
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
