#!/usr/bin/env node
import {App} from '@aws-cdk/core';
import {CleanVerticalPipelineStack} from '../lib/clean-vertical-pipeline-stack';
import {CleanVerticalRestStack} from "../lib/clean-vertical-rest-stack";
import {CleanVerticalHttpStack} from "../lib/clean-vertical-http-stack";
import {CleanVerticalRestOpenapiStack} from "../lib/clean-vertical-rest-openapi-stack";

const app = new App();

new CleanVerticalPipelineStack(app, 'CleanVerticalPipelineStack', {});

new CleanVerticalRestStack(app, 'DevCleanVerticalRestAPIStack', {
    env: {account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION},
})
new CleanVerticalHttpStack(app, 'DevCleanVerticalHttpAPIStack', {
    env: {account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION},
})
new CleanVerticalRestOpenapiStack(app, 'DevCleanVerticalRestOpenapiStack', {
    env: {account: process.env.CDK_DEFAULT_ACCOUNT, region: process.env.CDK_DEFAULT_REGION},
}).parse().then(() => {
    console.log('SYNTH')
    app.synth();
}).catch(e => {
    console.error(e)
    app.synth()
})

