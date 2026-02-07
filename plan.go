package main

type Plan struct {
	Name string
	List []Command
}

func (t *Plan) Add(lx int8, ly int8, rx int8, ry int8, during int64) Command {
	command := Command{Update: []int8{lx, ly, rx, ry}, Duration: int(during)}
	t.List = append(t.List, command)
	return command
}

type Command struct {
	Update   []int8
	Duration int
}
