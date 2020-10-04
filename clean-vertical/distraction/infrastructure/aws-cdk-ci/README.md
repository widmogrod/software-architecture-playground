# CD pipeline
Inspired by
https://aws.amazon.com/blogs/developer/cdk-pipelines-continuous-delivery-for-aws-cdk-applications/

Notes:
- Hard-coding AWS accounts IDs for stages like pre-prod, production makes sence
- Using account ID make sense for developer context 

Pros
- Definition is unified, transparent, consistent and easy to change
- Pipeline can update itself
- Clearly speparated developer account
- Clearly define way for setting up Integration tests
- You can express almost everything, and AWS services connection follow connection in code

Cons-ish
- Pipeline takes some time, compared to what I'm used to (1min CD) - but benefits are overwhelming to what I used to
- Remember about `cdk bootstrap with --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess` - soon it should be standard
- Remove construct, and it will be removed from AWS - that means that Database can be drop! Linter that could detect breaking changes would be beneficial.