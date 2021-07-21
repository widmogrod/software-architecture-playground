package probability

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProbabilitySpace(t *testing.T) {
	p := newProbabilitySpace()
	p.Increment("A")
	p.Increment("B")
	p.P_given("A", "B", 0.10)

	// P_given probability
	// P(A|B) = P(A ^ B) / P(B)
	// P_given independence under C
	// P(A|B,C) = P(A|C)*P(B|C)

	//p.GivenIndependent("A", "B")
	p.GivenMutuallyExclusive("A", "B")

	assert.EqualValues(t, .5, p.P("A"))
	assert.EqualValues(t, 0, p.P_and("A", "B"))
	assert.EqualValues(t, 1, p.P_or("A", "B"))
}

func TestProbabilityRollingDice(t *testing.T) {
	p := newProbabilitySpace()
	p.GivenMutuallyExclusiveEvents("1", "2", "3", "4", "5", "6")

	assert.EqualValues(t, 1./6., p.P("1"), "Probability or rolling 1 on six-sided fair dice should be 1/6")
	assert.EqualValues(t, 2./6., p.P_or("1", "2"), "Probability of rolling 1 or 2 on first throw of six-sided fair dice should be 1/3 ")
	assert.EqualValues(t, 0, p.P_and("A", "B"), "Probability of rolling 1 and 2 on first throw of six-sided fair dice should be 0")
}

func TestProbabilityRollingANumberOnDiceTwiceInRow(t *testing.T) {
	p := newProbabilitySpace()
	p.GivenMutuallyExclusiveEvents("1", "2", "3", "4", "5", "6")

	// P(1) ^ P(2) and -(P(1) and P(3)) and -P(1) and P(3)
	assert.EqualValues(t, 1./6., p.P("1"), "Probability or rolling 1 and then 2 on six-sided fair dice should be 1/6")
}
