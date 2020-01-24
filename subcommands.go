package cli

import "sort"

type subCommands struct {
	cmds    map[string]*Command
	nameLen int
}

func (sc *subCommands) get(name string) *Command {
	return sc.cmds[name]
}

func (sc *subCommands) set(name string, cmd *Command) {
	sc.cmds[name] = cmd
	if len(name) > sc.nameLen {
		sc.nameLen = len(name)
	}
}

func (sc *subCommands) len() int { return len(sc.cmds) }

func (sc *subCommands) visitAll(cb func(*Command)) {
	commandNames := []string{}
	for commandName := range sc.cmds {
		commandNames = append(commandNames, commandName)
	}
	sort.Strings(commandNames)
	for _, commandName := range commandNames {
		cb(sc.cmds[commandName])
	}
}
