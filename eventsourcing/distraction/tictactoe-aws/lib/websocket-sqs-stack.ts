import * as cdk from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as lambdanodejs from "aws-cdk-lib/aws-lambda-nodejs";
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import * as apigatewayv2 from '@aws-cdk/aws-apigatewayv2-alpha';
import * as apigatewayv2integrations from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import {SqsEventSource, DynamoEventSource} from 'aws-cdk-lib/aws-lambda-event-sources';
import * as golang from '@aws-cdk/aws-lambda-go-alpha';
import * as python from '@aws-cdk/aws-lambda-python-alpha';
import * as opensearchservice from "aws-cdk-lib/aws-opensearchservice";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as iam from "aws-cdk-lib/aws-iam";
import * as kinesis from "aws-cdk-lib/aws-kinesis";
import * as ecr_assets from "aws-cdk-lib/aws-ecr-assets";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as ecsp from "aws-cdk-lib/aws-ecs-patterns";


export class WebsocketSqSStack extends cdk.Stack {
    // constructor for the stack
    constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const queue = new sqs.Queue(this, 'tic-tac-toe-queue-sqs', {
            queueName: 'tictactoe-sqs-queue',
            // visibilityTimeout: cdk.Duration.seconds(20),
            // receiveMessageWaitTime: cdk.Duration.seconds(5),
        })

        const stream = new kinesis.Stream(this, 'Stream', {
            streamName: 'tictactie',
            shardCount: 1,
            retentionPeriod: cdk.Duration.hours(48),
        });

        const table = new dynamodb.Table(this, 'WebsocketSQSConnections', {
            partitionKey: {
                name: 'ID',
                type: dynamodb.AttributeType.STRING
            },
            sortKey: {
                name: 'Type',
                type: dynamodb.AttributeType.STRING
            },
            removalPolicy: cdk.RemovalPolicy.DESTROY, // not recommended for production code!
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            pointInTimeRecovery: false,
            stream: dynamodb.StreamViewType.NEW_AND_OLD_IMAGES,
            kinesisStream: stream,
        });

        const domain = new opensearchservice.Domain(this, 'DynamoDBProjection', {
            domainName: 'dynamodb-projection',
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            version: opensearchservice.EngineVersion.OPENSEARCH_1_3,
            fineGrainedAccessControl: {
                masterUserName: 'admin',
                masterUserPassword: cdk.SecretValue.unsafePlainText('nile!DISLODGE5clause')
            },
            capacity: {
                masterNodes: 0,
                dataNodes: 1,
                dataNodeInstanceType: ec2.InstanceType.of(
                    ec2.InstanceClass.T3,
                    ec2.InstanceSize.SMALL
                ).toString() + ".search",
            },
            ebs: {
                volumeSize: 20,
                volumeType: ec2.EbsDeviceVolumeType.GENERAL_PURPOSE_SSD,
            },
            zoneAwareness: {
                enabled:false,
            },
            logging: {
                slowSearchLogEnabled: true,
                appLogEnabled: true,
                slowIndexLogEnabled: true,
            },
            enforceHttps: true,
            encryptionAtRest: {enabled: true},
            nodeToNodeEncryption: true,
        });

        domain.addAccessPolicies(
            new iam.PolicyStatement({
                actions: ['es:*'],
                effect: iam.Effect.ALLOW,
                principals: [new iam.AnyPrincipal],
                resources: [`${domain.domainArn}/*`],
            })
        );

        const openSearchSync = new python.PythonFunction(this, 'DynamoDB2OpenSearch', {
            entry: './lambda/dynamo-db-to-open-search/',
            runtime: lambda.Runtime.PYTHON_3_8,
            index: 'main.py',
            handler: 'handler',
            environment: {
                OPENSEARCH_HOST: "https://" + domain.domainEndpoint,
            },
            timeout: cdk.Duration.minutes(1)
        });
        openSearchSync.addEventSource(new DynamoEventSource(table, {
            startingPosition: lambda.StartingPosition.LATEST,
        }));
        // TODO: Kibana requires to have backend user added to allow indexation of documents - manually!
        // fix this and make sure it's automatic
        domain.grantReadWrite( openSearchSync);
        domain.grantIndexReadWrite("*", openSearchSync);
        domain.grantIndexReadWrite("lambda-index", openSearchSync);

        // table.addGlobalSecondaryIndex({
        //     indexName: 'sessionId-index',
        //     partitionKey: {
        //         name: 'sessionId',
        //         type: dynamodb.AttributeType.STRING
        //     },
        //     sortKey: {
        //         name: 'connectionId',
        //         type: dynamodb.AttributeType.STRING
        //     }
        // })

        // define the connect handler
        const connectHandler = new lambdanodejs.NodejsFunction(this, "SQSConnectHandler", {
            entry: 'lambda/connectHandler.ts',
            environment: {
                TABLE_NAME: table.tableName,
            },
        });
        table.grantReadWriteData(connectHandler);

        // define the disconnect handler
        const disconnectHandler = new lambdanodejs.NodejsFunction(this, "SQSDisconnectHandler", {
            entry: 'lambda/diconnectHandler.ts',
            environment: {
                TABLE_NAME: table.tableName,
            },
        });
        table.grantReadWriteData(disconnectHandler);

