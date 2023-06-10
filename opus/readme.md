# Opus
## Development
```
brew instal localstack colima
pip install awscli-local

# alias npm="node --dns-result-order=ipv4first $(which npm)"
npm install -g aws-cdk-local
```

```
colima start
localstack start 
// or docker run --rm -it -p 4566:4566 -p 4510-4559:4510-4559 localstack/localstack
awslocal s3 ls
```