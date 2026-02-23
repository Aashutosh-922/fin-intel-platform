package risk

type Factor struct {
	Name   string
	Score  int
	Reason string
}

type Rule interface {
	Name() string
	Evaluate(ctx Context) (Factor, bool)
}