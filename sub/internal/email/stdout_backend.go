package email

import "fmt"

type StdoutBackend struct{}

func NewStdoutBackend() *StdoutBackend {
	return &StdoutBackend{}
}

func (s *StdoutBackend) Send(to, subject, body string) error {
	fmt.Println("To:", to)
	fmt.Println("Subject:", subject)
	fmt.Println("Body:", body)
	return nil
}
