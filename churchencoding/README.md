# Introduction
Having fun with **Church encoding in Golang**, or how to emulate sum types.
Work is base on the article https://www.haskellforall.com/2021/01/the-visitor-pattern-is-essentially-same.html

You can see example in this package, below you can find my thoughts on Church encoding in Golangm after experimenting with it a little.

## What I like
- **Intellectual satisfaction**. Reading and playing with this idea was fun, help me understand Visitor pattern better, and most importantly expose me to new concepts, and help me deepen my understanding on amazing people in Computer Science world whose level of knowledge is purely inspiring!
- **Power of functions**. Demonstration, power of fundamental aspects of computation, and how the Lambda Calculus is there even if you don't "see it".
- **"Defering decisions"**. I put it quotes, because data structure like Tree needs to encode how structure will be traversed. In my opinion, its limits how data structures are open for modifications. In a way this also can be a good think. Nevertheless, structure can be easily "pattern match" :)
- **"Pattern matching"**, and to be precise exhaustive pattern matching where you cannot skip new case, otherwise it won't compile. This also can be good when a new case appear, and you want to know were to change code. In golang, when you use `case-switch` on type, you loose those insides, and only explicit runtime panic can help you.

## What I don't like
- **Need to type cast**. Take a look at `preorder` [function](./tree.go) in interpretation part is `.([]int)`. We know that this translates to lack of type polymorphism in Golang
- **Cognitive load**. Idea in itself is quite exiting, but implications hider practical usecases, especially when you consider production environment were many engineers with different level of experience are moving on the codebase. Cognitive load that is required to introduce change, debug is impractical in Golang.
- **Rigid**. Recursive structure like `Tree` encodes how it can be traversed, and I cannot work with it as data structure and ie. implement easily and efficient breath-first-search
