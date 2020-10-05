import * as golang from 'aws-lambda-golang';
import * as apigatewayv2 from '@aws-cdk/aws-apigatewayv2';
import {CfnOutput, Construct, Stack, StackProps} from "@aws-cdk/core";

export class CleanVerticalHttpStack extends Stack {
    public readonly apiUrl: CfnOutput

    constructor(scope: Construct, id: string, props: StackProps) {
        super(scope, id, props);

        const helloLambda = new golang.GolangFunction(this, 'functions/hello', {});

        const httpApi = new apigatewayv2.HttpApi(this, 'clean-vertical-gateway-2');
        httpApi.addRoutes({
            path: '/hello',
            methods: [apigatewayv2.HttpMethod.GET],
            integration: new apigatewayv2.LambdaProxyIntegration({
                handler: helloLambda,
            }),
        });

        this.apiUrl = new CfnOutput(this, 'CleanVerticalAPIUrl', {
            value: httpApi.url || '',
        });
    }
}