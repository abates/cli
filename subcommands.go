package cli

import "sort"

type subCommands []*Command

func (s subCommands) Len() int           { return len(s) }
func (s subCommands) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s subCommands) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s subCommands) maxLen() int {
	l := 0
	for _, cmd := range s {
		if len(cmd.Name) > l {
			l = len(cmd.Name)
		}
	}
	return l
}

func (s subCommands) get(name string) *Command {
	s.sort()
	i := sort.Search(len(s), func(i int) bool { return s[i].Name >= name })
	if i != len(s) {
		if s[i].Name == name {
			return s[i]
		}
	}
	return nil
}

func (s subCommands) sort() {
	sort.Sort(s)
}
