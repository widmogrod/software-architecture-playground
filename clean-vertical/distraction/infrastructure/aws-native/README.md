### Ideas to explore

- OpenAPI component with type [schemas to document events](https://swagger.io/specification/#schema-object) - [example inspiration](https://github.com/aws-samples/aws-serverless-ecommerce-platform/blob/master/products/resources/events.yaml)
- OpenAPI as a way to document API generate models for lambda conversion + SDK!

- Introduce database migrations
- Chaos engineering as a way of testing hypothesis
- Integration tests

### How to start local development?

Make sure everything is build and is up to date
```
sam build
sam local start-api 
```

Execute sample requests
```
http localhost:3000/hello\?name=guest 
http post localhost:3000/register_with_email Authorization:"Bearer jwt" EmailAddress=a@a.com
```

When you do changes, you need only to rebuild application, 
there is no need to restart server when you don't introduce new endpoints
```
sam build
```


When you deploy this project on AWS, then endpoints like `/hello` require username and password in DynamoDB table.
When you have this then you can just
```
curl -u <username>:<password>  https://______/Prod/hello  
```

### How to deploy CD Pipeline?
```
npm install
cdk bootstrap --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess
npx cdk synth
npx cdk deploy CleanVerticalPipeline
```


### CD pipeline
Inspired by
https://aws.amazon.com/blogs/developer/cdk-pipelines-continuous-delivery-for-aws-cdk-applications/

Notes:
- Hard-coding AWS accounts IDs for stages like pre-prod, production makes sence
- Using account ID make sense for developer context 
- Dev deployment requires attachments of direct stacks to the app `npx cdk synth && npx cdk deploy DevCleanVerticalHttpAPI`

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
