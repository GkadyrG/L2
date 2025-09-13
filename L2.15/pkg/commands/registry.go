package commands

import (
	"io"

	"github.com/GkadyrG/L2/L2.15/pkg/environment"
)

type Handler interface {
	Name() string
	Run(args []string, env environment.Env, in io.Reader, out io.Writer) error
}

type CommandRegistry struct {
	handlers map[string]Handler
}

func NewRegistry() *CommandRegistry {
	r := &CommandRegistry{
		handlers: make(map[string]Handler),
	}

	// Регистрируем встроенные команды
	r.Add(&ChangeDirectory{})
	r.Add(&PrintWorkingDirectory{})
	r.Add(&Echo{})
	r.Add(&ProcessList{})
	r.Add(&KillProcess{})

	return r
}

func (r *CommandRegistry) Add(h Handler) {
	r.handlers[h.Name()] = h
}

func (r *CommandRegistry) Get(name string) (Handler, bool) {
	h, exists := r.handlers[name]
	return h, exists
}
