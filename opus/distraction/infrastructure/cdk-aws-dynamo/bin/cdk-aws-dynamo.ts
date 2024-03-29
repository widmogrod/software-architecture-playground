#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import {ContinuousDeliveryStack} from "../lib/continuous-delivery-stack";
import {DatabaseStack} from "../lib/database-stack";

const app = new cdk.App();
// new ContinuousDeliveryStack(app, 'ContinuousDeliveryStack', {});
new DatabaseStack(app, 'DevDatabaseStack', {});