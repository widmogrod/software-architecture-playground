# Introduction
Having fun with **Church encoding in Golang**, or how to emulate sum types.
Work is base on the article https://www.haskellforall.com/2021/01/the-visitor-pattern-is-essentially-same.html

You can see example in this package, below small snippet and my thoughts on Church encoding in Golangm after experimenting with it a little.

```go
type shape = float64

type (
	Circle    = func(float64, float64, float64) shape
	Rectangle = func(float64, float64, float64, float64) shape
	Shape     = func(Circle, Rectangle) shape
)

func area(match Shape) float64 {
    return match(func(x, y, r float64) shape {
    	// Circle
        return math.Pi * r * r
    }, func(x, y, w, h float64) shape {
    	// Rectangle
        return w * h
    })
}
```

## What I like
- **Intellectual satisfaction**. Reading and playing with this idea was fun, help me understand Visitor pattern better, and most importantly expose me to new concepts, and help me deepen my understanding on amazing people in Computer Science world whose level of knowledge is purely inspiring!
- **Power of functions**. Demonstration, power of fundamental aspects of computation, and how the Lambda Calculus is there even if you don't "see it".
- **"Defering decisions"**. I put it quotes, because data structure like Tree needs to encode how structure will be traversed. In my opinion, its limits how data structures are open for modifications. In a way this also can be a good thing. Nevertheless, structure can be easily "pattern match" :)
- **"Pattern matching"**, and to be precise exhaustive pattern matching where you cannot skip new case, otherwise it won't compile. This also can be good when a new case appear, and you want to know were to change code. In golang, when you use `case-switch` on type, you loose those insides, and only explicit runtime panic can help you.

## What I don't like
- **Need to type cast**. Take a look at `preorder` [function](./tree.go) in interpretation part is `.([]int)`. We know that this translates to lack of type polymorphism in Golang
- **Cognitive load**. Idea in itself is quite exiting, but implications hinder practical use cases, especially when you consider production environment were many engineers with different level of experience are moving on the codebase. Cognitive load that is required to introduce change, debug is impractical in Golang.
- **Rigid**. Recursive structure like `Tree` encodes how it can be traversed, and I cannot work with it as data structure and ie. implement easily and efficient breath-first-search

## What are other alternatives to missing sum types / union types / variants in Golang?
There are few. Except of course mentioned "church endcoding" and visitor patter. 

Use interface with a "tag method" that each type must im mark with:
```go
type A struct{}
type B struct{}

func (a *A) isMySum() {}
func (b *B) isMySum() {}

type mySum interface {
	isMySum()
}

func someOperation(input mySum) {
	switch intpu.(type) {
	case *A: // ... do something for A
	case *B: // ... do something for B
    }
}
```

And few more... you can find nice explanation and break down by Will Sewel in his blog post [Alternatives to sum types in Go](https://making.pusher.com/alternatives-to-sum-types-in-go/).
