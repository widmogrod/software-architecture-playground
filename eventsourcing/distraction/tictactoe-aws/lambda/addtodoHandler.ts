import {APIGatewayProxyEvent} from 'aws-lambda';

import * as AWS from 'aws-sdk';

const ddb = new AWS.DynamoDB.DocumentClient({apiVersion: '2012-08-10', region: process.env.AWS_REGION});

export const handler = async (event: APIGatewayProxyEvent) => {
    let connectionData;

    const tableName = process.env.TABLE_NAME;

    if (!tableName) {
        throw new Error('tableName not specified in process.env.TABLE_NAME');
    }

    if (!event.body) {
        throw new Error('event body is missing');
    }

    try {
        connectionData = await ddb.scan({TableName: tableName, ProjectionExpression: 'connectionId'}).promise();
    } catch (e) {
        return {statusCode: 500, body: e};
    }

    const apigwManagementApi = new AWS.ApiGatewayManagementApi({
        apiVersion: '2018-11-29',
        endpoint: event.requestContext.domainName + '/' + event.requestContext.stage,
    });

    const postData = JSON.parse(event.body).data;

    const postCalls = (connectionData.Items ?? []).map(async ({connectionId}) => {
        try {
            await apigwManagementApi.postToConnection({ConnectionId: connectionId, Data: postData}).promise();
        } catch (e: any) {
            if (e.statusCode === 410) {
                console.log(`Found stale connection, deleting ${connectionId}`);
                await ddb.delete({TableName: tableName, Key: {connectionId}}).promise();
            } else {
                throw e;
            }
        }
    });

    try {
        await Promise.all(postCalls);
    } catch (e) {
        return {statusCode: 500, body: e};
    }

    return {statusCode: 200, body: 'Data sent.'};
};