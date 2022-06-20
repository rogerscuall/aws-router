package awsrouter

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// AwsRouter is an interface with the methods needed for routing.
type AwsRouter interface {
	DescribeTransitGateways(ctx context.Context, params *ec2.DescribeTransitGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewaysOutput, error)
	DescribeTransitGatewayRouteTables(ctx context.Context, params *ec2.DescribeTransitGatewayRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewayRouteTablesOutput, error)
	SearchTransitGatewayRoutes(ctx context.Context, params *ec2.SearchTransitGatewayRoutesInput, optFns ...func(*ec2.Options)) (*ec2.SearchTransitGatewayRoutesOutput, error)
}

// TgwRouteTable holds the Route Table ID, a list of routes and other RouteTable info.
// Represents a Route Table of a Transit Gateway in AWS.
type TgwRouteTable struct {
	ID     string
	Name   string
	Data   types.TransitGatewayRouteTable
	Routes []types.TransitGatewayRoute
}

// Bytes returns the JSON representation of the TgwRouteTable as a slice of bytes.
func (t *TgwRouteTable) Bytes() []byte {
	b, _ := json.Marshal(t)
	return b
}

// newTgwRouteTable creates a TgwRouteTable from an AWS TGW Route Table.
func newTgwRouteTable(t types.TransitGatewayRouteTable) *TgwRouteTable {
	// rt is the TgwRouteTable
	rt := &TgwRouteTable{}

	// if tgwRouteTable has no id return the empty TgwRouteTable struct.
	if t.TransitGatewayRouteTableId == nil {
		return rt
	}

	name, err := GetNamesFromTags(t.Tags)
	if err != nil {
		name = *t.TransitGatewayRouteTableId
	}

	rt.ID = *t.TransitGatewayRouteTableId
	rt.Data = t
	rt.Name = name

	return rt
}
