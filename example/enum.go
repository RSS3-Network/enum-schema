package example

//go:generate go run ../ --type=Fruit --trimprefix=F --transform=snake  --indent
type Fruit uint8

const (
	FApple  Fruit = iota // apple
	FBanana              // bananana
	FCherry              // cherries
)
