import * as golang from 'aws-lambda-golang';
import * as apigateway from '@aws-cdk/aws-apigateway';
import {CfnOutput, Construct, Stack, StackProps} from "@aws-cdk/core";
import SwaggerParser = require("@apidevtools/swagger-parser");

export class CleanVerticalRestOpenapiStack extends Stack {
    public readonly apiUrl: CfnOutput
    private readonly restApi: apigateway.RestApi

    constructor(scope: Construct, id: string, props: StackProps) {
        super(scope, id, props);

        this.restApi = new apigateway.RestApi(this, 'clean-vertical-openrest-api');

        // this.apiUrl = new CfnOutput(this, 'CleanVerticalRestOpenAPIUrl', {
        //     value: this.restApi.urlForPath('/') || '',
        // });
    }

    public parse() {
        let api = SwaggerParser.parse('./openapi.yaml')
        let p = api.then((doc) => {
            for (let key in doc.paths) {
                let path = doc.paths[key];
                for (let method in path) {
                    let declaration = path[method]
                    if ('x-lambda-path' in declaration) {
                        let lambdaPath = `../${declaration['x-lambda-path']}`
                        let resourceName = key.replace(/\//, '')

                        this.restApi.root.addResource(resourceName).addMethod(method.toUpperCase(), new apigateway.LambdaIntegration(
                            new golang.GolangFunction(this, lambdaPath, {}),
                            {}
                        ))
                    }
                }
            }
        })
        return p;
    }
}