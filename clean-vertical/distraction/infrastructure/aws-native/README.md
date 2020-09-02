### How to start local development?

Make sure everything is build and is up to date
```
sam build
sam local start-api 
```

Execute sample requests
```
http post http://127.0.0.1:3000/register_with_email EmailAddress=a@a.com  
```

When you do changes, you need only to rebuild application, 
there is no need to restart server when you don't introduce new endpoints
```
sam build
```
