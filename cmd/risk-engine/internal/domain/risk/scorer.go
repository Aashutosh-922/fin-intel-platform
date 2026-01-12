type Scorer struct {
    rules []Rule
}

func (s *Scorer) Score(ctx Context) (int, []Factor) {
    total := 0
    factors := []Factor{}

    for _, rule := range s.rules {
        score, triggered := rule.Evaluate(ctx)
        if triggered {
            total += score
            factors = append(factors, Factor{
                Name: rule.Name(),
                Contribution: score,
            })
        }
    }
    return total, factors
}
