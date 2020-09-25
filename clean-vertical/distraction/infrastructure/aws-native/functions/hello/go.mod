module github.com/widmogrod/software-architecture-playground/clean-vertical/distraction/infrastructure/aws-native/functions/hello

go 1.15

replace github.com/widmogrod/software-architecture-playground v0.0.0-20200908164406-4dffa2e08cb3 => ../../../../../../a

require (
	github.com/aws/aws-lambda-go v1.19.1
	github.com/badoux/checkmail v1.2.1 // indirect
	github.com/widmogrod/software-architecture-playground v0.0.0-20200908164406-4dffa2e08cb3
)
