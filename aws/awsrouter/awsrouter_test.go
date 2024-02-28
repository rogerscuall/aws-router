package awsrouter

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/rogerscuall/aws-router/ports"
)

type TgwDescriberImpl struct{}

// listDescribeTransitGatewaysOutput is a mock of DescribeTransitGatewaysOutput
// there are multiple TransitGateways in this mock
var listDescribeTransitGatewaysOutput *ec2.DescribeTransitGatewaysOutput = &ec2.DescribeTransitGatewaysOutput{
	TransitGateways: []types.TransitGateway{
		{
			TransitGatewayId: aws.String("tgw-0d7f9b0a"),
			Tags: []types.Tag{
				{Key: aws.String("Name"),
					Value: aws.String("testA")},
			},
			State:       "available",
			Description: aws.String("testA"),
		},
		{
			TransitGatewayId: aws.String("tgw-0d7f9b0b"),
			Tags: []types.Tag{
				{Key: aws.String("Name"),
					Value: aws.String("testB")},
			},
			State:       "available",
			Description: aws.String("testB"),
		},
		{
			TransitGatewayId: aws.String("tgw-0d7f9b0c"),
			Tags: []types.Tag{
				{Key: aws.String("Name"),
					Value: aws.String("testC")},
			},
			State:       "available",
			Description: aws.String("testC"),
		},
	},
}

var listTgw []*Tgw = []*Tgw{
	{
		ID:   "tgw-0d7f9b0a",
		Name: "testA",
		Data: listDescribeTransitGatewaysOutput.TransitGateways[0],
	},
	{
		ID:   "tgw-0d7f9b0b",
		Name: "testB",
		Data: listDescribeTransitGatewaysOutput.TransitGateways[1],
	},
	{
		ID:   "tgw-0d7f9b0c",
		Name: "testC",
		Data: listDescribeTransitGatewaysOutput.TransitGateways[2],
	},
}

var listDescribeTransitGatewayRouteTablesOutput *ec2.DescribeTransitGatewayRouteTablesOutput = &ec2.DescribeTransitGatewayRouteTablesOutput{
	TransitGatewayRouteTables: []types.TransitGatewayRouteTable{
		{
			TransitGatewayRouteTableId: aws.String("rtb-0d7f9b0a"),
			Tags: []types.Tag{
				{Key: aws.String("Name"),
					Value: aws.String("testA")},
			},
			State:            "available",
			TransitGatewayId: aws.String("tgw-0d7f9b0a"),
		},
		{
			TransitGatewayRouteTableId: aws.String("rtb-0d7f9b0b"),
			Tags: []types.Tag{
				{Key: aws.String("Name"),
					Value: aws.String("testB")},
			},
			State:            "available",
			TransitGatewayId: aws.String("tgw-0d7f9b0a"),
		},
		{
			TransitGatewayRouteTableId: aws.String("rtb-0d7f9b0c"),
			Tags: []types.Tag{
				{Key: aws.String("Name"),
					Value: aws.String("testC")},
			},
			State:            "available",
			TransitGatewayId: aws.String("tgw-0d7f9b0b"),
		},
	},
}

var listSearchTransitGatewayRoutesOutput *ec2.SearchTransitGatewayRoutesOutput = &ec2.SearchTransitGatewayRoutesOutput{
	Routes: []types.TransitGatewayRoute{
		{
			DestinationCidrBlock:      aws.String("10.0.0.0/16"),
			State:                     "active",
			Type:                      "static",
			TransitGatewayAttachments: []types.TransitGatewayRouteAttachment{},
		},
		{
			DestinationCidrBlock:      aws.String("10.0.1.0/24"),
			State:                     "blackhole",
			Type:                      "static",
			TransitGatewayAttachments: []types.TransitGatewayRouteAttachment{},
		},
		{
			DestinationCidrBlock:      aws.String("10.0.2.0/24"),
			State:                     "active",
			Type:                      "static",
			TransitGatewayAttachments: []types.TransitGatewayRouteAttachment{},
		},
	},
}

var listGetTransitGatewayRouteTableAssociationsOutput *ec2.GetTransitGatewayRouteTableAssociationsOutput = &ec2.GetTransitGatewayRouteTableAssociationsOutput{
	Associations: []types.TransitGatewayRouteTableAssociation{
		{
			ResourceId:                 aws.String("vpc-0af25be733475a425"),
			ResourceType:               "vpc",
			TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec95f"),
		},
		{
			ResourceId:                 aws.String("tgw-04408890ef44df3e3"),
			ResourceType:               "peering",
			TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec96f"),
		},
		{
			ResourceId:                 aws.String("tgw-attach-09db78f3e74abf792"),
			ResourceType:               "connect",
			TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec97f"),
		},
		{
			ResourceId:                 aws.String("3c1a5494-3491-481d-b82d-7e2c61204f3f"),
			ResourceType:               "direct-connect-gateway",
			TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec99f"),
		},
	},
}

