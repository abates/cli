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

/*func (sc *subCommands) usage(ind *indenter) {
	commandNames := []string{}
	for commandName := range sc.cmds {
		commandNames = append(commandNames, commandName)
	}
	sort.Strings(commandNames)
	ind.Indentln("Commands:")
	nameFmt := fmt.Sprintf("%%-%ds", sc.nameLen)
	for _, commandName := range commandNames {
		command := sc.cmds[commandName]
		ind.Indentf(nameFmt, commandName)
		if command.usageStr == "" {
			if command.description != "" {
				ind.Printf(" %s\n", command.description)
			} else {
				ind.Println()
			}
		} else {
			ind.Printf(" %s\n", command.usageStr)
			if command.description != "" {
				ind.Indentf("%s %s\n", strings.Repeat(" ", sc.nameLen), command.description)
			}
		}

		if len(command.subCommands.cmds) > 0 {
			command.usage(&indenter{writer: ind.writer, count: ind.count + sc.nameLen})
		}
	}
	ind.Println()
}*/
