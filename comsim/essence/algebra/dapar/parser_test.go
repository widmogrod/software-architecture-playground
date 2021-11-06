package dapar

import (
	"testing"
)

func TestParser_TrivialSpec(t *testing.T) {
	SpecRunner(t, Parse, TrivialSpec)
}
func TestParser_AdvanceSpec(t *testing.T) {
	SpecRunner(t, Parse, AdvanceSpec)
}
