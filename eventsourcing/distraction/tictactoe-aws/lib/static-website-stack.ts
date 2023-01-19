import * as cdk from "aws-cdk-lib";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as s3d from "aws-cdk-lib/aws-s3-deployment";
import * as cf from "aws-cdk-lib/aws-cloudfront";
import * as cfo from "aws-cdk-lib/aws-cloudfront-origins";
import * as path from "path";

export class StaticWebsiteStack extends cdk.Stack {
    // constructor for the stack
    constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const bucket = new s3.Bucket(this, 'Bucket', {
            accessControl: s3.BucketAccessControl.PRIVATE,
        })

        const deployment = new s3d.BucketDeployment(this, 'BucketDeployment', {
            destinationBucket: bucket,
            sources: [s3d.Source.asset(path.resolve(__dirname, '../../tictactoe-app/build'))]
        })

        const originAccessIdentity = new cf.OriginAccessIdentity(this, 'OriginAccessIdentity');
        bucket.grantRead(originAccessIdentity);

       const distribution = new cf.Distribution(this, 'Distribution', {
            defaultRootObject: 'index.html',
            defaultBehavior: {
                origin: new cfo.S3Origin(bucket, {
                    originAccessIdentity: originAccessIdentity,
                }),
            },
        })

        new cdk.CfnOutput(this, "DistributionDomainName", {
            value: distribution.distributionDomainName,
        });
    }
}