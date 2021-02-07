package churchencoding

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShape(t *testing.T) {
	var (
		exampleCircle    Shape = _Circle(0.2, 1.4, 4.5)
		exampleRectangle Shape = _Rectangle(1.3, 3.1, 10.3, 7.7)
	)

	assert.InEpsilon(t, 63.6, area(exampleCircle), 0.1)
	assert.InEpsilon(t, 79.3, area(exampleRectangle), 0.1)
}
