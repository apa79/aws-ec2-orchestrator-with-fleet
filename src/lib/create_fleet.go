package lib

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func CreateFleet(svc *ec2.EC2) (fleetId *string, err error) {

	r := rand.New(rand.NewSource(time.Now().Unix()))

	var ltid *string

	defer func(e error) {
		if ltid != nil && e != nil {
			log.Println("clean up...")
			if _, err1 := svc.DeleteLaunchTemplate(&ec2.DeleteLaunchTemplateInput{LaunchTemplateId: ltid}); err1 != nil {
				log.Printf("unexpected error when deleting launch template, %s\n", err1)
			} else {
				log.Printf("launch template (%s) already deleted\n", aws.StringValue(ltid))
			}
		}
	}(err)

	now := time.Now()

	vpcId, err := DescribeDefaultVpc(svc)
	if err != nil {
		return nil, err
	}

	subnets, err := DescribeSubnets(svc, vpcId)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("order-%d-%d%d-%d", now.Day(), now.Hour(), now.Minute(), r.Intn(10000))
	ltid, err = createLaunchTemplate(svc, id)
	if err != nil {
		return nil, err
	}
	log.Printf("launch template created, %s\n", aws.StringValue(ltid))

	var overrides []*ec2.FleetLaunchTemplateOverridesRequest
	for _, s := range subnets {
		overrides = append(overrides,
			&ec2.FleetLaunchTemplateOverridesRequest{
				InstanceRequirements: &ec2.InstanceRequirementsRequest{
					AllowedInstanceTypes: []*string{aws.String("t2.medium"), aws.String("t2.small"), aws.String("t2.micro")},
					//ExcludedInstanceTypes:                 []*string{aws.String("r*"), aws.String("c*"), aws.String("m*")},
					//SpotMaxPricePercentageOverLowestPrice: aws.Int64(999999),
					CpuManufacturers: []*string{aws.String("intel")},
					MemoryMiB: &ec2.MemoryMiBRequest{
						Min: aws.Int64(0),
					},
					VCpuCount: &ec2.VCpuCountRangeRequest{
						Min: aws.Int64(0),
					},
				},
				//SubnetId: s,
			})
		if len(overrides) >= 1 {
			log.Printf("%s\n", aws.StringValue(s))
			break
		}
	}
	log.Printf("overrides: %d\n", len(overrides))

	resp, err := svc.CreateFleet(&ec2.CreateFleetInput{
		Type: aws.String("instant"),
		SpotOptions: &ec2.SpotOptionsRequest{
			AllocationStrategy: aws.String("price-capacity-optimized"),
		},
		TargetCapacitySpecification: &ec2.TargetCapacitySpecificationRequest{
			TotalTargetCapacity:       aws.Int64(4),
			OnDemandTargetCapacity:    aws.Int64(0),
			DefaultTargetCapacityType: aws.String("spot"),
			//TargetCapacityUnitType:    aws.String("vcpu"),
		},
		LaunchTemplateConfigs: []*ec2.FleetLaunchTemplateConfigRequest{
			{
				LaunchTemplateSpecification: &ec2.FleetLaunchTemplateSpecificationRequest{
					LaunchTemplateId: ltid,
					Version:          aws.String("$Latest"),
				},
				Overrides: overrides,
			},
		},
	})
	log.Printf("fleet output:\nid: %s\n", aws.StringValue(resp.FleetId))
	if len(resp.Errors) > 0 {
		for _, e := range resp.Errors {
			log.Printf("error: %s\n", e.String())
		}
	}

	fleetId = resp.FleetId
	return
}

func createLaunchTemplate(svc *ec2.EC2, id string) (*string, error) {
	imageId, err := DescribeImage(svc)
	if err != nil {
		return nil, err
	}

	var sgIDs []*string
	sgOutput, err := svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-name"),
				Values: []*string{aws.String("default")},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	for _, sg := range sgOutput.SecurityGroups {
		sgIDs = append(sgIDs, sg.GroupId)
	}

	input := &ec2.CreateLaunchTemplateInput{
		LaunchTemplateName: aws.String(id),
		LaunchTemplateData: &ec2.RequestLaunchTemplateData{
			ImageId:          imageId,
			InstanceType:     aws.String("t2.micro"),
			SecurityGroupIds: sgIDs,
			InstanceMarketOptions: &ec2.LaunchTemplateInstanceMarketOptionsRequest{
				MarketType: aws.String("spot"),
			},
		},
	}

	output, err := svc.CreateLaunchTemplate(input)
	if err != nil {
		return nil, err
	}

	return output.LaunchTemplate.LaunchTemplateId, nil
}

func DescribeDefaultVpc(svc ec2iface.EC2API) (*string, error) {
	output, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("is-default"),
				Values: []*string{aws.String("true")},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	for _, v := range output.Vpcs {
		return v.VpcId, nil
	}

	return nil, fmt.Errorf("cannot find default vpc")
}

func DescribeSubnets(svc ec2iface.EC2API, vpcId *string) ([]*string, error) {
	output, err := svc.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpcId},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	values := []*string{}
	for _, s := range output.Subnets {
		values = append(values, s.SubnetId)
	}
	return values, nil
}

func DescribeImage(svc ec2iface.EC2API) (*string, error) {
	output, err := svc.DescribeImages(&ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("architecture"),
				Values: []*string{aws.String("x86_64")},
			},
			{
				Name:   aws.String("is-public"),
				Values: []*string{aws.String("true")},
			},
			{
				Name:   aws.String("owner-alias"),
				Values: []*string{aws.String("amazon")},
			},
			{
				Name:   aws.String("name"),
				Values: []*string{aws.String("al2023-ami-2023.3.202402*")},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	for _, img := range output.Images {
		return img.ImageId, nil
	}

	return nil, fmt.Errorf("cannot find any amazon linux 2023 image")
}
