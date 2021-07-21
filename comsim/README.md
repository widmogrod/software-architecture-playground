# Communication patters

Idea is to demonstrate different communication patterns between different domains.

Message passing
- Actor
- Event Sourcing
- Streaming
- Lambda
- etc

```go

domA.StreamCDC() // can stream as kafka or rabbit or gRPC or ...
domB.StreamCDC() // simulate via pooling
domC.StreamCDC() // webhooks

// defines how to invoke methods
// it has a registry
// is a lambda, can create functions on demand
r = runtime.New()

// events subscriber that orchestrates when and how to invoke business logic
// is a broker between changes and commands
// guaranties processing in order

w = eventsubscriber.New()
w.When(domA.Change, r.Invoke(aggregateA.MethodA))
w.When(domB.Change, r.Invoke(aggreatteA.MethodB))

// or
w = eventsubscriber.New()
w.When(domA.Change, r.Invoke(state.TransitionA))
w.When(domB.Change, r.Invoke(state.TransitionB))

// When(event, handler(event, state) => Either(Ok,KO)) =>
// When(KO, handler(KO))
```
