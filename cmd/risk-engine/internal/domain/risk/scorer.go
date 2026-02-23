package risk

type Scorer struct {
	rules []Rule
}

func NewScorer(rules []Rule) *Scorer {
	return &Scorer{rules: rules}
}

func (s *Scorer) Score(ctx Context) (int, []Factor) {
	total := 0
	var factors []Factor

	for _, rule := range s.rules {
		if factor, triggered := rule.Evaluate(ctx); triggered {
			total += factor.Score
			factors = append(factors, factor)
		}
	}
	return total, factors
}