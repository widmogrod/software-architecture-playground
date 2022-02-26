#Introduction
Welcome to a project where I explore different ways of solving a business problems in Golang using different Software Architecture, Computational Models, that can help me strike following goals

## Traits that I search in Software Architecture
### Scales Development Teams
I'm interested in learning how a software architecture will behave, when company grows and new engineers with different skill levels join the project. 
- How easy it is to introduce change by a junior developer? 
- How easy is to form new team around solution without slowing down everyone? 

### Enable Engineers Productivity by NOT distracting*
We could say a lot of things about engineers productivity, but my definition is that Software Architecture enable engineers productivity by not distracting engineers! 
You maybe experience this by yourself, when you have mental model of a problem, and you just implement working solution in "few breaths". 

It's a state where when you need any components like database, object storage, cache, authorisation, etc. you have it and you use it, and what's the best, there is not chasm between local and production development, it works everywhere.
It's a state where you solve hard business problems, and don't think how to run migration on a database without introducing backward incompatible change and cause downtime

### Delivers Quality
I'm thinking here about quality that customers perceive, which simply means be ready to serve the customers needs, without them need to notice that technology is in their way.
On a level of Software Architecture it translates to scalability, security and all necessary traits that make technology transparent for customers.

## Experiments with Software Architecture
Currently, you can find those experiments, some of them are not connected to those values, but nothertheless serve purpose to discover them
- [Experiment with Clean Architecture and Vertical Architecture in Golang](./clean-vertical/README.md)
- [Experiment with EventSourcing in Golang](./eventsourcing/README.md)
- [Church encoding in Golang](./churchencoding/README.md)
- [McCarthy's Ambiguous Operator & SatSolver in Golang](./continuations/README.md)

## Other links
- [Latency Numbers Every Programmer Should Know](https://colin-scott.github.io/personal_website/research/interactive_latency.html)