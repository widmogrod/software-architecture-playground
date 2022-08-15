import {SecretValue, Stack, StackProps, Duration} from 'aws-cdk-lib';

import {Construct} from 'constructs';
import * as pipelines from 'aws-cdk-lib/pipelines';
import * as codebuild from 'aws-cdk-lib/aws-codebuild';
import {CleanVerticalStage} from "./clean-vertical-stage";

import * as targets from 'aws-cdk-lib/aws-events-targets';
import * as events from 'aws-cdk-lib/aws-events';

/**
 * The stack that defines the application pipeline
 */
export class CleanVerticalPipelineStack extends Stack {
    constructor(scope: Construct, id: string, props?: StackProps) {
        super(scope, id, props);

        const pipeline = new pipelines.CodePipeline(this, 'Pipeline', {
                pipelineName: 'CleanVerticalPipeline',
                dockerEnabledForSynth: true,
                // dockerEnabledForSelfMutation: true,
                codeBuildDefaults: {
                    cache: codebuild.Cache.local(codebuild.LocalCacheMode.DOCKER_LAYER)
                },
                synth: new pipelines.ShellStep('Synth', {
                    input: pipelines.CodePipelineSource.gitHub(
                        "widmogrod/software-architecture-playground",
                        "master",
                        {
                            authentication: SecretValue.secretsManager('github-token'),
                        }
                    ),

                    commands: [
                        'cd clean-vertical/distraction/infrastructure/aws-native; npm install',
                        'cd clean-vertical/distraction/infrastructure/aws-native; npm run build',
                        'cd clean-vertical/distraction/infrastructure/aws-native; npm run cdk synth',
                    ],
                    primaryOutputDirectory: 'clean-vertical/distraction/infrastructure/aws-native/cdk.out',
                }),
            })
        ;

        const preprod = new CleanVerticalStage(this, 'PreProd', {
            // env: { account: 'ACCOUNT1', region: 'us-east-2' }
        })

        pipeline.addStage(preprod, {
            post: [
                new pipelines.ShellStep('TestService', {
                    envFromCfnOutputs: {
                        'ENDPOINT_URL': preprod.restApiUrl,
                    },
                    commands: [
                        // Use 'curl' to GET the given URL and fail if it returns an error
                        'curl -Ssf $ENDPOINT_URL/hello?name=PreProd-test',
                    ],
                })
            ]
        })

        pipeline.addStage(new CleanVerticalStage(this, 'Prod', {
            // env: { account: 'ACCOUNT2', region: 'us-west-2' }
        }));

        // // kick off the pipeline every day
        // const rule = new events.Rule(this, 'Weekly', {
        //     schedule: events.Schedule.rate(Duration.days(7)),
        // });
        // rule.addTarget(new targets.CodePipeline(pipeline.pipeline));
    }
}