package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type ErrorHandling int

const (
	ExitOnError     ErrorHandling = iota // Print usage and Call os.Exit(2).
	ContinueOnError                      // Return a descriptive error.
	PanicOnError                         // Call panic with a descriptive error.
)

type indenter struct {
	writer io.Writer
	count  int
}

func (ind *indenter) Print(a ...interface{}) {
	fmt.Fprint(ind.writer, a...)
}

func (ind *indenter) Printf(format string, a ...interface{}) {
	fmt.Fprintf(ind.writer, format, a...)
}

func (ind *indenter) Println(a ...interface{}) {
	fmt.Fprintln(ind.writer, a...)
}

func (ind *indenter) Indentln(a ...interface{}) {
	fmt.Fprint(ind.writer, strings.Repeat("  ", ind.count))
	fmt.Fprintln(ind.writer, a...)
}

func (ind *indenter) Indentf(format string, a ...interface{}) {
	fmt.Fprint(ind.writer, strings.Repeat("  ", ind.count))
	fmt.Fprintf(ind.writer, format, a...)
}

// CommandFunc is the callback function that will be executed when a
// command is called. If the CommandFunc returns a non-nil error
// then processing stops immediately
type CommandFunc func(name string, args ...string) ([]string, error)

// Command represents a single cli command. The idea is that a cli app
// is run such as:
//    program cmd <flags>
// and can have nested commands:
//    program cmd1 <flags> subcmd1 <flags> ...
// a Command object represents a single command in the hierarchy and is
// a placeholder to register subcommands
type Command struct {
	Name        string
	Description string
	UsageStr    string
	Callback    CommandFunc
	SubCommands []*Command
	Flags       flag.FlagSet

	errorHandling ErrorHandling
	output        io.Writer
}

type Option func(*Command)

func UsageOption(usageStr string) Option { return func(cmd *Command) { cmd.UsageStr = usageStr } }

func DescOption(description string) Option {
	return func(cmd *Command) { cmd.Description = description }
}

func CallbackOption(callback CommandFunc) Option {
	return func(cmd *Command) { cmd.Callback = callback }
}

func OutputOption(output io.Writer) Option { return func(cmd *Command) { cmd.SetOutput(output) } }

func ErrorHandlingOption(errorHandling ErrorHandling) Option {
	return func(cmd *Command) { cmd.errorHandling = errorHandling }
}

// New will return a Command object that is initialized according
// to the supplied command options
func New(name string, options ...Option) *Command {
	cmd := &Command{
		Name:          name,
		output:        os.Stderr,
		errorHandling: ExitOnError,
	}

	for _, option := range options {
		option(cmd)
	}

	return cmd
}

// SubCommand adds a subcommand to the current command hierarchy
func (cmd *Command) SubCommand(name string, options ...Option) *Command {
	subCommand := New(name)
	subCommand.Flags.SetOutput(cmd.output)
	subCommand.errorHandling = cmd.errorHandling
	for _, option := range options {
		option(subCommand)
	}
	cmd.SubCommands = append(cmd.SubCommands, subCommand)
	return subCommand
}

// SetOutput will set the io.Writer used for printing usage
func (cmd *Command) SetOutput(writer io.Writer) {
	cmd.output = writer
}

func (cmd *Command) Usage() {
	ind := &indenter{writer: cmd.output}
	if ind.writer == nil {
		ind.writer = os.Stderr
	}
	cmd.usage(ind)
}

func (cmd *Command) usage(ind *indenter) {
	// count the number of flags that have been created
	numFlags := 0
	cmd.Flags.VisitAll(func(*flag.Flag) { numFlags++ })

	if ind.count == 0 {
		ind.Indentf("Usage: %s", cmd.Name)
		if cmd.UsageStr != "" {
			ind.Printf(" %s\n", cmd.UsageStr)
		} else {
			if numFlags > 0 {
				ind.Print(" [global options]")
			}

			if len(cmd.SubCommands) > 0 {
				ind.Printf(" <command> [command options]\n")
			} else {
				ind.Println()
			}
		}
	}
	builder := &strings.Builder{}
	cmd.Flags.SetOutput(builder)
	cmd.Flags.PrintDefaults()

	str := builder.String()
	if len(str) > 0 {
		for _, line := range strings.Split(str, "\n") {
			ind.Indentln(line)
		}
	}

	if len(cmd.SubCommands) > 0 {
		ind.Indentln("Commands:")
		nameFmt := fmt.Sprintf("%%-%ds", subCommands(cmd.SubCommands).maxLen())
		var prevCmd *Command
		subCommands(cmd.SubCommands).sort()
		for _, command := range cmd.SubCommands {
			if prevCmd != nil && len(prevCmd.SubCommands) == 0 && len(command.SubCommands) > 0 {
				ind.Println()
			}

			ind.Indentf(nameFmt, command.Name)
			if command.UsageStr == "" {
				if command.Description != "" {
					ind.Printf(" %s\n", command.Description)
				} else {
					ind.Println()
				}
			} else {
				ind.Printf(" %s\n", command.UsageStr)
				if command.Description != "" {
					ind.Indentf("%s %s\n", strings.Repeat(" ", subCommands(cmd.SubCommands).maxLen()), command.Description)
				}
			}

			command.usage(&indenter{writer: ind.writer, count: ind.count + subCommands(cmd.SubCommands).maxLen()})
			prevCmd = command
		}
		ind.Println()
	}
}

func (cmd *Command) handleErr(err error) error {
	if err != nil {
		if cmd.errorHandling == ExitOnError {
			ind := &indenter{writer: cmd.output}
			if cmd.output == nil {
				ind.writer = os.Stderr
			}
			ind.Printf("%v\n", err)
			if errors.Is(err, ErrUsage) {
				cmd.usage(ind)
			}
			os.Exit(2)
		} else if cmd.errorHandling == PanicOnError {
			panic(err)
		}
	}
	return err
}

// Run the command.
func (cmd *Command) runCallback(args []string) ([]string, error) {
	if cmd.Callback == nil {
		return args, ErrNoCommandFunc
	}
	return cmd.Callback(cmd.Name, args...)
}

func (cmd *Command) Lookup(name string) (subcmd *Command, found bool) {
	subcmd = subCommands(cmd.SubCommands).get(name)
	if subcmd != nil {
		found = true
	}
	return
}

func (cmd *Command) runSubcommand(args []string) ([]string, error) {
	var err error
	if len(cmd.SubCommands) > 0 {
		if len(args) < 1 {
			err = ErrRequiredCommand
		} else {
			subCmdName := args[0]
			subCmdArgs := args[1:]
			subCmd, found := cmd.Lookup(subCmdName)
			if !found {
				err = fmt.Errorf("%w %q", ErrUnknownCommand, subCmdName)
			} else {
				args, err = subCmd.Run(subCmdArgs)
			}
		}
	} else {
		err = fmt.Errorf("%w for %s", ErrNoCommandFunc, args[0])
	}
	return args, err
}

func (cmd *Command) Run(args []string) ([]string, error) {
	err := cmd.Flags.Parse(args)
	if err == nil {
		args = cmd.Flags.Args()
		args, err = cmd.runCallback(args)

		if len(cmd.SubCommands) > 0 && (err == nil || errors.Is(err, ErrNoCommandFunc)) {
			args, err = cmd.runSubcommand(args)
		}
	}

	return args, cmd.handleErr(err)
}
