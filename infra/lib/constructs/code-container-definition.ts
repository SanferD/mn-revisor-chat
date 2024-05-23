import * as ecs from "aws-cdk-lib/aws-ecs";
import * as ecr_assets from "aws-cdk-lib/aws-ecr-assets";
import { Construct } from "constructs";
import { TempLogGroup } from "./temp-log-group";
import * as helpers from "../helpers";

export interface CodeContainerDefinitionProps {
  taskDefinition: ecs.TaskDefinition;
  environment: { [key: string]: string };
  streamPrefix?: string;
}

export class CodeContainerDefinition extends ecs.ContainerDefinition {
  constructor(scope: Construct, id: string, props: CodeContainerDefinitionProps) {
    const containerDefinitionId = `${id}-container-dfn`;
    const dockerImageAssetId = `${id}-container-asset`;
    const tempLogGroupId = `${id}-container-dfn-log-group`;
    const streamPrefix = props.streamPrefix ?? `${id}-`;

    helpers.doMakeBuildEcs(id);
    const imgAsset = new ecr_assets.DockerImageAsset(scope, dockerImageAssetId, {
      directory: helpers.getCodeDirPath(),
      buildArgs: {
        BINARY_PATH: helpers.getBuildAssetPathRelativeToCodeDir(id),
      },
      file: helpers.getCmdDockerfilePathRelativeToCodeDir(id),
    });

    const tempLogGroup = new TempLogGroup(scope, tempLogGroupId);

    super(scope, containerDefinitionId, {
      image: ecs.ContainerImage.fromDockerImageAsset(imgAsset),
      taskDefinition: props.taskDefinition,
      environment: props.environment,
      logging: new ecs.AwsLogDriver({
        logGroup: tempLogGroup,
        streamPrefix,
      }),
    });
  }
}
