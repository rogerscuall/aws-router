package awsrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

//Tgw is the main data-structure, holds ID, Name, a list of TgwRouteTable and other TGW info.
// Represents a Transit Gateway in AWS.
type Tgw struct {
	ID          string
	Name        string
	RouteTables []*TgwRouteTable
	Data        types.TransitGateway
}

// Build a Tgw from a aws TGW.
func NewTgw(tgw types.TransitGateway) *Tgw {
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

// GetAllTgws returns a list of all the Transit Gateways in the account for specific region
func GetAllTgws(ctx context.Context, api AwsRouter) ([]*Tgw, error) {
	input := &ec2.DescribeTransitGatewaysInput{}
	result, err := GetTgw(ctx, api, input)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Transit Gateways: %w", err)
	}
	var tgws []*Tgw
	for _, tgw := range result.TransitGateways {
		tgws = append(tgws, NewTgw(tgw))
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


// GetTgwPath returns the best path TgwPath for two endpoints.
// 
func (t *Tgw) GetTgwPath(src, dest net.IPAddr) (*TgwPath, error) {
	/*
	IDEA is as follows:
	1. loop over all the route tables.
	2. find the best route in the route table.
	3. find if that route points to an attachment that is on that same route table.
	4. if it is, the route table is the closest to the address.
	*/
	return nil, nil
}