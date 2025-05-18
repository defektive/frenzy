package transformer

import (
	"log"
)

type ITransformer interface {
	Name() string
	Description() string
	Priority() int
	Process(in []byte) (out []byte, err error)
}

type Transformer struct {
	name        string
	description string
	priority    int
	processFunc func([]byte) (out []byte, err error)
}

func (t *Transformer) Name() string {
	return t.name
}

func (t *Transformer) Description() string {
	return t.description
}

func (t *Transformer) Priority() int {
	return t.priority
}

func (t *Transformer) Process(in []byte) (out []byte, err error) {
	return t.processFunc(in)
}

func NewTransformer(name string, description string, priority int, processFunc func([]byte) (out []byte, err error)) *Transformer {
	return &Transformer{
		name:        name,
		description: description,
		priority:    priority,
		processFunc: processFunc,
	}
}

func Run(input []byte, things []Transformer) []byte {
	for _, t := range things {
		out, err := t.Process(input)
		if err != nil {
			log.Printf("Error processing %s: %s\n", t.Name(), err)
			// it errored, just use input on the next one
			out = input
		}

		input = out
	}

	return input
}

//type Rule struct {
//	Name    string `json:"name"`
//	Search  string `json:"search"`
//	Replace string `json:"replace"`
//}
//
//func (r Rule) Send() string {
//
//}
//
//func (r Rule) Receive() string {
//
//}
