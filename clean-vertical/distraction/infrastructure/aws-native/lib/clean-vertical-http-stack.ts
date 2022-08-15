import * as golang from '@aws-cdk/aws-lambda-go-alpha';
import * as apigatewayv2 from '@aws-cdk/aws-apigatewayv2-alpha';
import * as apigatewayv2integrations from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import {CfnOutput, Stack, StackProps} from "aws-cdk-lib";
import {Construct} from "constructs";

export class CleanVerticalHttpStack extends Stack {
    public readonly apiUrl: CfnOutput

    constructor(scope: Construct, id: string, props: StackProps) {
        super(scope, id, props);

        const helloLambda = new golang.GoFunction(this, 'hello-func-1', {
            entry: 'functions/hello',
            // bundling: {
            //     commandHooks: {
            //         beforeBundling(inputDir: string, outputDir: string): string[] {
            //             return ['go test ./' + inputDir];
            //         },
            //         afterBundling(inputDir: string, outputDir: string): string[] {
            //             return ['go build -o ' + outputDir + '/hello ' + inputDir];
            //         }
            //     }
            // }
        });

        const httpApi = new apigatewayv2.HttpApi(this, 'clean-vertical-gateway-2');
        httpApi.addRoutes({
            path: '/hello',
            methods: [apigatewayv2.HttpMethod.GET],
            integration: new apigatewayv2integrations.HttpLambdaIntegration('http-go-hello-lambda', helloLambda),
        });

        this.apiUrl = new CfnOutput(this, 'CleanVerticalAPIUrl', {
            value: httpApi.url || '',
        });
    }
}