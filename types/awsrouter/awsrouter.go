/*
This library is a collection of calls to work with routing information on AWS.
*/
package awsrouter

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
)

// AwsRouter is an interface with the methods needed for routing.
type AwsRouter interface {
	DescribeTransitGateways(ctx context.Context, params *ec2.DescribeTransitGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewaysOutput, error)
	DescribeTransitGatewayRouteTables(ctx context.Context, params *ec2.DescribeTransitGatewayRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeTransitGatewayRouteTablesOutput, error)
	SearchTransitGatewayRoutes(ctx context.Context, params *ec2.SearchTransitGatewayRoutesInput, optFns ...func(*ec2.Options)) (*ec2.SearchTransitGatewayRoutesOutput, error)
}

//Tgw is the main data-structure, holds ID, Name, a list of TgwRouteTable and other TGW info.
// Represents a Transit Gateway in AWS.
type Tgw struct {
	ID          string
	Name        string
	RouteTables []*TgwRouteTable
	Data        types.TransitGateway
}

// Build a Tgw from a aws TGW.
func newTgw(tgw types.TransitGateway) *Tgw {
	// t is the Tgw
	t := &Tgw{}

	// if tgw has no id return the empty Tgw struct.
	if tgw.TransitGatewayId == nil {
		return t
	}

	name, err := GetNamesFromTags(tgw.Tags)
	if err != nil {
		name = *tgw.TransitGatewayId
	}
	t.ID = *tgw.TransitGatewayId
	t.Name = name
	t.Data = tgw
	return t
}

// Bytes returns the JSON representation of the Tgw as a slice of bytes.
func (t *Tgw) Bytes() []byte {
	b, _ := json.Marshal(t)
	return b
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

// UpdateRouteTables updates the field TgwRouteTables on a Tgw.
// An error will stop the processing returning the error wrapped.
func (t *Tgw) UpdateRouteTables(ctx context.Context, api AwsRouter) error {
	inputTgwRouteTable := TgwRouteTableInputFilter([]string{t.ID})
	resultTgwRouteTable, err := GetTgwRouteTables(context.TODO(), api, inputTgwRouteTable)
	if err != nil {
		return fmt.Errorf("error updating the route tables %w", err)
	}
	for _, tgwRouteTable := range resultTgwRouteTable.TransitGatewayRouteTables {
		t.RouteTables = append(t.RouteTables, newTgwRouteTable(tgwRouteTable))
	}
	return nil
}

// UpdateTgwRoutes updates the routes of a route table.
//
// TODO: Add some sentinel error message to notify if a the calls to UpdateTgwRoutes fail.
//
// The call use concurrency, because on each Tgw there are multiple Route Tables.
//
// TODO: add testing and include race condition detection.
//
// Each Tgw has a list of TgwRouteTables, each RouteTable gets is own goroutine.
func (t *Tgw) UpdateTgwRoutes(ctx context.Context, api AwsRouter) error {
	var wg sync.WaitGroup
	var err error
	for _, tgwRouteTable := range t.RouteTables {
		wg.Add(1)
		go func(routeTable *TgwRouteTable) {
			defer wg.Done()
			inputTgwSearchRoutes := TgwSearchRoutesInputFilter(routeTable.ID)
			resultTgwSearchRoutes, err := GetTgwRoutes(context.TODO(), api, inputTgwSearchRoutes)
			if err != nil {
				err = fmt.Errorf("error retrieve the table %s %w", routeTable.ID, err)
			}
			routeTable.Routes = resultTgwSearchRoutes.Routes
		}(tgwRouteTable)
	}
	wg.Wait()
	return err
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

// ExportRouteTableRoutesCsv creates a CSV with all the routes in one Tgw Route Table.
func ExportRouteTableRoutesCsv(w *csv.Writer, tgwrt TgwRouteTable) error {
	defer w.Flush()
	w.Write([]string{"Destination CIDR Block", "State", "Type"})
	for _, route := range tgwrt.Routes {
		state := fmt.Sprint(route.State)
		routeType := fmt.Sprint(route.Type)
		err := w.Write([]string{*route.DestinationCidrBlock, state, routeType})
		if err != nil {
			return fmt.Errorf("error writing to csv: %w", err)
		}
	}
	return nil
}

// GetAllTgws returns a list of all the Transit Gateways in the account for specific region
func GetAllTgws(ctx context.Context, api AwsRouter) ([]*Tgw, error) {
	input := &ec2.DescribeTransitGatewaysInput{}
	result, err := GetTgw(ctx, api, input)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Transit Gateways: %w", err)
	}
	var tgws []*Tgw
	for _, tgw := range result.TransitGateways {
		tgws = append(tgws, newTgw(tgw))
	}
	return tgws, nil
}

// UpdateRouting this functions is a helper that will update all routing information from AWS, returning a list of Tgw.
// The function will try to gather all the Route Tables and all the routes in the Route Tables.
// The function will return an error if it fails to gather a Transit Gateway or a Route Table, but it will continue
// if it fails to gather a route.
func UpdateRouting(ctx context.Context, api AwsRouter) ([]*Tgw, error) {
	tgws, err := GetAllTgws(ctx, api)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Transit Gateways: %w", err)
	}
	for _, tgw := range tgws {
		if tgw.UpdateRouteTables(context.TODO(), api); err != nil {
			return nil, fmt.Errorf("error retrieving Transit Gateway Route Tables: %w", err)
		}
	}
	// Get all routes from all route tables
	for _, tgw := range tgws {
		tgw.UpdateTgwRoutes(context.TODO(), api)
	}

	return tgws, nil
}
