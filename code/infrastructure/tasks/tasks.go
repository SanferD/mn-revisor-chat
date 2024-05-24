package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

type ECSHelper struct {
	client                   *ecs.Client
	triggerCrawlerTaskDfnARN string
	triggerCrawlerClusterARN string
	timeout                  time.Duration
	subnetIDs                []string
	securityGroupIDs         []string
}

func InitializeECSHelper(ctx context.Context, triggerCrawlerTaskDfnARN, triggerCrawlerClusterARN string, subnetIDs, securityGroupIDs []string, timeout time.Duration, endpoint *string) (*ECSHelper, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error on loading default config: %v", err)
	}
	client := ecs.NewFromConfig(cfg, func(o *ecs.Options) {
		if endpoint != nil {
			o.BaseEndpoint = aws.String(*endpoint)
		}
	})
	return &ECSHelper{
		client:                   client,
		triggerCrawlerTaskDfnARN: triggerCrawlerTaskDfnARN,
		triggerCrawlerClusterARN: triggerCrawlerClusterARN,
		timeout:                  timeout,
		subnetIDs:                subnetIDs,
		securityGroupIDs:         securityGroupIDs,
	}, nil
}

func (ecsHelper *ECSHelper) InvokeTriggerCrawler(ctx context.Context) error {
	inp := ecs.RunTaskInput{
		Cluster:        aws.String(ecsHelper.triggerCrawlerClusterARN),
		TaskDefinition: aws.String(ecsHelper.triggerCrawlerTaskDfnARN),
		NetworkConfiguration: &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				Subnets:        ecsHelper.subnetIDs,
				SecurityGroups: ecsHelper.securityGroupIDs,
			},
		},
		LaunchType: types.LaunchTypeFargate,
	}
	out, err := ecsHelper.client.RunTask(ctx, &inp)
	if err != nil {
		return fmt.Errorf("error on run task: %v", err)
	}
	if len(out.Failures) > 0 {
		failure := out.Failures[0]
		return fmt.Errorf("failures encountered when running task: reason=%s , detail=%s", *failure.Reason, *failure.Detail)
	}
	return nil
}

func (ecsHelper *ECSHelper) IsTriggerCrawlerAlreadyRunning(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, ecsHelper.timeout)
	defer cancel()
	inp := ecs.ListTasksInput{
		Cluster:       aws.String(ecsHelper.triggerCrawlerClusterARN),
		DesiredStatus: types.DesiredStatusRunning,
	}
	out, err := ecsHelper.client.ListTasks(ctx, &inp)
	if err != nil {
		return false, fmt.Errorf("error on list tasks: %v", err)
	}
	hasTasks := len(out.TaskArns) > 0
	return hasTasks, nil
}
