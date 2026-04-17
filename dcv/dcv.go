package main

type VirtualSession struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
	User  string `json:"user"`
}

func NewVirtualSession() error {
	return nil
}
