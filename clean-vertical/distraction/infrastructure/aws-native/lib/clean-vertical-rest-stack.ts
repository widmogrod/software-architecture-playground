import * as golang from 'aws-lambda-golang';
import * as apigateway from '@aws-cdk/aws-apigateway';
import * as codedeploy from '@aws-cdk/aws-codedeploy';
import * as cloudwatch from '@aws-cdk/aws-cloudwatch';
import * as lambda from '@aws-cdk/aws-lambda';
import {CfnOutput, Construct, Stack, StackProps} from "@aws-cdk/core";

export class CleanVerticalRestStack extends Stack {
    public readonly apiUrl: CfnOutput

    constructor(scope: Construct, id: string, props: StackProps) {
        super(scope, id, props);

        const helloLambda = new golang.GolangFunction(this, '../functions/hello', {});
        const version1Alias = new lambda.Alias(this, 'hello-lambda-alias', {
            aliasName: 'prod',
            version: helloLambda.latestVersion,
        });

        const application = new codedeploy.LambdaApplication(this, 'clean-vertical-rest-lambda-application', {
            applicationName: 'CleanVertical REST', // optional property
        });
        const deploymentGroup = new codedeploy.LambdaDeploymentGroup(this, 'BlueGreenDeployment', {
            application: application, // optional property: one will be created for you if not provided
            alias: version1Alias,
            deploymentConfig: codedeploy.LambdaDeploymentConfig.LINEAR_10PERCENT_EVERY_1MINUTE,
            alarms: [
                // pass some alarms when constructing the deployment group
                new cloudwatch.Alarm(this, 'CleanVerticalHelloErrors', {
                    comparisonOperator: cloudwatch.ComparisonOperator.GREATER_THAN_THRESHOLD,
                    threshold: 1,
                    evaluationPeriods: 1,
                    metric: version1Alias.metricErrors()
                })
            ]
        });

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