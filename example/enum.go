package example

//go:generate go run ../ --type=Fruit --line-comment ./
type Fruit uint8

const (
	Apple  Fruit = iota // apple
	Banana              // bananana
	Cherry              // cherries
)
