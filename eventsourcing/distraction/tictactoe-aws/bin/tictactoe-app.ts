#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import {ContinuousDeliveryStack} from "../lib/continuous-delivery-stack";
import {DatabaseStack} from "../lib/database-stack";
import {WebsocketStack} from "../lib/websocket-stack";
import {WebsocketSqSStack} from "../lib/websocket-sqs-stack";

const app = new cdk.App();
// new ContinuousDeliveryStack(app, 'ContinuousDeliveryStack', {});
// new DatabaseStack(app, 'DevDatabaseStack', {});
// new WebsocketStack(app, 'WebsocketStack', {});
new WebsocketSqSStack(app, 'WebsocketSqSStack', {});