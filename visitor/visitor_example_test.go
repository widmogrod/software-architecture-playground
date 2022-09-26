package visitor

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:generate go run cmd/govisitor/main.go -name=Vehicle -types=Plane,Car,Boat -path=visitor_example_visitor_test -packageName=visitor
type (
	Car   struct{}
	Plane struct{}
	Boat  struct{}
)

type testVisitor struct{}

func (t *testVisitor) VisitPlane(v *Plane) any { return fmt.Sprintf("Plane") }
func (t *testVisitor) VisitCar(v *Car) any     { return fmt.Sprintf("Car") }
func (t *testVisitor) VisitBoat(v *Boat) any   { return fmt.Sprintf("Boat") }

var _ VehicleVisitor = (*testVisitor)(nil)

func TestGeneratedVisitor(t *testing.T) {
	car := &Car{}
	plane := &Plane{}
	boat := &Boat{}

	visitor := &testVisitor{}
	assert.Equal(t, "Car", car.Accept(visitor))
	assert.Equal(t, "Plane", plane.Accept(visitor))
	assert.Equal(t, "Boat", boat.Accept(visitor))
}
