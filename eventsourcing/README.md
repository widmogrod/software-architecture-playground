# Ideas
Few things to explore with EventSourcing

## Automatic valid and in-valid state exploration testing
Mostly because Operations are separated to Handle(CMD) and Apply(Event)
What if I could list all commands that can be applied to Aggregate and
see what sequences result in success, and which don't

Consider sequence od events, some changes may result to different results:

Successful result when:
```
CreateOrder -> AddProduct -> PaymentCollected
```

Unsuccessful result, do to business constraing not allowing collect Payment to a Order without Products
```
PaymentCollected -> CreateOrder -> AddProduct
```

What if Automatic Testing could find sequence of events that when fallow result in error, 
assign probabilities and provide:
- Visualisation of all states that can be access
- Discover a pair of commands to always result to invalid result
- Discover longer sequences of events that don't make sense

It can be done exploring fact of `reproducibilit` that application of commands to an aggregate, 
should produce immutable list of events that we can be recreated state:

```go
a := NewOrderAggregate() 
a.Handle(CreateOrderCMD{})

// Reproduce aggregate from event log
aggssert.Reproducible(t, a, NewOrderAggregate())
```

Exploring states we could get matrix of valid states. 
Factorial complexity is a trouble, but exploring probability, 
maybe having fun with a tree structure we could learn and avoid patterns that most commonly result in errors or invalid states

```
   OC PC OS OD      Conditional Probability...
OC  x  √  √  √		(OC,PC) -> (PC,OS) -> (OS,OD)
PC  x  x  x  x
OS  x  x  x  x
OD  x  x  x  x
```

Example report could look like this in tests
```
[√] OrderCreateCMD               [√] OrderCollectPaymentsCMD    state=OrderAggregateState{OrderID:"1nApKSv3hsQyldA6su6V6qK95nJ", OrderCreatedAt...)
[x] OrderCollectPaymentsCMD      [x] OrderCreateCMD             err=Order does not exists
```

## Write down simple application to test idea
- TicTacToe game
- Road trip planner (collaborative adventure schedule)
- Auction Bidding platform (with Game theory twist - how to make optimise $)


## API that can reflect about itself
Aggregate could return list of actions that user can take with a given state.


## Protobuf validation
https://github.com/envoyproxy/protoc-gen-validate

## Explore languages
- https://www.choral-lang.org (research project of https://www.fabriziomontesi.com/projects/choco/)
- https://www.jolie-lang.org
- http://chor-lang.org/
- https://links-lang.org/


## Session types
https://www.researchgate.net/profile/Ugo_Deliguoro2/publication/225835882_Sessions_and_Session_Types_An_Overview/links/53d7745e0cf29f55cfb4cfc2/Sessions-and-Session-Types-An-Overview.pdf

Since session are bases on channels, then when
something happen with connection, then process may not be able to communicate
but what if we would emply idenpotency of request, consisten caching and state of operation
would be capure by ie eventsourcing aggregate?


## TicTacToe mkunion and others ideas 
- [ ] Keyed Aggregate to support Deletion (with versioning)
- [ ] storage.RepositoryWithAggregator to support index building
- [ ] DataProcessing in Golang
  - with support of versioning? 
  - with information about grouping keys?
  - with description of process rather than compiled variant of Keyed Aggregate?
- [ ] Subscription on queries
  - should solve problem with eventual consistency between storage systems
    - do it using data processing framework?
    - do it by OpenSearch percolator?
    - do it by ___?
