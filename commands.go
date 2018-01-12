package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type flagWriter struct {
	indent string
	writer io.Writer
}

func (f flagWriter) Write(p []byte) (n int, err error) {
	for _, line := range strings.Split(string(p), "\n") {
		fmt.Fprintf(f.writer, "%s%s", f.indent, line)
		if strings.TrimSpace(line) != "" {
			fmt.Fprintf(f.writer, "\n")
		}
	}
	return len(p), nil
}

// CommandFunc is the callback function that will be executed when a
// command is called
type CommandFunc func(args []string, subCommand *Command) error

// Command represents a single cli command. The idea is that a cli app
// is run such as:
//    program cmd <flags>
// and can have nested commands:
//    program cmd1 <flags> subcmd1 <flags> ...
// a Command object represents a single command in the hierarchy and is
// a placeholder to register subcommands
type Command struct {
	indent      string
	name        string
	usageStr    string
	description string
	callback    CommandFunc
	subCommands map[string]*Command
	out         io.Writer
	Flags       *flag.FlagSet
}

// Register a subcommand
func (c *Command) Register(name, usageStr, description string, callback CommandFunc) *Command {
	subCommand := &Command{
		name:        name,
		usageStr:    usageStr,
		description: description,
		callback:    callback,
		subCommands: make(map[string]*Command),
		out:         c.out,
		Flags:       flag.NewFlagSet(name, flag.ExitOnError),
	}
	c.subCommands[name] = subCommand
	return subCommand
}

func (c *Command) usage() {
	maxNameLen := 0
	var commandNames []string
	for name := range c.subCommands {
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
		commandNames = append(commandNames, name)
	}
	nameFmt := fmt.Sprintf("%s%%%ds %%s\n", c.indent, maxNameLen)
	sort.Strings(commandNames)

	if c.indent == "" {
		fmt.Fprintf(c.out, "Usage: %s [global options]", c.name)
		if len(c.subCommands) > 0 {
			fmt.Fprintf(c.out, " <command> [command options]\n")
		} else {
			fmt.Fprintf(c.out, "\n")
		}
	}
	c.Flags.SetOutput(&flagWriter{c.indent, c.out})
	c.Flags.PrintDefaults()

	if len(commandNames) > 0 {
		indent := strings.Repeat(" ", maxNameLen)
		fmt.Fprintf(c.out, "%sCommands:\n", c.indent)
		for _, commandName := range commandNames {
			command := c.subCommands[commandName]
			fmt.Fprintf(c.out, nameFmt, commandName, command.usageStr)
			if command.description != "" {
				fmt.Fprintf(c.out, "%s%s %s\n", c.indent, indent, command.description)
			}
			indent := fmt.Sprintf("%s%s  ", strings.Repeat(" ", maxNameLen), c.indent)
			command.indent = indent
			command.usage()
		}
		fmt.Fprintf(c.out, "\n")
	}
}

// Run the command. The first argument is the command that will
// be looked up in the list of subcommands. If found, the corresponding
// CommandFunc will be called with the next command (subcommand) that needs
// to be called. This allows a chaining effect for the commands so that
// setup and teardown can be performed by calling the subcommand from
// within the CommandFunc callback
func (c *Command) Run(args []string) (err error) {
	c.Flags.Parse(args)
	args = c.Flags.Args()
	var subCommand *Command

	if len(c.subCommands) > 0 {
		if len(c.Flags.Args()) < 1 {
			// TODO make this an error we can use in a conditional
			c.usage()
			os.Exit(2)
		}

		cmdName := args[0]
		args = args[1:]
		subCommand = c.subCommands[cmdName]
		if subCommand == nil {
			// TODO make this an error we can use in a conditional
			fmt.Fprintf(os.Stderr, "Unknown command %s\n", cmdName)
			os.Exit(3)
		}
	}

	if c.callback != nil {
		err = c.callback(args, subCommand)
	}

	return err
}
