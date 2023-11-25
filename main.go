package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

const (
	ins1 = "i-052551681090e99b3"
	ins2 = "i-009751b0ecd44df7b"
)

func main() {
	// 初始化 AWS 会话
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-east-1"), // 替换为您的 AWS 区域
	})
	if err != nil {
		log.Fatal("创建会话失败:", err)
	}

	// 创建 ELBv2 服务客户端
	svc := elbv2.New(sess)

	// 创建目标组
	tgInput := &elbv2.CreateTargetGroupInput{
		Name:       aws.String("b3-testEdison"),
		Port:       aws.Int64(80),      // 指定端口
		Protocol:   aws.String("HTTP"), // 指定协议
		VpcId:      aws.String("vpc-0cadb665c480c21d1"),
		TargetType: aws.String("instance"),
	}

	tgOutput, err := svc.CreateTargetGroup(tgInput)
	if err != nil {
		log.Fatal("创建目标组失败:", err)
	}
	fmt.Printf("创建目标组成功: %s\n", *tgOutput.TargetGroups[0].TargetGroupArn)

	// 实例名称数组
	//instanceNames := []string{"b3-MyFirstInstanceTest1", "b3-MyFirstInstanceTest2"}

	// 创建 EC2 服务客户端
	//ec2svc := ec2.New(sess)

	// 获取实例 ID
	// 这里您需要根据实例名称获取实例 ID，具体实现根据您的需求来调整

	// 假设您已获取到了实例 ID
	instanceIDs := []string{ins1, ins2}

	// 向目标组注册实例
	for _, instanceID := range instanceIDs {
		regInput := &elbv2.RegisterTargetsInput{
			TargetGroupArn: tgOutput.TargetGroups[0].TargetGroupArn,
			Targets: []*elbv2.TargetDescription{
				{
					Id: aws.String(instanceID),
				},
			},
		}
		_, err = svc.RegisterTargets(regInput)
		if err != nil {
			log.Fatal("无法注册实例到目标组:", err)
		}
		fmt.Printf("实例 %s 已注册到目标组\n", instanceID)
	}

	// 创建 Application Load Balancer
	createLBOutput, err := svc.CreateLoadBalancer(&elbv2.CreateLoadBalancerInput{
		Name:    aws.String("TestElbEdison"),
		Subnets: []*string{aws.String("subnet-02b0af7335b197cb8"), aws.String("subnet-0a7e140afbc1f8f9b")}, // 替换为您的子网 ID
		SecurityGroups: []*string{
			aws.String("sg-033a6552e3ffe1a48"), // 安全组
		},
		Scheme: aws.String("internet-facing"),
		Type:   aws.String("application"),
	})
	if err != nil {
		log.Fatal("创建负载均衡器失败:", err)
	}
	fmt.Println("创建负载均衡器成功:", *createLBOutput.LoadBalancers[0].LoadBalancerArn)

	// 假设您已经有了目标组的 ARN
	targetGroupArn := *tgOutput.TargetGroups[0].TargetGroupArn

	// 创建侦听器
	_, err = svc.CreateListener(&elbv2.CreateListenerInput{
		DefaultActions: []*elbv2.Action{
			{
				Type:           aws.String("forward"),
				TargetGroupArn: aws.String(targetGroupArn),
			},
		},
		LoadBalancerArn: createLBOutput.LoadBalancers[0].LoadBalancerArn,
		Port:            aws.Int64(80),
		Protocol:        aws.String("HTTP"),
	})
	if err != nil {
		log.Fatal("创建侦听器失败:", err)
	}
	fmt.Println("创建侦听器成功")
}
