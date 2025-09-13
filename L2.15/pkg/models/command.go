package models

type Cmd struct {
	Binary    string
	Arguments []string
	FileOps   []FileOperation
	NextPipe  *Cmd
	NextAnd   *Cmd
	NextOr    *Cmd
}

type FileOperation struct {
	Operation string // ">", "<"
	Filename  string
}

func (c *Cmd) Empty() bool {
	if c == nil {
		return true
	}
	return c.Binary == "" &&
		len(c.Arguments) == 0 &&
		len(c.FileOps) == 0 &&
		c.NextPipe == nil &&
		c.NextAnd == nil &&
		c.NextOr == nil
}
