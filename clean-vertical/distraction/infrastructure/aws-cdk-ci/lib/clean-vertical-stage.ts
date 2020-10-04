import {CfnOutput, Construct, Stage, StageProps} from "@aws-cdk/core";
import {CleanVerticalStack} from "./clean-vertical-stack";

export class CleanVerticalStage extends Stage {
    public readonly apiUrl: CfnOutput;

    constructor(scope: Construct, id: string, props: StageProps) {
        super(scope, id, props);

        const service = new CleanVerticalStack(this, 'clean-vertica-stack', {})

        this.apiUrl = service.apiUrl
    }
}