var listTgwAttachments []types.TransitGatewayRouteAttachment = []types.TransitGatewayRouteAttachment{
	{
		ResourceId:                 aws.String("vpc-0af25be733475a425"),
		ResourceType:               "vpc",
		TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec95f"),
	},
	{
		ResourceId:                 aws.String("tgw-04408890ef44df3e3"),
		ResourceType:               "peering",
		TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec96f"),
	},
	{
		ResourceId:                 aws.String("tgw-attach-09db78f3e74abf792"),
		ResourceType:               "connect",
		TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec97f"),
	},
	{
		ResourceId:                 aws.String("3c1a5494-3491-481d-b82d-7e2c61204f3f"),
		ResourceType:               "direct-connect-gateway",
		TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec99f"),
	},
}

// var listTgwAttachmentAssociations []types.TransitGatewayRouteTableAssociation = []types.TransitGatewayRouteTableAssociation{
// 	{
// 		ResourceId:                 aws.String("vpc-0af25be733475a425"),
// 		ResourceType:               "vpc",
// 		TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec95f"),
// 	},
// 	{
// 		ResourceId:                 aws.String("tgw-04408890ef44df3e3"),
// 		ResourceType:               "peering",
// 		TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec96f"),
// 	},
// 	{
// 		ResourceId:                 aws.String("tgw-attach-09db78f3e74abf792"),
// 		ResourceType:               "connect",
// 		TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec97f"),
// 	},
// 	{
// 		ResourceId:                 aws.String("3c1a5494-3491-481d-b82d-7e2c61204f3f"),
// 		ResourceType:               "direct-connect-gateway",
// 		TransitGatewayAttachmentId: aws.String("tgw-attach-080f3014bd52ec99f"),
// 	},
// }

// DescribeTransitGateways is a mock of DescribeTransitGateways
// it uses listDescribeTransitGatewaysOutput to return a list of TransitGateways
// depending on the filters in params, it will return one ore more TransitGateways
func (t TgwDescriberImpl) DescribeTransitGateways(ctx context.Context, params *ec2.DescribeTransitGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewaysOutput, error) {
	// if TransitGatewayIds is empty, return all TransitGateways
	if len(params.TransitGatewayIds) == 0 {
		return listDescribeTransitGatewaysOutput, nil
	}
	// if TransitGatewayIds is not empty, return only the TransitGateways that are in TransitGatewayIds
	var tgws []types.TransitGateway
	for _, tgw := range listDescribeTransitGatewaysOutput.TransitGateways {
		for _, tgwID := range params.TransitGatewayIds {
			if *tgw.TransitGatewayId == tgwID {
				tgws = append(tgws, tgw)
			}
		}
	}
	return &ec2.DescribeTransitGatewaysOutput{
		TransitGateways: tgws,
	}, nil
}

func (t TgwDescriberImpl) DescribeTransitGatewayRouteTables(ctx context.Context, params *ec2.DescribeTransitGatewayRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewayRouteTablesOutput, error) {
	filter := params.Filters
	// if the filter is empty, return all TransitGatewayRouteTables
	if len(filter) == 0 {
		return listDescribeTransitGatewayRouteTablesOutput, nil
	}
	// if the filter is not empty, return only the TransitGatewayRouteTables that are in the filter
	var tgwrtbs []types.TransitGatewayRouteTable
	for _, tgwrtb := range listDescribeTransitGatewayRouteTablesOutput.TransitGatewayRouteTables {
		for _, f := range filter {
			if *f.Name == "transit-gateway-id" {
				for _, tgwID := range f.Values {
					if *tgwrtb.TransitGatewayId == tgwID {
						tgwrtbs = append(tgwrtbs, tgwrtb)
					}
				}
			}
		}
	}
	return &ec2.DescribeTransitGatewayRouteTablesOutput{
		TransitGatewayRouteTables: tgwrtbs,
	}, nil
}

func (t TgwDescriberImpl) SearchTransitGatewayRoutes(ctx context.Context, params *ec2.SearchTransitGatewayRoutesInput, optFns ...func(*ec2.Options)) (*ec2.SearchTransitGatewayRoutesOutput, error) {
	// if the filter is empty, return all TransitGatewayRoutes
	filters := params.Filters
	if len(filters) == 0 {
		return listSearchTransitGatewayRoutesOutput, nil
	}
	// if the filter is not empty, return only the TransitGatewayRoutes that are in the filter
	var tgwrts []types.TransitGatewayRoute
	for _, tgwrt := range listSearchTransitGatewayRoutesOutput.Routes {
		for _, f := range filters {
			if *f.Name == "state" {
				for _, state := range f.Values {
					if fmt.Sprint(tgwrt.State) == state {
						tgwrts = append(tgwrts, tgwrt)
					}
				}
			}
		}
	}
	return &ec2.SearchTransitGatewayRoutesOutput{
		Routes: tgwrts,
	}, nil
}

