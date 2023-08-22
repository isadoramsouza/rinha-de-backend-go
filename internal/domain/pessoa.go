package domain

type Pessoa struct {
	ID         string `json:"id"`
	Apelido    string `json:"apelido"`
	Nome       string `json:"nome"`
	Nascimento string `json:"nascimento"`
	Stack      string `json:"stack"`
}
