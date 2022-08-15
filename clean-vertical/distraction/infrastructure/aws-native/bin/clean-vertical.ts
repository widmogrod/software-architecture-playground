#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { CleanVerticalPipelineStack } from '../lib/clean-vertical-pipeline-stack';
import {CleanVerticalRestStack} from "../lib/clean-vertical-rest-stack";
import {CleanVerticalHttpStack} from "../lib/clean-vertical-http-stack";

const app = new cdk.App();

new CleanVerticalPipelineStack(app, 'CleanVerticalPipelineStack', {});

new CleanVerticalRestStack(app, 'DevCleanVerticalRestAPIStack', {
    env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
})
new CleanVerticalHttpStack(app, 'DevCleanVerticalHttpAPIStack', {
    env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
})

app.synth();