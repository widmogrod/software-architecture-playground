package sat

//
//import (
//	"fmt"
//	"image/color/palette"
//)
//
///*
//
//Because there is quite big probability of variables forming same preposition pairs
//idea is to group them, and forming preposition tree[?]
//that can be traverse and backtrack when we find states that are false
//
//
//a b
//-b
//-a
//
//
//1. Find closures with one variable, because they must be true to solve constrains
//2. and rewrite constrains to eliminate opposite values in other closures
//
//a _
//-b
//-a
//
//3. Repeat.
//
//_ _
//-b
//-a
//
//4. Break when there are closures without values that cannot be proven
//
//What IF it looks like this?
//_ -b
//-b
//a
//
//then it solved, no free variables
//
//What if it looks like this?
//_ -b c
//-b
//a
//
//We can ignore c, because it does not play role for solution,
//but when we would like to be exhaustive,
//then we set value for c as true because there are not other connections
//
//What if it looks like this?
//a -b c
//-b c
//
//We can form groups, and solve for groups rather than other
//
//
//What if group look like this?
//a -b c
//b -c
//
//We can say that they're inverse on or ther must be true, but not both
//And in case they're not ordered, then Order statements!
//
//		b      -b
//       /  =    /
//      c      -c
//
//Link variables that must be opossite, and cannot both be true.
//- This step would reduce search space, because changes would also be constrained
//		-b       b
//       /   or   /
//      c      -c
//
//How data structure could look like that could could reduce space of proof?
//
//	a -b c b -c			a -b x b √		a √ x x √		x √ x x √
//	a b -c				a b √			a x √			x x √
//	b -c				b √ 			x √ 			x √
//
//	d					d				d				d
//
//min heap - where element that has smaller index must be distcharged firss
//max heap - where elements most buiquoutes are discharged first,
//			and they have option to backtrack and change opposite value
//
//			d(1)
//			 /	\
//
//
//			 c(3)
//			/   \
//		  b
//
//		a
//		 \
//		  -b b b c
//			\ \ \
//			 c -c -c -c
//
//				a
//			1 /
//			 b
//		       \ -1
//				-c
//			 -1 /  \ -1
//			 		x
//
//			 [  d ]
//			/   	\
//		 [ b ]		[ -c ]
//		      	   /	  \
//				[ a ]
//
//
//
//	a -b  c
//	a  b -c
//	   b -c
// 			 d
//
//*/
//
//type Vertex struct {
//	id VertexID
//}
//
//type Edge struct {
//	// -1 or 1
//	typ           string
//	closureLineNo int
//	from, to      VertexID
//}
//
//func (e *Edge) Equal(e2 *Edge) bool {
//	if e.to != e2.to {
//		return false
//	}
//	if e.from != e2.from {
//		return false
//	}
//	if e.typ != e2.typ {
//		return false
//	}
//	if e.closureLineNo != e2.closureLineNo {
//		return false
//	}
//
//	return true
//}
//
//func (s *solver) mkVertex(prep Preposition) *Vertex {
//	switch x := prep.(type) {
//	case *BoolVar:
//		v := Vertex{id: s.indexes[x]}
//		return &v
//	case *negation:
//		return s.mkVertex(x.b)
//	case *implication:
//		panic("don't support implication")
//	}
//
//	return nil
//}
//
//func (s *solver) typ(prep Preposition) string {
//	switch prep.(type) {
//	case *BoolVar:
//		return "1"
//	case *negation:
//		return "-1"
//	case *implication:
//		panic("don't support implication")
//	}
//
//	return "unknown-typ"
//}
//
//var _ Preposition = &DecisionVar{}
//
//type DecisionVar struct {
//	index int
//}
//
//func (d *DecisionVar) Not() Preposition {
//	panic("implement me")
//}
//
//func (d *DecisionVar) IsTrue() bool {
//	panic("implement me")
//}
//
//type State struct {
//	closures Closures
//	solves   map[Preposition][]int
//}
//
//func (s *solver) start(st *State) {
//	for lineNo, line := range st.closures {
//		for _, prep := range line {
//			switch x := prep.(type) {
//			case *BoolVar:
//				s.assumeTrueSAT(x, st)
//			case *negation:
//				s.assumeFalseSAT(x, st)
//
//			}
//
//		}
//	}
//}
//
//func (s *solver) assumeTrueSAT(x *BoolVar, st *State) {
//	//st.solves = append(st.solves, x)
//	closures := [][]Preposition{}
//	for lineNo, line := range st.closures {
//		for _, prep := range line {
//			switch x := prep.(type) {
//			case *BoolVar:
//				st.solves[prep] = append(st.solves[prep], lineNo)
//				s.assumeTrueSAT(x, st)
//			case *negation:
//				s.assumeFalseSAT(x, st)
//
//			}
//		}
//	}
//
//}
//func (s *solver) assumeFalseSAT(x *negation, st *State) {
//	st.solves = append(st.solves, x)
//}
//
//func (s *solver) OptimizedSolution() {
//	g := NewDAG()
//	_ = s.buildGraph(g)
//	//g.Print()
//
//	s.start(&State{
//		closures: s.closures,
//		solves:   nil,
//	})
//
//	//state := &SolState{
//	//	Assume:       map[VertexID]bool{},
//	//	LinesToSolve: len(s.closures),
//	//	LinesSolved:  map[int]bool{},
//	//}
//
//	//s.solve(g, v0, state)
//
//	//fmt.Printf("solved for: %v", state.Assume)
//}
//
//type SolState struct {
//	Assume       map[VertexID]bool
//	LinesToSolve int
//	LinesSolved  map[int]bool
//}
//
////func (s *solver) solve(g *DAG, v *Vertex, state *SolState) bool {
////	if state.LinesToSolve == 0 {
////		return true
////	}
////
////	edges := g.Edges(v.id)
////	if len(edges) == 0 {
////		// orphan, it must be true
////		if val, found := state.Assume[v.id]; found {
////			if val == false {
////				panic(fmt.Sprintf("this value x%d must be assume to be true", v.id))
////			}
////		} else {
////			state.Assume[v.id] = true
////		}
////
////		return true
////	}
////
////	for _, e := range edges {
////		// don't solve solved lines
////		if _, solved := state.LinesSolved[e.closureLineNo]; solved {
////			continue
////		}
////
////		state.LinesSolved[e.closureLineNo] = true
////		state.LinesToSolve--
////
////		// Edge states what value, target node should assume
////		if _, found := state.Assume[e.to]; found {
////			continue
////		}
////
////		if e.typ == "1" {
////			state.Assume[e.to] = true
////			if s.solve(g, e.to, state) {
////
////			}
////		} else {
////			state.Assume[e.to] = false
////		}
////	}
////}
//
//func (s *solver) buildGraph(g *DAG) *Vertex {
//	v0 := Vertex{id: -1}
//	g.SetVertex(v0)
//
//	for lineNo, ors := range s.closures {
//		var prev *Vertex
//		for _, or := range ors {
//			if prev == nil {
//				prev = s.mkVertex(or)
//				g.SetVertex(*prev)
//				g.SetEdge(Edge{
//					typ:           s.typ(or),
//					closureLineNo: lineNo,
//					from:          v0.id,
//					to:            prev.id,
//				})
//			} else {
//				next := s.mkVertex(or)
//				g.SetVertex(*next)
//				g.SetEdge(Edge{
//					typ:           s.typ(or),
//					closureLineNo: lineNo,
//					from:          prev.id,
//					to:            next.id,
//				})
//				prev = next
//			}
//
//		}
//	}
//
//	return &v0
//}
//
////type DecisionVar struct {
////	variable  *BoolVar
////	trueNext  *DecisionVar
////	falseNext *DecisionVar
////	parent    *DecisionVar
////}
////
////func (t *DecisionVar) Has() bool {
////	return t.variable != nil
////}
////
////func (t *DecisionVar) ActiveBranch() *DecisionVar {
////	return t.value
////}
////
////func (t *DecisionVar) CanBacktrack() bool {
////	return t.parent != nil
////}
////
////func (t *DecisionVar) Next() *DecisionVar {
////	if t.trueNext == nil && t.falseNext == nil {
////		true := &DecisionVar{
////			variable:  nil,
////			trueNext:  nil,
////			falseNext: nil,
////			parent:    t,
////		}
////		false := &DecisionVar{
////			variable:  nil,
////			trueNext:  nil,
////			falseNext: nil,
////			parent:    t,
////		}
////
////		t.trueNext = true
////		t.falseNext = false
////	}
////
////	return nil
////}
////
////func (t *DecisionVar) SetNext(variable *BoolVar) {
////	if t.Has() {
////		panic("next value already set")
////	}
////
////	t.variable = variable
////}
////
////func FindSolution(decision *DecisionVar, closures Closures) bool {
////	if len(closures) == 0 {
////		return true
////	}
////
////	if variable := decision.ActiveBranch(); variable != nil {
////		notSatisfied := Closures{}
////		for _, ors := range closures {
////			if !Satisfies(variable, ors) {
////				notSatisfied = append(notSatisfied, ors)
////			}
////		}
////
////		return FindSolution(decision.Next(), notSatisfied)
////	}
////
////	variable, has := PickVariable(notSatisfied)
////	if has {
////		decision.SetNext(variable)
////		return FindSolution(decision, closures)
////	}
////
////	if decision.CanBacktrack() {
////		return FindSolution(decision.Backtrack(), closures)
////	}
////
////	return false
////}