func (t TgwDescriberImpl) GetTransitGatewayRouteTableAssociations(ctx context.Context, params *ec2.GetTransitGatewayRouteTableAssociationsInput, optFns ...func(*ec2.Options)) (*ec2.GetTransitGatewayRouteTableAssociationsOutput, error) {
	return listGetTransitGatewayRouteTableAssociationsOutput, nil
}

func (t TgwDescriberImpl) DescribeTransitGatewayAttachments(ctx context.Context, params *ec2.DescribeTransitGatewayAttachmentsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewayAttachmentsOutput, error) {
	return nil, nil
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
				return cmp.Equal(x.TransitGatewayIds, y.TransitGatewayIds)
			})
			if dif := cmp.Diff(ports.TgwInputFilter(tt.args.tgwIDs), tt.want, comparer); dif != "" {
				t.Errorf("TgwInputFilter() = %v, want %v", ports.TgwInputFilter(tt.args.tgwIDs), tt.want)
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
			if got := ports.TgwRouteTableInputFilter(tt.args.tgwIDs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TgwRouteTableInputFilter() = %v, want %v", got, tt.want)
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
					{
						Name:   aws.String("state"),
						Values: []string{"blackhole"},
					},
				},
				TransitGatewayRouteTableId: aws.String("rtb-0d7f9b0c"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ports.TgwSearchRoutesInputFilter(tt.args.tgwRtID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TgwSearchRoutesInputFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateRouting(t *testing.T) {
	type args struct {
		ctx context.Context
		api ports.AWSRouter
	}
	tests := []struct {
		name    string
		args    args
		want    []*Tgw
		wantErr bool
	}{
		//TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateRouting(tt.args.ctx, tt.args.api)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRouting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateRouting() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newTgw(t *testing.T) {
	type args struct {
		tgw types.TransitGateway
	}
	tests := []struct {
		name string
		args args
		want *Tgw
	}{
		{
			"simple",
			args{
				listDescribeTransitGatewaysOutput.TransitGateways[0],
			},
			&Tgw{
				ID:   "tgw-0d7f9b0a",
				Name: "testA",
				Data: listDescribeTransitGatewaysOutput.TransitGateways[0],
			},
		},
		{
			"noName",
			args{
				types.TransitGateway{
					TransitGatewayId: aws.String("tgw-0d7f9b0a"),
				},
			},
			&Tgw{
				ID:   "tgw-0d7f9b0a",
				Name: "tgw-0d7f9b0a",
				Data: types.TransitGateway{TransitGatewayId: aws.String("tgw-0d7f9b0a")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTgw(tt.args.tgw); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTgw() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test function GetAllTgws
func TestGetAllTgws(t *testing.T) {
	type args struct {
		ctx context.Context
		api ports.AWSRouter
	}
	tests := []struct {
		name    string
		args    args
		want    []*Tgw
		wantErr bool
	}{
		{
			"simple",
			args{
				context.TODO(),
				TgwDescriberImpl{},
			},
			listTgw,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAllTgws(tt.args.ctx, tt.args.api)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllTgws() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllTgws() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newTgwAttachments(t *testing.T) {
	type args struct {
		att types.TransitGatewayRouteAttachment
	}
	tests := []struct {
		name string
		args args
		want *TgwAttachment
	}{
		{
			"vpc",
			args{
				listTgwAttachments[0],
			},
			&TgwAttachment{
				ID:         "tgw-attach-080f3014bd52ec95f",
				ResourceID: "vpc-0af25be733475a425",
				Type:       "vpc",
			},
		},
		{
			"tgw",
			args{
				listTgwAttachments[1],
			},
			&TgwAttachment{
				ResourceID: "tgw-04408890ef44df3e3",
				Type:       "peering",
				ID:         "tgw-attach-080f3014bd52ec96f",
			},
		},
		{
			"connect",
			args{
				listTgwAttachments[2],
			},
			&TgwAttachment{
				ResourceID: "tgw-attach-09db78f3e74abf792",
				Type:       "connect",
				ID:         "tgw-attach-080f3014bd52ec97f",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newTgwAttachment(tt.args.att); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTgw() = %v, want %v", got, tt.want)
			}
		})
	}
}
