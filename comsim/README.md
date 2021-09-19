# Communication patters

Idea is to demonstrate different communication patterns between different domains,
that share one trait message passing.

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

w = messagesubscriber.New()
w.When(domA.Change, r.Invoke(aggregateA.MethodA))
w.When(domB.Change, r.Invoke(aggreatteA.MethodB))

// or
w = messagesubscriber.New()
w.When(domA.Change, r.Invoke(state.TransitionA))
w.When(domB.Change, r.Invoke(state.TransitionB))

// When(event, handler(event, state) => Either(Ok,KO)) =>
// When(KO, handler(KO))
```

Stream of consciousness, beware:

In this project there is few artefacts
- golang generator of union types
- workpar a non-turing complete language for expresing workflows, that transpiles to AWS StepFunctions
- invoker - abstraction for function invocations, that can move computation from local to distributed
- streamer - abstraction on data streams, that was also used as an example of stream base invocations

All those experiments, are tools for thoughs to understand 
how to design or identify a language/solution 
that allow to write code at scale it when scaling is needed, 
removing need to think about it from day zero and all 
intricacies of distributed systems from code perspective
- transactions, consistency, resiliency, redundancy

There are few solutions to consistency that I know
- transactions - history based (RDMS,...)
- transactions - property based (CRDT,...)

They work fantasticly when we have priviladge to build around them, 
but when we have to compose bigger and bigger systems, especialy with microservices; 
some may use Redis, some may use PostgreSQL, some may use Kafka
and we want those components to be share consistent view about word 
- we may choose to have single source of thruth 
  - we may choose to have single database
  - and build dedicated projections form it (CQRS)
  - and use orchestration to coordinate communication (Workflow)
- we may also choose to ignore it and struggle with inconsistencies, 
  which may not be bad decisions in areas where we have locality of data or inconsistency is accepted trait

Transactions have nature of all-or-nothing
Workflows have nature of when everything is up, eventually we will have desired state

We could write code to coordinate communication between few sources of thruth as
- sequential code in any language, and when N function call fails we could retry whole sequence from start 0..N
  - State management must be written explicitly
  - Managing failures and retries done manually
  - Process description in one place
  - Sequencial process blocks and waits unit process is complete

- sequence of steps, where only steps that are failing are retry or direct to other operations; workflow language
  - Workflow runtime manage state and retries of invocations
  - Process description in one place
  - Workflow is non-blocking 
   
- reacting on message from different domains, it could be via subscribing to a changes log, or by consuming messages from inbox. 
  - Consuming process build and manage state.
  - Is independent of availability of specific components, 
    but this doesn't mean it can produce results when one of services that produce events is down
  - When messages are events then managing failure is much simplere, since something already happen and we need to only handle it, 
    and optionally manage internal state failures
    - Process is distributed and change requires modification across different services, sometimes with coordination


Thought process and components in this package aim to find a middle ground between mentioned styles,
Decoupling components like function from invocation creates components like
- function registry, that should also host input-output definitions; 
  - such feature enables possibility to deploy and colocate functions that need to be close
  - enables possibility to have function in different languages
  - enables possibility to type checking
- workflow, sequential coordination is possible when function are just invocations
  - workflow enables possibility to colocate functions close (deployment planner necessary)
  - workflow lifts retries, failure management, tracing to runtime
  - workflow is stateless so it can scale
- abstracting functions invocation, makes workflow process non-blocking, 
  - which enables logging function invocations and results, 
  - and append log of such events can enable total ordering which enables future capabilities like event listeners

To complete demonstration of such runtime I need to 
- [√] abstract function invocations
- [√] create function registry (without types first)
- [√] create workflow language that leverage function registry
- [√] example of in-memory workflow coordination and execution
- [√] example of transpiling workflow to AWS StepFunction

Other areas to experiment:
- [_] namespacing of flows can help with big teams and projects to stay independent
- [_] type checking and inference in function registry and flow language
- [_] generate lambda handlers - "transpile" to AWS Lambda or K8s
- [_] leverage flow information to collocate functions so invocation are efficient
- [_] example of complex coordination with external triggers (like webhooks)

From this exercise I notice that there is potential in further research in:
- function registry & deployment like AWS Lambda may in feature
  - introduce function input-output types declaration/types registry
  - introduce collocated deployments of functions
- type of such functions may have additional properties to indicate consistency, like write your own reads, 
  - type inference and checking may leverage consistency information to validate workflow structure
- intermediary non-turing complete language that I created give possibility to adapt other languages to write workflows,
  - potentially similar capability can be achieved in WASM and I should look into this more

Other things that I learn:
- golang code generation can be well integrated in daily workflow just: `go generate; go test; go build`