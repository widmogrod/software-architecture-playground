import {APIGatewayProxyEvent} from 'aws-lambda';

import * as AWS from 'aws-sdk';

const ddb = new AWS.DynamoDB.DocumentClient({apiVersion: '2012-08-10', region: process.env.AWS_REGION});

export const handler = async (event: APIGatewayProxyEvent) => {
    const queueurl = process.env.QUEUE_URL;
    const tableName = process.env.TABLE_NAME;

    if (!queueurl) {
        throw new Error('queueurl not specified in process.env.QUEUE_URL');
    }
    if (!tableName) {
        throw new Error('tableName not specified in process.env.TABLE_NAME');
    }

    if (!event.body) {
        throw new Error('event body is missing');
    }

    const sqs = new AWS.SQS({
        apiVersion: '2012-11-05',
        region: process.env.AWS_REGION,
    });

    try {
        await sqs.sendMessage({
            MessageBody: JSON.stringify({
                requestContext: event.requestContext,
                connectionId: event.requestContext.connectionId,
                body: event.body,
            }),
            QueueUrl: queueurl,
        }).promise()
    } catch (e) {
        console.log("error sending message to queue", e);
        return {statusCode: 500, body: e};
    }

    return {statusCode: 200, body: 'data sent to queue.'};
};