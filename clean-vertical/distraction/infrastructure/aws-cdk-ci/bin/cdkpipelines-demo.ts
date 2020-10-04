#!/usr/bin/env node
import { App } from '@aws-cdk/core';
import { CdkpipelinesDemoPipelineStack } from '../lib/cdkpipelines-demo-pipeline-stack';
import {CdkpipelinesDemoStage} from "../lib/cdkpipelines-demo-stage";
import {CleanVerticalRestStack} from "../lib/clean-vertical-rest-stack";
import {CleanVerticalHttpStack} from "../lib/clean-vertical-http-stack";

const app = new App();

new CdkpipelinesDemoPipelineStack(app, 'CdkpipelinesDemoPipelineStack', {});

new CdkpipelinesDemoStage(app, 'Dev', {
    env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
});
new CleanVerticalRestStack(app, 'DevCleanVerticalRestAPI', {
    env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
})
new CleanVerticalHttpStack(app, 'DevCleanVerticalHttpAPI', {
    env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
})

app.synth();