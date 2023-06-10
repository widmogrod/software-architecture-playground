import * as cdk from 'aws-cdk-lib';

import {Construct} from 'constructs';
import * as pipelines from 'aws-cdk-lib/pipelines';
import * as codebuild from 'aws-cdk-lib/aws-codebuild';


/**
 * The stack that defines the application pipeline
 */
export class ContinuousDeliveryStack extends cdk.Stack {
    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const pipeline = new pipelines.CodePipeline(this, 'Pipeline', {
            pipelineName: 'CleanVerticalPipeline',
            dockerEnabledForSynth: true,
            codeBuildDefaults: {
                cache: codebuild.Cache.local(codebuild.LocalCacheMode.DOCKER_LAYER)
            },
            synth: new pipelines.ShellStep('Synth', {
                input: pipelines.CodePipelineSource.gitHub(
                    "widmogrod/software-architecture-playground",
                    "master",
                    {
                        authentication: cdk.SecretValue.secretsManager('github-token'),
                    }
                ),

                commands: [
                    'cd clean-vertical/distraction/infrastructure/cdk-aws-dynamo; npm install',
                    'cd clean-vertical/distraction/infrastructure/cdk-aws-dynamo; npm run build',
                    'cd clean-vertical/distraction/infrastructure/cdk-aws-dynamo; npm run cdk synth',
                ],
                primaryOutputDirectory: 'clean-vertical/distraction/infrastructure/cdk-aws-dynamo/cdk.out',
            }),
        });

        // pipeline.addStage(new CleanVerticalStage(this, 'Prod', {
        //     env: { account: 'ACCOUNT2', region: 'us-west-2' }
        // }));
    }
}