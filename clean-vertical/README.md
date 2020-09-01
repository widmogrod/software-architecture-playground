# Take On Clean Architecture
Having fun with a software architecture to understand where are essential parts of application, 
and where are distractions that force us to focus on type two problems.

One of constrains is to organize code in such way that is oriented for change, 
and there is no need to jump between many directories.

Way of witting in this architecture should follow iterative process where
1. `Make it work` - write essential code a.k.a. business logic, you defer any infrastructure related distraction
2. `Make it testable` -  write unit test to ensure that it;s indeed what you expect from your code
3. `Make it ____` -  TODO write interpretation
4. `Make it ____` - TODO write specification
5. `Make it ____` - Time to distraction
6. `Make it ____` -


// Raw thoughts
// - When you defer implementation, write specification what behaviour/properties you expect from someone implementation

```
.
├── distraction
│   ├── artifacts
│   │   └── sdk
│   │       ├── go-grpc
│   │       ├── go-rest
│   │       └── js-graphql-datasource
│   ├── bridge
│   │   ├── grpc
│   │   └── rest
│   ├── entrypoint
│   │   ├── api-grpc
│   │   └── cron
│   └── infrastructure
│       ├── aws-native
│       └── docker-compose
└── essence
    ├── algebra
    ├── interpretation
    │   ├── inmemory
    │   └── schema
    │       ├── casandra
    │       └── postgres
    └── usecase

```
