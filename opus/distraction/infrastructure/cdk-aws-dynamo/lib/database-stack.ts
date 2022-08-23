// cdk dynamo db stack
import * as cdk from 'aws-cdk-lib';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as kinesis from 'aws-cdk-lib/aws-kinesis';
import * as opensearchservice from 'aws-cdk-lib/aws-opensearchservice';
import {StreamViewType} from "aws-cdk-lib/aws-dynamodb/lib/table";

// new class for the stack
export class DatabaseStack extends cdk.Stack {
    // constructor for the stack
    constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
        super(scope, id, props);
        const stream = new kinesis.Stream(this, 'Stream', {
            streamName: 'opus-stream-dev',
            shardCount: 1,
            retentionPeriod: cdk.Duration.hours(48),
        });
        const table = new dynamodb.Table(this, 'Table', {
            tableName: 'opus-table-dev',
            partitionKey: {name: 'id', type: dynamodb.AttributeType.STRING},
            sortKey: {name: 'entity', type: dynamodb.AttributeType.STRING},
            billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
            tableClass: dynamodb.TableClass.STANDARD_INFREQUENT_ACCESS,
            kinesisStream: stream,
            stream: dynamodb.StreamViewType.NEW_AND_OLD_IMAGES
        });
        const domain = new opensearchservice.Domain(this, 'Domain', {
            domainName: 'opus-domain-dev',
            version: opensearchservice.EngineVersion.OPENSEARCH_1_3,
            fineGrainedAccessControl: {
                masterUserName: 'admin',
                masterUserPassword: cdk.SecretValue.unsafePlainText('nile!DISLODGE5clause')
            },
            capacity: {
                masterNodes: 3,
                dataNodes: 2,
                dataNodeInstanceType: ec2.InstanceType.of(
                    ec2.InstanceClass.T3,
                    ec2.InstanceSize.SMALL
                ).toString() + ".search",
                masterNodeInstanceType: ec2.InstanceType.of(
                    ec2.InstanceClass.T3,
                    ec2.InstanceSize.SMALL
                ).toString() + ".search",
            },
            ebs: {
                volumeSize: 20,
            },
            zoneAwareness: {
                availabilityZoneCount: 2,
            },
            logging: {
                slowSearchLogEnabled: true,
                appLogEnabled: true,
                slowIndexLogEnabled: true,
            },
            enforceHttps: true,
            encryptionAtRest: {enabled: true},
            nodeToNodeEncryption: true,
            // advancedOptions: {
            //     'rest.action.multi.allow_explicit_index': 'false',
            //     'indices.fielddata.cache.size': '25',
            //     'indices.query.bool.max_clause_count': '2048',
            // },
        });

        domain.addAccessPolicies(
            new iam.PolicyStatement({
                actions: ['es:*'],
                effect: iam.Effect.ALLOW,
                principals: [new iam.AnyPrincipal],
                resources: [domain.domainArn, `${domain.domainArn}/*`],
            })
        );
        // domain.addAccessPolicies(
        //     new iam.PolicyStatement({
        //         actions: ['es:ESHttpPost', 'es:ESHttpPut'],
        //         effect: iam.Effect.ALLOW,
        //         principals: [new iam.AccountPrincipal('123456789012')],
        //         resources: [domain.domainArn, `${domain.domainArn}/*`],
        //     }),
        //     new iam.PolicyStatement({
        //         actions: ['es:ESHttpGet'],
        //         effect: iam.Effect.ALLOW,
        //         principals: [new iam.AccountPrincipal('123456789012')],
        //         resources: [
        //             `${domain.domainArn}/_all/_settings`,
        //             `${domain.domainArn}/_cluster/stats`,
        //             `${domain.domainArn}/index-name*/_mapping/type-name`,
        //             `${domain.domainArn}/roletest*/_mapping/roletest`,
        //             `${domain.domainArn}/_nodes`,
        //             `${domain.domainArn}/_nodes/stats`,
        //             `${domain.domainArn}/_nodes/*/stats`,
        //             `${domain.domainArn}/_stats`,
        //             `${domain.domainArn}/index-name*/_stats`,
        //             `${domain.domainArn}/roletest*/_stat`,
        //         ],
        //     }),
        // );

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

