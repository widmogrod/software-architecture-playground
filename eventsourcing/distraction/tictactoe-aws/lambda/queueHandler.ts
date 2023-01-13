import {SQSHandler, SQSEvent} from 'aws-lambda';

import * as AWS from 'aws-sdk';

const ddb = new AWS.DynamoDB.DocumentClient({apiVersion: '2012-08-10', region: process.env.AWS_REGION});

export const handler: SQSHandler = async (event: SQSEvent) => {
    const tableName = process.env.TABLE_NAME;

    if (!tableName) {
        throw new Error('tableName not specified in process.env.TABLE_NAME');
    }

    for (const record of event.Records) {
        const body = JSON.parse(record.body);
        const connectionId = body.connectionId;
        const postData = body.body;

        const apigwManagementApi = new AWS.ApiGatewayManagementApi({
            apiVersion: '2018-11-29',
            endpoint: body.requestContext.domainName + '/' + body.requestContext.stage,
        });

        try {
            await apigwManagementApi.postToConnection({ConnectionId: connectionId, Data: postData}).promise();
        } catch (e: any) {
            if (e.statusCode === 410) {
                console.log(`Found stale connection, deleting ${connectionId}`);
                await ddb.delete({
                    TableName: tableName,
                    Key: {connectionId}
                }).promise();
            } else {
                throw e;
            }
        }
    }
};