import * as cdk from "aws-cdk-lib";
import * as iam from "aws-cdk-lib/aws-iam";
// import * as lambda from "aws-cdk-lib/aws-lambda";
import * as lambdanodejs from "aws-cdk-lib/aws-lambda-nodejs";
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import * as apigatewayv2 from '@aws-cdk/aws-apigatewayv2-alpha';
import * as apigatewayv2integrations from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { SqsEventSource } from 'aws-cdk-lib/aws-lambda-event-sources';
export class WebsocketStack extends cdk.Stack {
    // constructor for the stack
    constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const queue  = new sqs.Queue(this, 'tic-tac-toe-queue', {
            queueName: 'tictactoe-queue',
            // visibilityTimeout: cdk.Duration.seconds(20),
            // receiveMessageWaitTime: cdk.Duration.seconds(5),
        })

        const table = new dynamodb.Table(this, 'WebsocketConnections', {
            partitionKey: {
                name: 'connectionId',
                type: dynamodb.AttributeType.STRING
            },
        });

        // define the connect handler
        const connectHandler = new lambdanodejs.NodejsFunction(this, "ConnectHandler", {
            entry: 'lambda/connectHandler.ts',
            environment: {
                TABLE_NAME: table.tableName,
            },
        });
        table.grantReadWriteData(connectHandler);

        // define the disconnect handler
        const disconnectHandler = new lambdanodejs.NodejsFunction(this, "DisconnectHandler", {
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
        const addtodoHandler = new lambdanodejs.NodejsFunction(this, "AddTodoHandler", {
            entry: 'lambda/addtodoHandler.ts',
            environment: {
                TABLE_NAME: table.tableName,
                QUEUE_URL: queue.queueUrl,
            },
        });
        table.grantReadWriteData(addtodoHandler)
        queue.grantSendMessages(addtodoHandler)

        // const helloLambda = new golang.GoFunction(this, 'hello-func-1', {
        //     entry: 'functions/hello',
        //     // bundling: {
        //     //     commandHooks: {
        //     //         beforeBundling(inputDir: string, outputDir: string): string[] {
        //     //             return ['go test ./' + inputDir];
        //     //         },
        //     //         afterBundling(inputDir: string, outputDir: string): string[] {
        //     //             return ['go build -o ' + outputDir + '/hello ' + inputDir];
        //     //         }
        //     //     }
        //     // }
        // });

        // define the websocket API
        const webSocketApi = new apigatewayv2.WebSocketApi(this, "Tic", {
            connectRouteOptions: {
                integration: new apigatewayv2integrations.WebSocketLambdaIntegration('connect', connectHandler),
            },
            disconnectRouteOptions: {
                integration: new apigatewayv2integrations.WebSocketLambdaIntegration('disconnect', disconnectHandler),
            },
            defaultRouteOptions: {
                integration: new apigatewayv2integrations.WebSocketLambdaIntegration('default', addtodoHandler),
            },
        });

        const apiStage = new apigatewayv2.WebSocketStage(this, 'DevStage', {
            webSocketApi,
            stageName: 'dev',
            autoDeploy: true,
        });

        webSocketApi.grantManageConnections(addtodoHandler)

        // const actionHandler = new lambdanodejs.NodejsFunction(this, "ActionHandler", {
        //     entry: 'lambda/actionHandler.ts',
        //     environment: {
        //         API_ID: webSocketApi.apiId,
        //         STAGE: apiStage.stageName,
        //     },
        // });
        // actionHandler.addEventSource(new SqsEventSource(queue, {
        //     batchSize: 1,
        //     maxBatchingWindow: cdk.Duration.seconds(0),
        // }))
        // queue.grantConsumeMessages(actionHandler)

        // define the websocket API endpoint
        new cdk.CfnOutput(this, "WebsocketEndpoint", {
            value: webSocketApi.apiEndpoint,
        });
    }
}