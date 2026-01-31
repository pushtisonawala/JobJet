package spinner

import "fmt"

type Spinner struct {
	msg string
}

func New(msg string) *Spinner {
	return &Spinner{msg: msg}
}

func (s *Spinner) Start() {
	fmt.Println(s.msg)
}

func (s *Spinner) Stop() {
	// No-op for placeholder
}
