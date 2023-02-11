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

export class WebsocketSqSStack extends cdk.Stack {
    // constructor for the stack
    constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const queue = new sqs.Queue(this, 'tic-tac-toe-queue-sqs', {
            queueName: 'tictactoe-sqs-queue',
            // visibilityTimeout: cdk.Duration.seconds(20),
            // receiveMessageWaitTime: cdk.Duration.seconds(5),
        })

        const table = new dynamodb.Table(this, 'WebsocketSQSConnections', {
            partitionKey: {
                name: 'key',
                type: dynamodb.AttributeType.STRING
            },
            removalPolicy: cdk.RemovalPolicy.DESTROY, // not recommended for production code!
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            pointInTimeRecovery: false,
            stream: dynamodb.StreamViewType.NEW_AND_OLD_IMAGES,

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
                masterNodes: 3,
                dataNodes: 2,
                dataNodeInstanceType: ec2.InstanceType.of(
                    ec2.InstanceClass.T3,
                    ec2.InstanceSize.SMALL
                ).toString() + ".search",
                masterNodeInstanceType: ec2.InstanceType.of(
                    ec2.InstanceClass.T3,
                    ec2.InstanceSize.SMALL
                ).toString() + ".search",
            },
            ebs: {
                volumeSize: 20,
            },
            zoneAwareness: {
                availabilityZoneCount: 2,
            },
            logging: {
                slowSearchLogEnabled: true,
                appLogEnabled: true,
                slowIndexLogEnabled: true,
            },
            enforceHttps: true,
            encryptionAtRest: {enabled: true},
            nodeToNodeEncryption: true,
            // advancedOptions: {
            //     'rest.action.multi.allow_explicit_index': 'false',
            //     'indices.fielddata.cache.size': '25',
            //     'indices.query.bool.max_clause_count': '2048',
            // },
        });

        domain.addAccessPolicies(
            new iam.PolicyStatement({
                actions: ['es:*'],
                effect: iam.Effect.ALLOW,
                principals: [new iam.AnyPrincipal],
                resources: [domain.domainArn, `${domain.domainArn}/*`],
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

        const queueHandler = new golang.GoFunction(this, 'SQSQueueHandlerGo', {
            // entry: 'lambda/go-tic-reciver',
            entry: 'lambda/go-tic-game-handler',
            environment: {
                TABLE_NAME: table.tableName,
                OPENSEARCH_HOST: "https://" + domain.domainEndpoint,
            },
        });
        queueHandler.addEventSource(new SqsEventSource(queue, {
            batchSize: 1,
            maxBatchingWindow: cdk.Duration.seconds(0),
        }))
        queue.grantConsumeMessages(queueHandler)
        webSocketApi.grantManageConnections(queueHandler)
        table.grantReadWriteData(queueHandler)


        // define the websocket API endpoint
        new cdk.CfnOutput(this, "WebsocketSQSEndpoint", {
            value: webSocketApi.apiEndpoint,
        });
    }
}