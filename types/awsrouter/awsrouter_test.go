package awsrouter

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/go-cmp/cmp"
	//"github.com/google/go-cmp/cmp"
)

type TgwDescriberImpl struct{}

func (t TgwDescriberImpl) DescribeTransitGateways(ctx context.Context, params *ec2.DescribeTransitGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewaysOutput, error) {
	return &ec2.DescribeTransitGatewaysOutput{
		TransitGateways: []types.TransitGateway{
			{
				TransitGatewayId: aws.String("tgw-0d7f9b0c"),
				Tags: []types.Tag{
					{Key: aws.String("Name"),
						Value: aws.String("test")},
				},
				State:       "available",
				Description: aws.String("test"),
			},
		},
	}, nil
}

func (t TgwDescriberImpl) DescribeTransitGatewayRouteTables(ctx context.Context, params *ec2.DescribeTransitGatewayRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewayRouteTablesOutput, error) {
	return &ec2.DescribeTransitGatewayRouteTablesOutput{
		TransitGatewayRouteTables: []types.TransitGatewayRouteTable{
			{
				TransitGatewayRouteTableId: aws.String("rtb-0d7f9b0c"),
				Tags: []types.Tag{
					{Key: aws.String("Name"),
						Value: aws.String("test")},
				},
				State: "available",
			},
		},
	}, nil
}

func (t TgwDescriberImpl) SearchTransitGatewayRoutes(ctx context.Context, params *ec2.SearchTransitGatewayRoutesInput, optFns ...func(*ec2.Options)) (*ec2.SearchTransitGatewayRoutesOutput, error) {
	return &ec2.SearchTransitGatewayRoutesOutput{
		Routes: []types.TransitGatewayRoute{
			{
				DestinationCidrBlock: aws.String("10.0.1.0/24"),
				State:                "active",
				Type:                 "static",
			},
			{
				DestinationCidrBlock: aws.String("10.0.2.0/24"),
				State:                "active",
				Type:                 "static",
			},
		},
	}, nil

}

func TestGetTgw(t *testing.T) {
	type args struct {
		ctx   context.Context
		api   AwsRouter
		input *ec2.DescribeTransitGatewaysInput
	}
	tests := []struct {
		name    string
		args    args
		want    *ec2.DescribeTransitGatewaysOutput
		wantErr bool
	}{
		{"TestGetTgw1",
			args{context.TODO(),
				TgwDescriberImpl{},
				&ec2.DescribeTransitGatewaysInput{},
			},
			&ec2.DescribeTransitGatewaysOutput{
				TransitGateways: []types.TransitGateway{
					{
						TransitGatewayId: aws.String("tgw-0d7f9b0c"),
						Tags: []types.Tag{
							{Key: aws.String("Name"),
								Value: aws.String("test")},
						},
						State:       "available",
						Description: aws.String("test"),
					},
				}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTgw(tt.args.ctx, tt.args.api, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTgw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTgw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTgwInputFilter(t *testing.T) {
	type args struct {
		tgwIDs []string
	}
	tests := []struct {
		name string
		args args
		want *ec2.DescribeTransitGatewaysInput
	}{
		{
			"TestTgwInputFilter1",
			args{
				[]string{"tgw-0d7f9b0c", "tgw-0d7f9b0d"},
			},
			&ec2.DescribeTransitGatewaysInput{
				TransitGatewayIds: []string{"tgw-0d7f9b0c", "tgw-0d7f9b0d"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparer := cmp.Comparer(func(x, y *ec2.DescribeTransitGatewaysInput) bool {
				if cmp.Equal(x.TransitGatewayIds, y.TransitGatewayIds) {
					return true
				}
				return false
			})
			if dif := cmp.Diff(TgwInputFilter(tt.args.tgwIDs), tt.want, comparer); dif != "" {
				t.Errorf("TgwInputFilter() = %v, want %v", TgwInputFilter(tt.args.tgwIDs), tt.want)
			}
		})
	}
}

func TestTgwRouteTableInputFilter(t *testing.T) {
	type args struct {
		tgwIDs []string
	}
	tests := []struct {
		name string
		args args
		want *ec2.DescribeTransitGatewayRouteTablesInput
	}{
		{"simple",
			args{
				[]string{"tgw-0d7f9b0c", "tgw-0d7f9b0d"},
			},
			&ec2.DescribeTransitGatewayRouteTablesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("transit-gateway-id"),
						Values: []string{"tgw-0d7f9b0c", "tgw-0d7f9b0d"},
					},
				},
			},
		},
		{"empty",
			args{
				[]string{},
			},
			&ec2.DescribeTransitGatewayRouteTablesInput{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TgwRouteTableInputFilter(tt.args.tgwIDs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TgwRouteTableInputFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTgwRouteTables(t *testing.T) {
	type args struct {
		ctx   context.Context
		api   AwsRouter
		input *ec2.DescribeTransitGatewayRouteTablesInput
	}
	tests := []struct {
		name    string
		args    args
		want    *ec2.DescribeTransitGatewayRouteTablesOutput
		wantErr bool
	}{
		{
			"test1",
			args{
				context.TODO(),
				TgwDescriberImpl{},
				&ec2.DescribeTransitGatewayRouteTablesInput{},
			},
			&ec2.DescribeTransitGatewayRouteTablesOutput{
				TransitGatewayRouteTables: []types.TransitGatewayRouteTable{
					{
						TransitGatewayRouteTableId: aws.String("rtb-0d7f9b0c"),
						Tags: []types.Tag{
							{Key: aws.String("Name"),
								Value: aws.String("test")},
						},
						State: "available",
					}},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTgwRouteTables(tt.args.ctx, tt.args.api, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTgwRouteTables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTgwRouteTables() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTgwSearchRoutesInputFilter(t *testing.T) {
	type args struct {
		tgwRtID      string
		routeFilters []types.Filter
	}
	tests := []struct {
		name string
		args args
		want *ec2.SearchTransitGatewayRoutesInput
	}{
		{
			"empty",
			args{
				"rtb-0d7f9b0c",
				[]types.Filter{},
			},
			&ec2.SearchTransitGatewayRoutesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("state"),
						Values: []string{"active"},
					},
				},
				TransitGatewayRouteTableId: aws.String("rtb-0d7f9b0c"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TgwSearchRoutesInputFilter(tt.args.tgwRtID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TgwSearchRoutesInputFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
