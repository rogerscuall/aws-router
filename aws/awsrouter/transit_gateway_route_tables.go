package awsrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/alexeyco/simpletable"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/fatih/color"
)

// TgwRouteTable holds the Route Table ID, a list of routes and other RouteTable info.
// Represents a Route Table of a Transit Gateway in AWS.
type TgwRouteTable struct {
	ID          string
	Name        string
	Data        types.TransitGatewayRouteTable
	Routes      []types.TransitGatewayRoute
	Attachments []TgwAttachment
}

// Bytes returns the JSON representation of the TgwRouteTable as a slice of bytes.
func (t *TgwRouteTable) Bytes() []byte {
	b, _ := json.Marshal(t)
	return b
}

// BestRouteToIP returns the best route to a given IP address for a given TgwRouteTable.
// Only one route can be the best route, and is returned.
// If no route is found, the function returns the empty TransitGatewayRoute.
func (t TgwRouteTable) BestRouteToIP(ipAddress net.IP) (types.TransitGatewayRoute, error) {
	// mask is the subnet mask
	var mask net.IPMask

	// result is the route table with the longest prefix match or the higher mask.
	result := types.TransitGatewayRoute{}
	for _, route := range t.Routes {
		_, subnet, err := net.ParseCIDR(*route.DestinationCidrBlock)
		if err != nil {
			return types.TransitGatewayRoute{}, fmt.Errorf("error parsing the CIDR %w", err)
		}
		if subnet.Contains(ipAddress) {
			// currentMask is the mask for the current route.
			currentMask := subnet.Mask
			// currentMaskSize is the mask length for the current route.
			currentMaskSize, _ := currentMask.Size()
			if mask == nil {
				mask = currentMask
				result = route
			} else {
				maskSize, _ := mask.Size()
				if currentMaskSize > maskSize {
					mask = currentMask
					result = route
				}
			}
		}
	}
	return result, nil
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

// Update the attachments of a TgwRouteTable.
func (t *TgwRouteTable) UpdateAttachments(ctx context.Context, api AwsRouter) error {
	// get the attachments for the route table
	input := TgwRouteTableAssociationInputFilter(t.ID)
	attachments, err := GetTgwRouteTableAssociations(ctx, api, input)
	if err != nil {
		return err
	}
	if len(attachments.Associations) < 1 {
		t.Attachments = []TgwAttachment{}
		return nil
	}
	for _, a := range attachments.Associations {
		attType := fmt.Sprint(a.ResourceType)
		newAttachment := TgwAttachment{
			ID:         *a.TransitGatewayAttachmentId,
			ResourceID: *a.ResourceId,
			Type:       attType,
		}
		t.Attachments = append(t.Attachments, newAttachment)
	}
	return nil
}

// TgwRouteTableSelectionPriority select the best route table from a list of TgwRouteTables to the specific destination.
//
func TgwRouteTableSelectionPriority(rts []*TgwRouteTable, src net.IP) (*TgwRouteTable, error) {
	var srcAttachment *TgwAttachment
	// cfg, err := config.LoadDefaultConfig(context.TODO())
	// if err != nil {
	// 	return nil, fmt.Errorf("error loading config: %w", err)
	// }
	// client := ec2.NewFromConfig(cfg)
	for _, rt := range rts {
		// r is the best route to an IP address
		r, err := rt.BestRouteToIP(src)
		if err != nil {
			return nil, err
		}
		if r.DestinationCidrBlock == nil {
			continue
		}
		switch r.Type {
		case types.TransitGatewayRouteTypePropagated:
			srcAttachment = GetAttachmentsFromTgwRoute(r)[0]
			break
		case types.TransitGatewayRouteTypeStatic:
			fmt.Println("Not implemented")
			// find if the attachment is associated with the same route table
			// where the route is.
		}
	}
	fmt.Println("Attachment is: ", srcAttachment.ResourceID)
	return nil, nil
}

// findBestRoutePrefix find the best route prefix from a list of TgwRouteTables to specific address
func findBestRoutePrefix(rts []*TgwRouteTable, ipAddr net.IP) (net.IPNet, error) {
	// brp is the best route prefix CIDR
	var brp net.IPNet

	for _, rt := range rts {
		// r is the best route to an IP address
		r, err := rt.BestRouteToIP(ipAddr)
		if err != nil {
			return net.IPNet{}, err
		}
		if r.DestinationCidrBlock == nil {
			continue
		}
		_, currentSubnet, err := net.ParseCIDR(*r.DestinationCidrBlock)
		if err != nil {
			return net.IPNet{}, fmt.Errorf("error parsing the CIDR for %v. %w", rt.Data, err)
		}
		if brp.IP == nil {
			brp = *currentSubnet
		}
		if brp.Contains(currentSubnet.IP) {
			brp = *currentSubnet
		}
	}
	return brp, nil
}

// FilterRouteTableRoutesPerPrefix returns only the routes in the route table that match specific prefix.
// Every Route Table has only one route per prefix.
// The return list is created out of new TgwRouteTable structs, that copy only the matching route to the new table.
func FilterRouteTableRoutesPerPrefix(rts []*TgwRouteTable, prefix net.IPNet) ([]TgwRouteTable, error) {
	var result []TgwRouteTable
	for _, rt := range rts {
		for _, r := range rt.Routes {
			_, currentSubnet, err := net.ParseCIDR(*r.DestinationCidrBlock)
			if err != nil {
				return nil, fmt.Errorf("error parsing the CIDR for %v. %w", rt.Data, err)
			}
			currentSubnetMask := currentSubnet.Mask.String()
			if currentSubnet.IP.Equal(prefix.IP) && prefix.Mask.String() == currentSubnetMask {
				// Create a new TgwRouteTable from the current route table.
				newRt := TgwRouteTable{
					ID:     rt.ID,
					Name:   rt.Name,
					Routes: []types.TransitGatewayRoute{r},
				}
				result = append(result, newRt)
			}
		}
	}
	return result, nil
}

// PrintRoutesInTable creates a table to print the routes in a route table.
func (t *TgwRouteTable) PrintRoutesInTable() {
	// Header Color
	headerColor := color.New(color.FgBlue, color.Bold)

	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: headerColor.Sprint("Destination CIDR")},
			{Align: simpletable.AlignCenter, Text: headerColor.Sprint("State")},
			{Align: simpletable.AlignCenter, Text: headerColor.Sprint("Route Type")},
			{Align: simpletable.AlignCenter, Text: headerColor.Sprint("Prefix List")},
		},
	}
	// Blackhole routes color
	blackholeColor := color.New(color.FgHiRed, color.Italic)
	// Regular routes color
	regularColor := color.New(color.FgHiGreen, color.Bold)
	for _, route := range t.Routes {
		dstCidr := *route.DestinationCidrBlock
		state := fmt.Sprint(route.State)
		routeType := fmt.Sprint(route.Type)
		prefixList := "-"
		if route.PrefixListId != nil {
			prefixList = fmt.Sprint(*route.PrefixListId)
		}
		if route.PrefixListId == nil {
			prefixList = "-"
		}
		if state == "active" {
			state = regularColor.Sprint(state)
			dstCidr = regularColor.Sprint(dstCidr)
			routeType = regularColor.Sprint(routeType)
			prefixList = regularColor.Sprint(prefixList)
		} else {
			state = blackholeColor.Sprint(state)
			dstCidr = blackholeColor.Sprint(dstCidr)
			routeType = blackholeColor.Sprint(routeType)
			prefixList = blackholeColor.Sprint(prefixList)

		}

		row := []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: dstCidr},
			{Align: simpletable.AlignCenter, Text: state},
			{Align: simpletable.AlignCenter, Text: routeType},
			{Align: simpletable.AlignCenter, Text: prefixList},
		}
		table.Body.Cells = append(table.Body.Cells, row)
	}
	table.Footer = &simpletable.Footer{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Span: 4, Text: headerColor.Sprintf("Route Table Name: %v", t.Name)},
		},
	}
	fmt.Println(table.String())
}
