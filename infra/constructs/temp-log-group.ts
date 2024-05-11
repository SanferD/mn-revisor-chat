import * as cdk from 'aws-cdk-lib'
import * as logs from 'aws-cdk-lib/aws-logs'
import { Construct } from 'constructs';

export interface TempLogGroupProps extends logs.LogGroupProps {}

export class TempLogGroup extends logs.LogGroup {
    constructor(scope: Construct, id: string, props?: TempLogGroupProps) {
        const _props = {...props}
        _props.removalPolicy ??= cdk.RemovalPolicy.DESTROY
        _props.retention ??= logs.RetentionDays.TWO_WEEKS
        super(scope, id, _props)
    }
}
