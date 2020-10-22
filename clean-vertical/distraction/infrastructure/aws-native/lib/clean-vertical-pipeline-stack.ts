import * as codepipeline from '@aws-cdk/aws-codepipeline';
import * as codepipeline_actions from '@aws-cdk/aws-codepipeline-actions';
import {Construct, Duration, SecretValue, Stack, StackProps} from '@aws-cdk/core';
import {CdkPipeline, ShellScriptAction, SimpleSynthAction} from '@aws-cdk/pipelines';
import {CleanVerticalStage} from "./clean-vertical-stage";
import * as targets from '@aws-cdk/aws-events-targets';
import * as events from '@aws-cdk/aws-events';

/**
 * The stack that defines the application pipeline
 */
export class CleanVerticalPipelineStack extends Stack {
    constructor(scope: Construct, id: string, props?: StackProps) {
        super(scope, id, props);

        const sourceArtifact = new codepipeline.Artifact();
        const cloudAssemblyArtifact = new codepipeline.Artifact();

        const pipeline = new CdkPipeline(this, 'Pipeline', {
            // The pipeline name
            pipelineName: 'CleanVerticalPipeline',
            cloudAssemblyArtifact,

            // Where the source can be found
            sourceAction: new codepipeline_actions.GitHubSourceAction({
                actionName: 'GitHub',
                output: sourceArtifact,
                oauthToken: SecretValue.secretsManager('github-token'),
                owner: 'widmogrod',
                repo: 'software-architecture-playground',
            }),

            // How it will be built and synthesized
            synthAction: SimpleSynthAction.standardNpmSynth({
                sourceArtifact,
                cloudAssemblyArtifact,

                // We need a build step to compile the TypeScript Lambda
                buildCommand: 'npm run build',
                subdirectory: 'clean-vertical/distraction/infrastructure/aws-native'
            }),
        });

        const preprod = new CleanVerticalStage(this, 'PreProd', {
            // env: { account: 'ACCOUNT1', region: 'us-east-2' }
        })

        const preprodStage = pipeline.addApplicationStage(preprod);
        preprodStage.addActions(new ShellScriptAction({
            actionName: 'TestService',
            useOutputs: {
                // Get the stack Output from the Stage and make it available in
                // the shell script as $ENDPOINT_URL.
                ENDPOINT_URL: pipeline.stackOutput(preprod.restApiUrl),
            },
            commands: [
                // Use 'curl' to GET the given URL and fail if it returns an error
                'curl -Ssf $ENDPOINT_URL/hello?name=PreProd-test',
            ],
        }));

        pipeline.addApplicationStage(new CleanVerticalStage(this, 'Prod', {
            // env: { account: 'ACCOUNT2', region: 'us-west-2' }
        }));

        // kick off the pipeline every day
        const rule = new events.Rule(this, 'Daily', {
            schedule: events.Schedule.rate(Duration.days(1)),
        });
        rule.addTarget(new targets.CodePipeline(pipeline.codePipeline));
    }
}