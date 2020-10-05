#!/usr/bin/env node
import { App } from '@aws-cdk/core';
import { CleanVerticalPipeline } from '../lib/clean-vertical-pipeline';
import {CleanVerticalRestStack} from "../lib/clean-vertical-rest-stack";
import {CleanVerticalHttpStack} from "../lib/clean-vertical-http-stack";

const app = new App();

new CleanVerticalPipeline(app, 'CleanVerticalPipeline', {});

new CleanVerticalRestStack(app, 'DevCleanVerticalRestAPI', {
    env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
})
new CleanVerticalHttpStack(app, 'DevCleanVerticalHttpAPI', {
    env: { account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION },
})

app.synth();