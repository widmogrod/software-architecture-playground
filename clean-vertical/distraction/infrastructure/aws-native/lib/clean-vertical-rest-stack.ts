import * as golang from 'aws-lambda-golang';
import * as apigateway from '@aws-cdk/aws-apigateway';
import * as codedeploy from '@aws-cdk/aws-codedeploy';
import * as cloudwatch from '@aws-cdk/aws-cloudwatch';
import * as lambda from '@aws-cdk/aws-lambda';
import {Tracing} from '@aws-cdk/aws-lambda';
import {CfnOutput, Construct, Stack, StackProps} from "@aws-cdk/core";

export class CleanVerticalRestStack extends Stack {
    public readonly apiUrl: CfnOutput

    constructor(scope: Construct, id: string, props: StackProps) {
        super(scope, id, props);

        const insightsExtensionLayer = this.createLambdaInsightsExtensionLayer();

        const helloLambda = new golang.GolangFunction(this, '../functions/hello', {
            tracing: Tracing.ACTIVE,
            profiling: true,
            layers: [
                insightsExtensionLayer,
            ],
            deadLetterQueueEnabled: true,
        });

        const helloLambdaLiveVersion = helloLambda.currentVersion.addAlias('live2')

        new codedeploy.LambdaDeploymentGroup(this, 'rest-api-deployment-group', {
            alias: helloLambdaLiveVersion,
            deploymentConfig: codedeploy.LambdaDeploymentConfig.CANARY_10PERCENT_5MINUTES,
            alarms: [
                // pass some alarms when constructing the deployment group
                new cloudwatch.Alarm(this, 'CleanVerticalHelloErrors', {
                    comparisonOperator: cloudwatch.ComparisonOperator.GREATER_THAN_THRESHOLD,
                    threshold: 1,
                    evaluationPeriods: 1,
                    metric: helloLambdaLiveVersion.metricErrors()
                })
            ]
        });

        const restApi = new apigateway.RestApi(this, 'clean-vertical-rest-api');

        restApi.root.addResource('hello').addMethod('GET', new apigateway.LambdaIntegration(
            helloLambda,
            {}
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

    private createLambdaInsightsExtensionLayer() {
        const region = 'eu-west-1'
        const layerArn = `arn:aws:lambda:${region}:580247275435:layer:LambdaInsightsExtension:2`;
        return lambda.LayerVersion.fromLayerVersionArn(this, `LayerFromArn`, layerArn);
    }
}