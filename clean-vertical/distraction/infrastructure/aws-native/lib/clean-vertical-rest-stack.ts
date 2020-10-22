import * as golang from 'aws-lambda-golang';
import * as apigateway from '@aws-cdk/aws-apigateway';
import * as codedeploy from '@aws-cdk/aws-codedeploy';
import * as cloudwatch from '@aws-cdk/aws-cloudwatch';
import * as lambda from '@aws-cdk/aws-lambda';
import {Tracing} from '@aws-cdk/aws-lambda';
import {CfnOutput, Construct, Stack, StackProps} from "@aws-cdk/core";
import * as iam from '@aws-cdk/aws-iam';

export class CleanVerticalRestStack extends Stack {
    public readonly apiUrl: CfnOutput

    constructor(scope: Construct, id: string, props: StackProps) {
        super(scope, id, props);

        const insightsExtensionLayer = this.createLambdaInsightsExtensionLayer();
        const insightsAppConfigLayer = this.createLambdaAppConfigExtensionLayer();

        const helloLambda = new golang.GolangFunction(this, '../functions/hello', {
            tracing: Tracing.ACTIVE,
            layers: [
                insightsExtensionLayer,
                insightsAppConfigLayer
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

        const testConfigLamnda = new lambda.Function(this, 'test-config-lambda-id', {
            runtime: lambda.Runtime.PYTHON_3_8,
            handler: 'index.handler',
            code: lambda.Code.fromAsset(__dirname + '/../functions/appconfig'),
            layers: [insightsAppConfigLayer],
            tracing: Tracing.ACTIVE,
        });
        testConfigLamnda.addToRolePolicy(new iam.PolicyStatement({
            effect: iam.Effect.ALLOW,
            resources: ["*"],
            actions: [
                "appconfig:GetConfiguration",
            ]
        }))

        restApi.root.addResource('config').addMethod('GET', new apigateway.LambdaIntegration(
            testConfigLamnda,
            {}
        ))

        const testConfigLamnda2 = new lambda.Function(this, 'test-config2-lambda-id', {
            runtime: lambda.Runtime.NODEJS_12_X,
            handler: 'index.handler',
            code: lambda.Code.fromAsset('functions/appconfig', {
                bundling: {
                    image: lambda.Runtime.NODEJS_12_X.bundlingDockerImage,
                    command: [
                        'bash', '-c', [
                            'ls -la .',
                            'npm install',
                            'cp -r /asset-input/* /asset-output/',
                        ].join('&&'),
                    ],
                    user: 'root',
                },
            }),
            layers: [insightsAppConfigLayer],
        });
        testConfigLamnda2.addToRolePolicy(new iam.PolicyStatement({
            effect: iam.Effect.ALLOW,
            resources: ["*"],
            actions: [
                "appconfig:GetConfiguration",
            ]
        }))

        restApi.root.addResource('config2').addMethod('GET', new apigateway.LambdaIntegration(
            testConfigLamnda2,
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
        return lambda.LayerVersion.fromLayerVersionArn(this, `LambdaInsightsExtensionLayerFromArn`, layerArn);
    }

    private createLambdaAppConfigExtensionLayer() {
        const layerArn = `arn:aws:lambda:eu-west-1:434848589818:layer:AWS-AppConfig-Extension:1`;
        return lambda.LayerVersion.fromLayerVersionArn(this, `AWS-AppConfig-ExtensionLayerFromArn`, layerArn);
    }

    // private config() {
    //     const app = new appconfig.CfnApplication(this, "cv-test-app-id", {
    //         name: "clean-vertical-config-application",
    //     })
    //
    //     const profile = new appconfig.CfnConfigurationProfile(this, "cv-test-config-profile-id", {
    //         applicationId: app.logicalId,
    //         name: "clean-vertical-config-profile",
    //         locationUri: "hosted",
    //     })
    //
    //     const devEnv = new appconfig.CfnEnvironment(this, 'cv-test-environment-id', {
    //         applicationId: app.logicalId,
    //         name: "cv-dev-env"
    //     })
    //
    //     const confVersion = new appconfig.CfnHostedConfigurationVersion(this, 'cv-test-conf-version-id', {
    //         applicationId: app.logicalId,
    //         configurationProfileId: profile.logicalId,
    //         content: JSON.stringify({
    //             stage: 'dev',
    //             chaos: false,
    //         }),
    //         contentType: "application/json",
    //     })
    //
    //     const deployStrategy = new appconfig.CfnDeploymentStrategy(this, 'cv-test-deplpoy-strategy', {
    //
    //     })
    //
    //     new appconfig.CfnDeployment(this, 'cv-test-deployment-id', {
    //         applicationId: app.logicalId,
    //         configurationProfileId: profile.logicalId,
    //         environmentId: devEnv.logicalId,
    //         configurationVersion: `${confVersion.latestVersionNumber}`,
    //         deploymentStrategyId: deployStrategy.logicalId,
    //     })
    // }
}