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


Raw thoughts
- When you defer implementation, write specification what behaviour/properties you expect from someone implementation
- Spec tests should not contain testing/check since they may be used as integration tests and to big cardinality of inputs will increase integration tests
- Spec tests can be parametrised
- The more I work with separation on essence and distraction, the more boilerplate code I see that is a must `.github/dependabot.yaml`, `ci.yml`, ...
  One idea that pops in my mind... what if I could not need to write different .yaml files but just generate them? or even better lift development of 
  the project to meta-repository structure where I'm not distracted by code that connects many external services.
  I could focus purely on writing code that matters. One project that is interesting [projen](https://github.com/eladb/projen)
  
  
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


TODO
- [_] Single source of truth for artifacts. Github workflow actions should invoke either bash or Makefiles to generate artifacts like golang & js sdk. 
This going to separate invocation from implementation, and enable developers to generate artifacts locally.
- [_] Introduce AWS AppConfig and see what role it can play with canary deployments, chaos engineering, etc.
- [_] Play with AWS StepFunctions as a validation of concept of independent implementation from interpretation.