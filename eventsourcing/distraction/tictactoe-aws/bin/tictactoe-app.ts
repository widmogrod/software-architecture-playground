#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import {WebsocketSqSStack} from "../lib/websocket-sqs-stack";
import {StaticWebsiteStack} from "../lib/static-website-stack";

const app = new cdk.App();
new WebsocketSqSStack(app, 'WebsocketSqSStack', {});
new StaticWebsiteStack(app, 'StaticWebsiteStack', {});
