import * as cdk from 'aws-cdk-lib';
import * as iam from 'aws-cdk-lib/aws-iam';
import { Construct } from 'constructs';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as ecsp from 'aws-cdk-lib/aws-ecs-patterns';
import * as ecr_assets from 'aws-cdk-lib/aws-ecr-assets';

export class FargateStack extends cdk.Stack {
    readonly loadBalancerDnsName: string;
    readonly taskRole: iam.IRole;

    constructor(scope: Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const dockerImageAsset = new ecr_assets.DockerImageAsset(this, 'LiveSelectDocker', {
            directory: './fargate/live-select/', // The path to the directory containing the Dockerfile
            platform: ecr_assets.Platform.LINUX_ARM64,
            extraHash: '4',
            invalidation: {
                extraHash: true,
            }
        });

        const taskDefinition = new ecs.FargateTaskDefinition(this, 'LiveSelectTask', {
            cpu: 256,
            memoryLimitMiB: 512,
            runtimePlatform: {
                operatingSystemFamily: ecs.OperatingSystemFamily.LINUX,
                cpuArchitecture: ecs.CpuArchitecture.ARM64,
            },
        });

        taskDefinition.addContainer('LiveSelectContainer', {
            image: ecs.ContainerImage.fromDockerImageAsset(dockerImageAsset),
            entryPoint: ['./main'],
            healthCheck: {
                command: ['CMD-SHELL', 'curl -f http://localhost:80/health || exit 1'],
            },
            portMappings: [{
                containerPort: 80,
                hostPort: 80,
            }],
            logging: new ecs.AwsLogDriver({
                streamPrefix: 'LiveSelectContainer',
            }),
        })

        const fargateService = new ecsp.ApplicationLoadBalancedFargateService(this, 'LiveSelectServer', {
            taskDefinition: taskDefinition,
            publicLoadBalancer: true,
        });

        this.loadBalancerDnsName = fargateService.loadBalancer.loadBalancerDnsName;
        this.taskRole = taskDefinition.taskRole;

        new cdk.CfnOutput(this, 'LiveSelectEndpoint', {
            value: fargateService.loadBalancer.loadBalancerDnsName,
            description: 'The DNS name of the load balancer for the Fargate service',
            exportName: 'ServiceEndpoint',
        });
    }
}