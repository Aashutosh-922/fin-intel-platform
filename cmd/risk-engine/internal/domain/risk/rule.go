type Rule interface {
    Name() string
    Evaluate(ctx Context) (score int, triggered bool)
}