        // // define the default handler
        // const defaultHandler = new lambda.Function(this, "DefaultHandler", {
        //     runtime: lambda.Runtime.NODEJS_12_X,
        //     handler: "default.handler",
        //     code: lambda.Code.fromAsset("lambda"),
        // });
        // define the send message handler
        const receiveHandler = new lambdanodejs.NodejsFunction(this, "SQSReceiveHandler", {
            entry: 'lambda/recieveHandler.ts',
            environment: {
                TABLE_NAME: table.tableName,
                QUEUE_URL: queue.queueUrl,
            },
        });
        table.grantReadWriteData(receiveHandler)
        queue.grantSendMessages(receiveHandler)

        // define the websocket API
        const webSocketApi = new apigatewayv2.WebSocketApi(this, "Tic", {
            connectRouteOptions: {
                integration: new apigatewayv2integrations.WebSocketLambdaIntegration('connect', connectHandler),
            },
            disconnectRouteOptions: {
                integration: new apigatewayv2integrations.WebSocketLambdaIntegration('disconnect', disconnectHandler),
            },
            defaultRouteOptions: {
                integration: new apigatewayv2integrations.WebSocketLambdaIntegration('default', receiveHandler),
            },
        });

        const apiStage = new apigatewayv2.WebSocketStage(this, 'DevStage', {
            webSocketApi,
            stageName: 'dev',
            autoDeploy: true,
        });

        const dockerImageAsset = new ecr_assets.DockerImageAsset(this, 'LiveSelectDocker', {
            directory: './fargate/live-select/',
            platform: ecr_assets.Platform.LINUX_ARM64,
            extraHash: '19',
            invalidation: {
                extraHash: true,
            }
        });

        const taskDefinition = new ecs.FargateTaskDefinition(this, 'LiveSelectTask', {
            cpu: 256,
            memoryLimitMiB: 512,
            runtimePlatform: {
                operatingSystemFamily: ecs.OperatingSystemFamily.LINUX,
                cpuArchitecture: ecs.CpuArchitecture.ARM64,
            },
        });

        taskDefinition.addContainer('LiveSelectContainer', {
            image: ecs.ContainerImage.fromDockerImageAsset(dockerImageAsset),
            entryPoint: ['./main'],
            healthCheck: {
                command: ['exit 0'],
                // command: ['CMD-SHELL', 'curl -f localhost/health || exit 1'],
                interval: cdk.Duration.seconds(10),
                retries: 3,
                timeout: cdk.Duration.seconds(2),
            },
            portMappings: [{
                containerPort: 80,
                hostPort: 80,
            }],
            logging: new ecs.AwsLogDriver({
                streamPrefix: 'LiveSelectContainer',
            }),
            environment: {
                TABLE_NAME: table.tableName,
                OPENSEARCH_HOST: "https://" + domain.domainEndpoint,
                KINESIS_STREAM_NAME: stream.streamName,
                DOMAIN_NAME: webSocketApi.apiEndpoint,
                STAGE_NAME: apiStage.stageName,
            }
        })

        const fargateService = new ecsp.ApplicationLoadBalancedFargateService(this, 'LiveSelectServer', {
            taskDefinition: taskDefinition,
            publicLoadBalancer: true,
        });
        queue.grantConsumeMessages(taskDefinition.taskRole)
        webSocketApi.grantManageConnections(taskDefinition.taskRole)
        table.grantReadWriteData(taskDefinition.taskRole)
        stream.grantReadWrite(taskDefinition.taskRole)


        const queueHandler = new golang.GoFunction(this, 'SQSQueueHandlerGo', {
            // entry: 'lambda/go-tic-reciver',
            entry: 'lambda/go-tic-game-handler',
            environment: {
                TABLE_NAME: table.tableName,
                OPENSEARCH_HOST: "https://" + domain.domainEndpoint,
                KINESIS_STREAM_NAME: stream.streamName,
                LIVE_SELECT_SERVER_ENDPOINT: "http://" + fargateService.loadBalancer.loadBalancerDnsName,
                DOMAIN_NAME: webSocketApi.apiEndpoint,
                STAGE_NAME: apiStage.stageName,
            },
        });
        queueHandler.addEventSource(new SqsEventSource(queue, {
            batchSize: 1,
            maxBatchingWindow: cdk.Duration.seconds(0),
        }))
        queue.grantConsumeMessages(queueHandler)
        webSocketApi.grantManageConnections(queueHandler)
        table.grantReadWriteData(queueHandler)


        const liveSelectPush = new golang.GoFunction(this, 'LiveSelectPushGo', {
            entry: 'lambda/dybamo-db-to-live-select',
            environment: {
                LIVE_SELECT_SERVER_ENDPOINT: "http://" + fargateService.loadBalancer.loadBalancerDnsName,
            },
        });
        liveSelectPush.addEventSource(new DynamoEventSource(table, {
            startingPosition: lambda.StartingPosition.LATEST,
        }));

        // define the websocket API endpoint
        new cdk.CfnOutput(this, "WebsocketSQSEndpoint", {
            value: webSocketApi.apiEndpoint,
        });
    }
}