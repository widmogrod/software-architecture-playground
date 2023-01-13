import * as cdk from "aws-cdk-lib";
import * as lambdanodejs from "aws-cdk-lib/aws-lambda-nodejs";
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import * as apigatewayv2 from '@aws-cdk/aws-apigatewayv2-alpha';
import * as apigatewayv2integrations from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import {SqsEventSource} from 'aws-cdk-lib/aws-lambda-event-sources';
import * as golang from '@aws-cdk/aws-lambda-go-alpha';

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
                name: 'connectionId',
                type: dynamodb.AttributeType.STRING
            },
            // sortKey: {
            //     name: 'sessionId',
            //     type: dynamodb.AttributeType.STRING,
            // },
            removalPolicy: cdk.RemovalPolicy.DESTROY, // not recommended for production code!
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            pointInTimeRecovery: false,
        });
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
            entry: 'lambda/go-tic-reciver',
            environment: {
                TABLE_NAME: table.tableName,
            },
        });
        queueHandler.addEventSource(new SqsEventSource(queue, {
            batchSize: 1,
            maxBatchingWindow: cdk.Duration.seconds(0),
        }))
        queue.grantConsumeMessages(queueHandler)
        webSocketApi.grantManageConnections(queueHandler)


        // define the websocket API endpoint
        new cdk.CfnOutput(this, "WebsocketSQSEndpoint", {
            value: webSocketApi.apiEndpoint,
        });
    }
}