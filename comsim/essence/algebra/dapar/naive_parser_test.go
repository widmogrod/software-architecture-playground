package dapar

import (
	"testing"
)

func TestNaiveParser_TrivialSpec(t *testing.T) {
	t.Skip()
	SpecRunner(t, NaiveParse, TrivialSpec)
}
func TestNaiveParser_AdvanceSpec(t *testing.T) {
	t.Skip()
	SpecRunner(t, NaiveParse, AdvanceSpec)
}
