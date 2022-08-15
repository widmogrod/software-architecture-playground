import {CfnOutput, Stage, StageProps} from "aws-cdk-lib";
import {Construct} from "constructs";
import {CleanVerticalHttpStack} from "./clean-vertical-http-stack";
import {CleanVerticalRestStack} from "./clean-vertical-rest-stack";

export class CleanVerticalStage extends Stage {
    public readonly httpApiUrl: CfnOutput;
    public readonly restApiUrl: CfnOutput;

    constructor(scope: Construct, id: string, props: StageProps) {
        super(scope, id, props);

        const httpApi = new CleanVerticalHttpStack(this, 'clean-vertica-http-api-stack', {})
        const restApi = new CleanVerticalRestStack(this, 'clean-vertical-rest-api-stack', {})


        this.httpApiUrl = httpApi.apiUrl
        this.restApiUrl = restApi.apiUrl
    }
}