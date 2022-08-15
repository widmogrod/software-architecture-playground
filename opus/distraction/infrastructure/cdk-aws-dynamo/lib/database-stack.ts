// cdk dynamo db stack
import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as kinesis from 'aws-cdk-lib/aws-kinesis';
import {StreamViewType} from "aws-cdk-lib/aws-dynamodb/lib/table";

// new class for the stack
export class DatabaseStack extends cdk.Stack {
    // constructor for the stack
    constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
        super(scope, id, props);
        const stream = new kinesis.Stream(this, 'Stream', {
            streamName: 'opus-stream',
            shardCount: 1,
            retentionPeriod: cdk.Duration.hours(48),
        });
        const table = new dynamodb.Table(this, 'Table', {
            tableName: 'opus-table',
            partitionKey: {name: 'id', type: dynamodb.AttributeType.STRING},
            sortKey: {name: 'entity', type: dynamodb.AttributeType.STRING},
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            tableClass: dynamodb.TableClass.STANDARD_INFREQUENT_ACCESS,
            kinesisStream: stream,
            stream: dynamodb.StreamViewType.NEW_AND_OLD_IMAGES
        });

        // const readScaling = table.autoScaleReadCapacity({ minCapacity: 1, maxCapacity: 50 });
        //
        // readScaling.scaleOnUtilization({
        //     targetUtilizationPercent: 50,
        // });
        //
        // readScaling.scaleOnSchedule('ScaleUpInTheMorning', {
        //     schedule: appscaling.Schedule.cron({ hour: '8', minute: '0' }),
        //     minCapacity: 20,
        // });
        //
        // readScaling.scaleOnSchedule('ScaleDownAtNight', {
        //     schedule: appscaling.Schedule.cron({ hour: '20', minute: '0' }),
        //     maxCapacity: 20,
        // });
    }
}

