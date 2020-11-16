package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

var (
	ErrUnknownCommand  = errors.New("Unknown command")
	ErrRequiredCommand = errors.New("A command is required")
	ErrNoCommandFunc   = errors.New("No callback function was provided for the command")
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
type CommandFunc func(name string) error

// Command represents a single cli command. The idea is that a cli app
// is run such as:
//    program cmd <flags>
// and can have nested commands:
//    program cmd1 <flags> subcmd1 <flags> ...
// a Command object represents a single command in the hierarchy and is
// a placeholder to register subcommands
type Command struct {
	name          string
	usageStr      string
	description   string
	callback      CommandFunc
	args          []string
	subCommand    *Command
	subCommands   subCommands
	errorHandling ErrorHandling

	output io.Writer

	Arguments Arguments
	Flags     flag.FlagSet
}

type Option func(*Command)

func UsageOption(usageStr string) Option { return func(cmd *Command) { cmd.usageStr = usageStr } }

func DescOption(description string) Option {
	return func(cmd *Command) { cmd.description = description }
}

func CallbackOption(callback CommandFunc) Option {
	return func(cmd *Command) { cmd.callback = callback }
}

func ArgCallbackOption(cb interface{}) Option {
	v := reflect.ValueOf(cb)
	if v.Kind() != reflect.Func {
		panic("Can only callback func types")
	}

	return func(cmd *Command) {
		//variables := []reflect.Value{}
		variables := []interface{}{}
		t := reflect.TypeOf(cb)
		for i := 0; i < t.NumIn(); i++ {
			switch t.In(i) {
			case reflect.TypeOf(false):
				var b bool
				cmd.Arguments.Bool(&b, "")
				variables = append(variables, &b)
			case reflect.TypeOf(time.Duration(0)):
				var d time.Duration
				cmd.Arguments.Duration(&d, "")
				variables = append(variables, &d)
			case reflect.TypeOf(float64(0)):
				var f float64
				cmd.Arguments.Float64(&f, "")
				variables = append(variables, &f)
			case reflect.TypeOf(int(0)):
				var i int
				cmd.Arguments.Int(&i, "")
				variables = append(variables, &i)
			case reflect.TypeOf(int64(0)):
				var i int64
				cmd.Arguments.Int64(&i, "")
				variables = append(variables, &i)
			case reflect.TypeOf(""):
				var s string
				cmd.Arguments.String(&s, "")
				variables = append(variables, &s)
			case reflect.TypeOf(uint(0)):
				var u uint
				cmd.Arguments.Uint(&u, "")
				variables = append(variables, &u)
			case reflect.TypeOf(uint64(0)):
				var u uint64
				cmd.Arguments.Uint64(&u, "")
				variables = append(variables, &u)
			}
		}

		cmd.callback = func(string) error {
			values := []reflect.Value{}
			for _, v := range variables {
				values = append(values, reflect.Indirect(reflect.ValueOf(v)))
			}
			ret := v.Call(values)
			if len(ret) > 0 {
				if ret[len(ret)-1].CanInterface() {
					i := ret[len(ret)-1].Interface()
					if err, ok := i.(error); ok {
						return err
					}
				}
			}
			return nil
		}
	}
}

func OutputOption(output io.Writer) Option { return func(cmd *Command) { cmd.SetOutput(output) } }

func ErrorHandlingOption(errorHandling ErrorHandling) Option {
	return func(cmd *Command) { cmd.errorHandling = errorHandling }
}

// New will return a Command object that is initialized according
// to the supplied command options
func New(name string, options ...Option) *Command {
	cmd := &Command{
		name:          name,
		output:        os.Stderr,
		errorHandling: ExitOnError,
		subCommands:   subCommands{cmds: make(map[string]*Command)},
	}

	for _, option := range options {
		option(cmd)
	}

	return cmd
}

// SubCommand adds a subcommand to the current command hierarchy
func (cmd *Command) SubCommand(name string, options ...Option) *Command {
	subCommand := New(name, options...)
	subCommand.Flags.SetOutput(cmd.output)
	cmd.subCommands.set(name, subCommand)
	return subCommand
}

// SetOutput will set the io.Writer used for printing usage
func (cmd *Command) SetOutput(writer io.Writer) {
	cmd.output = writer
}

func (cmd *Command) usage(ind *indenter) {
	// count the number of flags that have been created
	numFlags := 0
	cmd.Flags.VisitAll(func(*flag.Flag) { numFlags++ })

	if ind.count == 0 {
		ind.Indentf("Usage: %s", cmd.name)
		if cmd.usageStr != "" {
			ind.Printf(" %s\n", cmd.usageStr)
		} else {
			if numFlags > 0 {
				ind.Print(" [global options]")
			}

			if cmd.subCommands.len() > 0 {
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

	if cmd.subCommands.len() > 0 {
		ind.Indentln("Commands:")
		nameFmt := fmt.Sprintf("%%-%ds", cmd.subCommands.nameLen)
		var prevCmd *Command
		cmd.subCommands.visitAll(func(command *Command) {
			if prevCmd != nil && prevCmd.subCommands.len() == 0 && command.subCommands.len() > 0 {
				ind.Println()
			}

			ind.Indentf(nameFmt, command.name)
			if command.usageStr == "" {
				if command.description != "" {
					ind.Printf(" %s\n", command.description)
				} else {
					ind.Println()
				}
			} else {
				ind.Printf(" %s\n", command.usageStr)
				if command.description != "" {
					ind.Indentf("%s %s\n", strings.Repeat(" ", cmd.subCommands.nameLen), command.description)
				}
			}

			command.usage(&indenter{writer: ind.writer, count: ind.count + cmd.subCommands.nameLen})
			prevCmd = command
		})
		ind.Println()
	}
}

func (cmd *Command) handleErr(err error) error {
	if err != nil {
		if cmd.errorHandling == ExitOnError {
			ind := &indenter{writer: cmd.output}
			ind.Printf("%v\n", err)
			cmd.usage(ind)
			os.Exit(2)
		} else if cmd.errorHandling == PanicOnError {
			panic(err)
		}
	}
	return err
}

// Parse the arguments and make them ready to run
func (cmd *Command) Parse(args []string) (err error) {
	cmd.Flags.Parse(args)
	cmd.args = cmd.Flags.Args()

	if cmd.Arguments.Len() > 0 {
		err = cmd.handleErr(cmd.Arguments.Parse(cmd.args))
		cmd.args = cmd.Arguments.Args()
	}

	if cmd.subCommands.len() > 0 {
		if len(cmd.args) < 1 {
			err = cmd.handleErr(ErrRequiredCommand)
		}

		if err == nil {
			cmdName := cmd.args[0]
			cmd.args = cmd.args[1:]
			cmd.subCommand = cmd.subCommands.get(cmdName)
			if cmd.subCommand == nil {
				err = cmd.handleErr(ErrUnknownCommand)
			} else {
				cmd.subCommand.Parse(cmd.args)
			}
		}
	}
	return err
}

// Run the command.
func (cmd *Command) Run() (err error) {
	if cmd.callback == nil {
		if cmd.subCommand == nil {
			err = ErrNoCommandFunc
		} else {
			err = cmd.subCommand.Run()
		}
	} else {
		err = cmd.callback(cmd.name)
		if err == nil && cmd.subCommand != nil {
			err = cmd.subCommand.Run()
		}
	}

	return err
}
