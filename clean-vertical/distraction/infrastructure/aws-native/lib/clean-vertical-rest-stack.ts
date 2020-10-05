import * as golang from 'aws-lambda-golang';
import * as apigateway from '@aws-cdk/aws-apigateway';
import {CfnOutput, Construct, Stack, StackProps} from "@aws-cdk/core";

export class CleanVerticalRestStack extends Stack {
    public readonly apiUrl: CfnOutput

    constructor(scope: Construct, id: string, props: StackProps) {
        super(scope, id, props);

        const helloLambda = new golang.GolangFunction(this, '../functions/hello', {});

        const restApi = new apigateway.RestApi(this, 'clean-vertical-rest-api');

        restApi.root.addResource('hello').addMethod('GET', new apigateway.LambdaIntegration(
            helloLambda,
            {

            }
        ))

        restApi.root.addResource('demo').addMethod('GET', new apigateway.MockIntegration({
            integrationResponses: [{
                statusCode: '200',
            }],
            passthroughBehavior: apigateway.PassthroughBehavior.NEVER,
            requestTemplates: {
                'application/json': '{ "statusCode": 200 }',
            },
        }), {
            methodResponses: [{statusCode: '200'}],
        });

        this.apiUrl = new CfnOutput(this, 'CleanVerticalRestAPIUrl', {
            value: restApi.url || '',
        });
    }
}