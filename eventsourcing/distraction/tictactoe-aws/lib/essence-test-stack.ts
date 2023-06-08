import * as cdk from "aws-cdk-lib";
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as opensearchservice from "aws-cdk-lib/aws-opensearchservice";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as kinesis from "aws-cdk-lib/aws-kinesis";


export class EssenceTestStack extends cdk.Stack {
    // constructor for the stack
    constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const stream = new kinesis.Stream(this, 'Stream', {
            streamName: 'test-record-stream',
            retentionPeriod: cdk.Duration.hours(24),
            streamMode: kinesis.StreamMode.ON_DEMAND,
            encryption: kinesis.StreamEncryption.UNENCRYPTED,
        });

        const table = new dynamodb.Table(this, 'WebsocketSQSConnections', {
            tableName: 'test-repo-record',
            partitionKey: {
                name: 'ID',
                type: dynamodb.AttributeType.STRING
            },
            sortKey: {
                name: 'Type',
                type: dynamodb.AttributeType.STRING
            },
            removalPolicy: cdk.RemovalPolicy.DESTROY, // not recommended for production code!
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            pointInTimeRecovery: false,
            // stream: dynamodb.StreamViewType.NEW_AND_OLD_IMAGES,
            // kinesisStream: stream,
        });

        const domain = new opensearchservice.Domain(this, 'DynamoDBProjection', {
            domainName: 'test-search',
            removalPolicy: cdk.RemovalPolicy.DESTROY,
            version: opensearchservice.EngineVersion.OPENSEARCH_1_3,
            // fineGrainedAccessControl: {
            //     masterUserName: 'admin',
            //     masterUserPassword: cdk.SecretValue.unsafePlainText('nile!DISLODGE5clause')
            // },
            capacity: {
                masterNodes: 0,
                dataNodes: 1,
                dataNodeInstanceType: ec2.InstanceType.of(
                    ec2.InstanceClass.T3,
                    ec2.InstanceSize.SMALL
                ).toString() + ".search",
            },
            // ebs: {
            //     volumeSize: 20,
            //     volumeType: ec2.EbsDeviceVolumeType.GENERAL_PURPOSE_SSD,
            // },
            zoneAwareness: {
                enabled: false,
            },
            // logging: {
            //     slowSearchLogEnabled: true,
            //     appLogEnabled: true,
            //     slowIndexLogEnabled: true,
            // },
            enforceHttps: true,
            // encryptionAtRest: {enabled: true},
            // nodeToNodeEncryption: true,
        });

        // define the websocket API endpoint
        new cdk.CfnOutput(this, "Kinesis", {
            value: stream.streamName,
        });
        new cdk.CfnOutput(this, "DynamoDB", {
            value: table.tableName,
        });
        new cdk.CfnOutput(this, "OpenSearch", {
            value: domain.domainEndpoint,
        });
    }
}