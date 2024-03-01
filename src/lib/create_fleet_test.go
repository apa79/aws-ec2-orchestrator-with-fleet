package lib

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/stretchr/testify/assert"
)

type mockedEC2 struct {
	ec2iface.EC2API
	DescribeSubnetsOutput ec2.DescribeSubnetsOutput
}

func (m mockedEC2) DescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	return &m.DescribeSubnetsOutput, nil
}

var (
	describeSubnetsOutputSingleResult = ec2.DescribeSubnetsOutput{
		Subnets: []*ec2.Subnet{
			{
				SubnetId: aws.String("subnet-c825a9f6"),
				VpcId:    aws.String("vpc-3ce97d46"),
			},
		},
	}

	describeSubnetsOutputMultiResults = ec2.DescribeSubnetsOutput{
		Subnets: []*ec2.Subnet{
			{
				SubnetId: aws.String("subnet-c825a9f6"),
				VpcId:    aws.String("vpc-3ce97d46"),
			},
			{
				SubnetId: aws.String("subnet-d6f9f19c"),
				VpcId:    aws.String("vpc-3ce97d46"),
			},
			{
				SubnetId: aws.String("subnet-d2b7f0fc"),
				VpcId:    aws.String("vpc-3ce97d46"),
			},
		},
	}
)

func TestDescribeSubnets(t *testing.T) {
	cases := []struct {
		Name     string
		Resp     ec2.DescribeSubnetsOutput
		Expected []string
	}{
		{
			Name:     "SingleSubnet",
			Resp:     describeSubnetsOutputSingleResult,
			Expected: []string{"subnet-c825a9f6"},
		},
		{
			Name:     "MultiSubnets",
			Resp:     describeSubnetsOutputMultiResults,
			Expected: []string{"subnet-c825a9f6", "subnet-d6f9f19c", "subnet-d2b7f0fc"},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			subnets, err := DescribeSubnets(
				mockedEC2{DescribeSubnetsOutput: c.Resp}, aws.String("vpc-3ce97d46"))

			var values []string
			for _, s := range subnets {
				values = append(values, aws.StringValue(s))
			}

			assert.Equal(t, c.Expected, values)

			assert.NoError(t, err)
		})
	}

}
