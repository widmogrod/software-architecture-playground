package probability

type Variable = string
type Probability = float64
type Occurrences = float64

func newProbabilitySpace() *probabilitySpace {
	return &probabilitySpace{
		total:             0,
		marginal:          make(map[Variable]Occurrences),
		conditional:       make(map[Variable]map[Variable]Probability),
		mutuallyExclusive: make(map[Variable]map[Variable]struct{}),
	}
}

type probabilitySpace struct {
	total             Occurrences
	marginal          map[Variable]Occurrences
	conditional       map[Variable]map[Variable]Probability
	mutuallyExclusive map[Variable]map[Variable]struct{}
}

func (p *probabilitySpace) Increment(x Variable) {
	p.AddOccurrences(x, 1)
}

func (p *probabilitySpace) AddOccurrences(x Variable, o Occurrences) {
	p.marginal[x] += o
	p.total += o
}

func (p *probabilitySpace) Marginal(x Variable) Probability {
	return p.marginal[x] / p.total
}

func (p *probabilitySpace) P_given(x Variable, given Variable, f Probability) {
	if _, ok := p.conditional[x]; !ok {
		p.conditional[x] = make(map[Variable]Probability)
	}
	p.conditional[x][given] = f
}

func (p *probabilitySpace) P(x Variable) Probability {
	return p.Marginal(x)
}

func (p *probabilitySpace) P_and(x, y Variable) Probability {
	// TODO most likely this If is pointless
	if _, ok := p.mutuallyExclusive[x]; ok {
		if _, ok := p.mutuallyExclusive[x][y]; ok {
			// When event A and B cannot occur simultaneously
			// They are called mutually exclusive events
			// P (A and B) = P(A) * P(B) = 0
			return 0
		}
	}

	if _, ok := p.conditional[x]; ok {
		if _, ok := p.conditional[x][y]; ok {
			// P(A and B) = P(A|B)*P(B)
			return p.conditional[x][y] * p.P(y)
		}
	}

	// Given Independent
	// P(A and B) = P(A) * P(B)
	return p.P(x) * p.P(y)
}

func (p *probabilitySpace) P_or(x, y Variable) Probability {
	// When event A and B cannot occur simultaneously
	// They are called mutually exclusive events
	// P (A and B) = P(A) * P(B) = 0
	// P (A or B) = P(A) + P(B) - P(A and B)

	return p.P(x) + p.P(y) - p.P_and(x, y)
}

//func (p *probabilitySpace) GivenIndependent(x, y Variable) {
//
//}

func (p *probabilitySpace) GivenMutuallyExclusive(x, y Variable) {
	if _, ok := p.mutuallyExclusive[x]; !ok {
		p.mutuallyExclusive[x] = make(map[Variable]struct{})
	}
	p.mutuallyExclusive[x][y] = struct{}{}
}

func (p *probabilitySpace) GivenMutuallyExclusiveEvents(xs ...Variable) {
	var prev Variable
	for _, x := range xs {
		if prev != "" {
			p.GivenMutuallyExclusive(prev, x)
		}

		p.Increment(x)

		prev = x
	}
}
