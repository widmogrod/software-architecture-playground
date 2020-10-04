import * as golang from 'aws-lambda-golang';
import * as apigateway from '@aws-cdk/aws-apigateway';
import {CfnOutput, Construct, Stack, StackProps} from "@aws-cdk/core";

export class CleanVerticalStack extends Stack {
    public readonly apiUrl: CfnOutput

    constructor(scope: Construct, id: string, props: StackProps) {
        super(scope, id, props);

        const backend = new golang.GolangFunction(this, '../../aws-native/functions/hello', {});
        const api = new apigateway.LambdaRestApi(this, 'clean-vertical-gateway', {
            description: 'Clean Vertical Gateway',
            handler: backend,
            proxy: false,
        });

        const items = api.root.addResource('hello');
        items.addMethod('GET');

        this.apiUrl = new CfnOutput(this, 'Url', {
            value: api.url,
        });
    }
}