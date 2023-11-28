package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

// const (
// 	ins1 = "i-052551681090e99b3"
// 	ins2 = "i-009751b0ecd44df7b"
// 	varElb
// 	varElbTarget
// )

func main() {

	varIns1 := os.Getenv("Ins1")                 // webproxy 实例1
	varIns2 := os.Getenv("Ins2")                 // webproxy 实例2
	varElbBatchName := os.Getenv("ElbBatchName") // 负载均衡 和 目标组 前缀 如：b3-

	fmt.Printf("varIns1 : %s\n", varIns1)
	fmt.Printf("varIns2 : %s\n", varIns2)
	fmt.Printf("varElbBatchName : %s\n", varElbBatchName)

	if varIns1 == "" || varIns2 == "" || varElbBatchName == "" {
		log.Fatalf("值不能为空")
	}

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
		Name:       aws.String(varElbBatchName + "webproxyElb"), // 目标组名
		Port:       aws.Int64(80),                               // 指定端口
		Protocol:   aws.String("HTTP"),                          // 指定协议
		VpcId:      aws.String("vpc-0cadb665c480c21d1"),
		TargetType: aws.String("instance"),
	}

	tgOutput, err := svc.CreateTargetGroup(tgInput)
	if err != nil {
		log.Fatal("创建目标组失败:", err)
	}
	fmt.Printf("创建目标组成功: %s\n", *tgOutput.TargetGroups[0].TargetGroupArn)

	// 假设您已获取到了实例 ID
	instanceIDs := []string{varIns1, varIns2}

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
		Name:    aws.String(varElbBatchName + "webproxy_ElbTarget"),                                        //负载均衡名
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
