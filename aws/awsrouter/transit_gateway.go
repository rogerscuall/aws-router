package awsrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/rogerscuall/aws-router/ports"
)

// Tgw is the main data-structure, holds ID, Name, a list of TgwRouteTable and other TGW info.
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
func (t *Tgw) UpdateRouteTables(ctx context.Context, api ports.AWSRouter) error {
	// Update the Route Tables
	inputTgwRouteTable := ports.TgwRouteTableInputFilter([]string{t.ID})
	resultTgwRouteTable, err := ports.GetTgwRouteTables(context.TODO(), api, inputTgwRouteTable)
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
func (t *Tgw) UpdateTgwRoutes(ctx context.Context, api ports.AWSRouter) error {
	var wg sync.WaitGroup
	var err error
	for _, tgwRouteTable := range t.RouteTables {
		wg.Add(1)
		go func(routeTable *TgwRouteTable) {
			defer wg.Done()
			inputTgwSearchRoutes := ports.TgwSearchRoutesInputFilter(routeTable.ID)
			resultTgwSearchRoutes, err := ports.GetTgwRoutes(context.TODO(), api, inputTgwSearchRoutes)
			if err != nil {
				err = fmt.Errorf("error retrieve the table %s %w", routeTable.ID, err)
			}
			routeTable.Routes = resultTgwSearchRoutes.Routes
		}(tgwRouteTable)
	}
	wg.Wait()
	return err
}

// UpdateTgwRouteTablesAttachments updates the Attachments of a TgwRouteTable.
func (t *Tgw) UpdateTgwRouteTablesAttachments(ctx context.Context, api ports.AWSRouter) error {
	for _, tgwRouteTable := range t.RouteTables {
		input := ports.TgwRouteTableAssociationInputFilter(tgwRouteTable.ID)
		result, err := ports.GetTgwRouteTableAssociations(ctx, api, input)
		if err != nil {
			return fmt.Errorf("error retrieving Transit Gateway Route Table Associations: %w", err)
		}
		err = tgwRouteTable.UpdateAttachments(ctx, result)
		if err != nil {
			return fmt.Errorf("error updating the route table %s %w", tgwRouteTable.ID, err)
		}

		// Update attachment names
		for _, att := range tgwRouteTable.Attachments {
			attInput := ec2.DescribeTransitGatewayAttachmentsInput{}
			attInput.TransitGatewayAttachmentIds = []string{att.ID}
			attOutput, err := ports.GetTgwAttachments(ctx, api, &attInput)
			if err != nil {
				return fmt.Errorf("error retrieving Transit Gateway Attachments: %w", err)
			}
			if len(attOutput.TransitGatewayAttachments) != 1 {
				fmt.Print("there is more than one attachment with the same ID")
			}
			tags := attOutput.TransitGatewayAttachments[0].Tags
			if len(tags) == 0 {
				continue
			}
			name, err := GetNamesFromTags(tags)
			if err == nil {
				att.Name = name
				fmt.Printf("Attachment %s has name %s\n", att.ID, att.Name)
			}
		}
	}
	return nil
}

// GetAllTgws returns a list of all the Transit Gateways in the account for specific region
func GetAllTgws(ctx context.Context, api ports.AWSRouter) ([]*Tgw, error) {
	input := &ec2.DescribeTransitGatewaysInput{}
	result, err := ports.GetTgw(ctx, api, input)
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
func UpdateRouting(ctx context.Context, api ports.AWSRouter) ([]*Tgw, error) {
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

func (t *Tgw) GetTgwRouteTableByID(id string) (*TgwRouteTable, error) {
	for _, tgwRouteTable := range t.RouteTables {
		if tgwRouteTable.ID == id {
			return tgwRouteTable, nil
		}
	}
	return nil, fmt.Errorf("route table %s not found", id)
}

// GetDirectlyConnectedAttachment returns the route and attachment that is directly connected to the ipAddress.
// In the case of ECMP we can have more than one attachment per route.
// In the majority of the cases we will have only one attachment per route.
// If we have two or more attachments this function is unable will return the first attachment and the route table associated to it.
func (t *Tgw) GetDirectlyConnectedAttachment(ipAddress net.IP) (TgwRouteTable, []*TgwAttachment, error) {
	var rt TgwRouteTable
	var attachment []*TgwAttachment
	// find the best route prefix to the ipAddress
	prefix, err := findBestRoutePrefix(t.RouteTables, ipAddress)
	if err != nil {
		return rt, attachment, fmt.Errorf("error finding the best route prefix: %w", err)
	}
	// find the route tables that have route to the prefix
	listRouteTable, err := FilterRouteTableRoutesPerPrefix(t.RouteTables, prefix)
	if err != nil {
		return rt, attachment, fmt.Errorf("error finding the route table: %w", err)
	}
	// find the attachment that is directly connected to the prefix
	attachment = getDirectlyConnectedAttachmentFromTgwRoute(listRouteTable)
	if len(attachment) == 0 {
		return rt, attachment, fmt.Errorf("error finding the attachment: %w", err)
	}
	// find the route table associated to the attachment
	// In the case of ECMP we can have more than one attachment per route.
	// In the majority of the cases we will have only one attachment per route.
	// If we have two or more attachments this function is unable will return the first attachment and the route table associated to it.
	for _, tgwRt := range t.RouteTables {
		for _, att := range tgwRt.Attachments {
			if att.ID == attachment[0].ID {
				rt = *tgwRt
				return rt, attachment, nil
			}
		}
	}
	return rt, attachment, nil
}
