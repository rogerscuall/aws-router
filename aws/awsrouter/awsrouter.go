/*
This library is a collection of calls to work with routing information on AWS.
*/
package awsrouter

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
)

// AwsRouter is an interface with the methods needed for routing.
type AwsRouter interface {
	DescribeTransitGateways(ctx context.Context, params *ec2.DescribeTransitGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewaysOutput, error)
	DescribeTransitGatewayRouteTables(ctx context.Context, params *ec2.DescribeTransitGatewayRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewayRouteTablesOutput, error)
	SearchTransitGatewayRoutes(ctx context.Context, params *ec2.SearchTransitGatewayRoutesInput, optFns ...func(*ec2.Options)) (*ec2.SearchTransitGatewayRoutesOutput, error)
	GetTransitGatewayRouteTableAssociations(ctx context.Context, params *ec2.GetTransitGatewayRouteTableAssociationsInput, optFns ...func(*ec2.Options)) (*ec2.GetTransitGatewayRouteTableAssociationsOutput, error)
	DescribeTransitGatewayAttachments(ctx context.Context, params *ec2.DescribeTransitGatewayAttachmentsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewayAttachmentsOutput, error)
}

// TgwInputFilter returns a filter for the DescribeTransitGatewaysInput.
// tgwIDs is a list of Transit Gateway IDs.
func TgwInputFilter(tgwIDs []string) *ec2.DescribeTransitGatewaysInput {
	input := &ec2.DescribeTransitGatewaysInput{}
	if len(tgwIDs) > 0 {
		input.TransitGatewayIds = tgwIDs
	}
	return input
}

// GetTgw returns a list of the Transit Gateways that match the input filter.
func GetTgw(ctx context.Context, api AwsRouter, input *ec2.DescribeTransitGatewaysInput) (*ec2.DescribeTransitGatewaysOutput, error) {
	return api.DescribeTransitGateways(ctx, input)
}

// TgwRouteTableInputFilter returns a filter for the DescribeTransitGatewayRouteTables.
// The tgwIDs is a list of Transit Gateway IDs.
// An empty tgwIDs creates a filter that returns all Transit Gateway Route Tables in the account.
func TgwRouteTableInputFilter(tgwIDs []string) *ec2.DescribeTransitGatewayRouteTablesInput {
	if len(tgwIDs) == 0 {
		return &ec2.DescribeTransitGatewayRouteTablesInput{}
	}
	var filter []types.Filter
	filter = append(filter, types.Filter{
		Name:   aws.String("transit-gateway-id"),
		Values: tgwIDs,
	})
	input := &ec2.DescribeTransitGatewayRouteTablesInput{
		Filters: filter,
	}
	return input
}

// GetTgwRouteTables returns a list of the Transit Gateway Route Tables that match the input filter.
// and empty input filter creates a filter that returns all Transit Gateway Route Tables in the account.
func GetTgwRouteTables(ctx context.Context, api AwsRouter, input *ec2.DescribeTransitGatewayRouteTablesInput) (*ec2.DescribeTransitGatewayRouteTablesOutput, error) {
	return api.DescribeTransitGatewayRouteTables(ctx, input)
}

// TgwSearchRoutesInputFilter returns a filter for the SearchTransitGatewayRoutes.
// tgwRtID is the Transit Gateway Route Table ID.
// routeFilters is an optional list of filters used to specify routes matching operations
// like longest-prefix-match, prefix-match, or exact-match.
// If the routeFilters is empty, the default filter is used to match active and blackhole routes.
// Active and blackhole routes are should be all the routes that affect the routing inside a route table.

func TgwSearchRoutesInputFilter(tgwRtID string, routeFilters ...types.Filter) *ec2.SearchTransitGatewayRoutesInput {
	var filters []types.Filter
	//default filter if no filters are provided
	if len(routeFilters) == 0 {
		filters = []types.Filter{
			{
				Name:   aws.String("state"),
				Values: []string{"active"},
			},
			{
				Name:   aws.String("state"),
				Values: []string{"blackhole"},
			},
		}
	}
	for _, filter := range routeFilters {
		filters = append(filters, filter)
	}
	input := &ec2.SearchTransitGatewayRoutesInput{
		Filters:                    filters,
		TransitGatewayRouteTableId: aws.String(tgwRtID),
	}
	return input
}

//GetTgwRoutes returns a list of the Transit Gateway Routes that match the input filter for specific Route Table.
func GetTgwRoutes(ctx context.Context, api AwsRouter, input *ec2.SearchTransitGatewayRoutesInput) (*ec2.SearchTransitGatewayRoutesOutput, error) {
	return api.SearchTransitGatewayRoutes(ctx, input)
}

func TgwRouteTableAssociationInputFilter(tgwRtID string, attachmentFilters ...types.Filter) *ec2.GetTransitGatewayRouteTableAssociationsInput {
	var filters []types.Filter
	//default filter if no filters are provided
	if len(attachmentFilters) != 0 {
		for _, filter := range attachmentFilters {
			filters = append(filters, filter)
		}
	}

	input := &ec2.GetTransitGatewayRouteTableAssociationsInput{
		Filters:                    filters,
		TransitGatewayRouteTableId: aws.String(tgwRtID),
	}
	return input
}

func GetTgwRouteTableAssociations(ctx context.Context, api AwsRouter, input *ec2.GetTransitGatewayRouteTableAssociationsInput) (*ec2.GetTransitGatewayRouteTableAssociationsOutput, error) {
	return api.GetTransitGatewayRouteTableAssociations(ctx, input)
}

func TgwAttachmentInputFilter(attachmentFilters ...types.Filter) *ec2.DescribeTransitGatewayAttachmentsInput {
	var filters []types.Filter
	//default filter if no filters are provided
	if len(attachmentFilters) != 0 {
		for _, filter := range attachmentFilters {
			filters = append(filters, filter)
		}
	}
	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		Filters: filters,
	}
	return input
}

// TgwGetAttachments describe the configuration of the TGW Attachments.
func TgwGetAttachments(ctx context.Context, api AwsRouter, input *ec2.DescribeTransitGatewayAttachmentsInput) (*ec2.DescribeTransitGatewayAttachmentsOutput, error) {
	return api.DescribeTransitGatewayAttachments(ctx, input)
}
