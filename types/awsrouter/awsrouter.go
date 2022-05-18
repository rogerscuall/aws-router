package awsrouter

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
)

type AwsRouter interface {
	DescribeTransitGateways(ctx context.Context, params *ec2.DescribeTransitGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewaysOutput, error)
	DescribeTransitGatewayRouteTables(ctx context.Context, params *ec2.DescribeTransitGatewayRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewayRouteTablesOutput, error)
	SearchTransitGatewayRoutes(ctx context.Context, params *ec2.SearchTransitGatewayRoutesInput, optFns ...func(*ec2.Options)) (*ec2.SearchTransitGatewayRoutesOutput, error)
}

//Tgw is the main datastructure, holds ID, Name, a list of TgwRouteTable and other TGW info.
// TODO: Implement a constructor that adds Name.
// Remove the Tgw from the fields.
type Tgw struct {
	TgwId          string
	TgwName        string
	TgwRouteTables []*TgwRouteTable
	TgwData        types.TransitGateway
}

// TgwRouteTable holds the Route Table ID, a list of routes and other RouteTable info.
type TgwRouteTable struct {
	TgwRouteTableId string
	TwgRouteTable   types.TransitGatewayRouteTable
	TgwRoutes       []types.TransitGatewayRoute
}

// GetTgwRoutes it will populate the TgwRouteTable with the Routes.
// TODO: Add some sentinel error message to notify if a the calls to GetTgwRoutes fail.
// The call use concurrency, because on each Tgw there are multiple Route Tables.
// TODO: add testing and include race condition detection.
// Each Tgw has a list of RouteTables, each RouteTable gets is own goroutine.
func (t *Tgw) GetTgwRoutes(ctx context.Context, api AwsRouter) error {
	var wg sync.WaitGroup
	var err error
	for _, tgwRouteTable := range t.TgwRouteTables {
		wg.Add(1)
		go func(routeTable *TgwRouteTable) {
			defer wg.Done()
			inputTgwSearchRoutes := TgwSearchRoutesInputFilter(routeTable.TgwRouteTableId)
			resultTgwSearchRoutes, err := GetTgwRoutes(context.TODO(), api, inputTgwSearchRoutes)
			if err != nil {
				err = fmt.Errorf("wrapping the original error => %w", err)
			}
			routeTable.TgwRoutes = resultTgwSearchRoutes.Routes
		}(tgwRouteTable)
	}
	wg.Wait()
	return err
}

// GetTgwRouteTables
// TODO: implement a GetTgwRouteTables similar to the previous one.

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
// tgwIDs is a list of Transit Gateway IDs.
// and empty tgwIDs creates a filter that returns all Transit Gateway Route Tables in the account.
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
// routeFilters is an optional list of filters used to specify routes matching operations.
// like longest-prefix-match, prefix-match, or exact-match.
// filters are analyzed like a logical AND if more than one is specified.

func TgwSearchRoutesInputFilter(tgwRtID string, routeFilters ...types.Filter) *ec2.SearchTransitGatewayRoutesInput {
	// TODO: filters are analyzed like a logical AND if more than one is specified, not sure if makes sense for routing to have more than one.
	var filters []types.Filter
	//default filter if no filters are provided
	if len(routeFilters) == 0 {
		filters = []types.Filter{
			{
				Name:   aws.String("state"),
				Values: []string{"active"},
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

// GetTgwRoutesConcurrently returns a list of the Transit Gateway Routes that match the input filter for specific Route Table.
// func GetTgwRoutesConcurrently(ctx context.Context, client *ec2.Client, tgws []*) ( error) {
// 	var wg sync.WaitGroup
// 	var tgwRoutes []types.TransitGatewayRoute
// 	for _, routeTable := range routeTables {
// 		wg.Add(1)
// 		go func(routeTable types.TransitGatewayRouteTable) {
// 			defer wg.Done()
// 			inputTgwSearchRoutes := TgwSearchRoutesInputFilter(*routeTable.TransitGatewayRouteTableId)
// 			resultTgwSearchRoutes, err := GetTgwRoutes(ctx, client, inputTgwSearchRoutes)
// 			if err != nil {
// 				fmt.Printf("Error retrieving the routes in the route table %s %v\n", *routeTable.TransitGatewayRouteTableId, err)
// 			}
// 			tgwRoutes = append(tgwRoutes, resultTgwSearchRoutes.Routes...)
// 		}(routeTable)
// 	}
// 	wg.Wait()
// 	return tgwRoutes, nil
// }
