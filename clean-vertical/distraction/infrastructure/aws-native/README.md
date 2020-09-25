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
