package awsrouter

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

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

// FindTheMostSpecificRoute from a list of TgwRouteTable identifies the best most specific routes to reach a destination.
// Each Route Table can have one or more routes to reach a destination, this function will find the most specific CIDR among all the route tables.
// If more than one route table has that specific CIDR it will be returned in []TgwRouteTable, each element in []TgwRouteTable has only the route or routes that match that most specific CIDR.

func FindTheMostSpecificRoute(rts []TgwRouteTable, src net.IP) (net.IPNet, []TgwRouteTable, error) {
	var subnet net.IPNet
	var bestMask int
	var listOfRouteTables []TgwRouteTable
	for _, rt := range rts {
		r, err := rt.BestRouteToIP(src)
		if err != nil {
			return net.IPNet{}, nil, err
		}
		if r.DestinationCidrBlock == nil {
			continue
		}
		_, currentSubnet, _ := net.ParseCIDR(*r.DestinationCidrBlock)
		currentMask, _ := currentSubnet.Mask.Size()
		if bestMask == 0 || currentMask > bestMask {
			subnet = *currentSubnet
			newRT := rt
			newRT.Routes = []types.TransitGatewayRoute{r,}
			listOfRouteTables = append(listOfRouteTables, newRT)
			bestMask, _ = subnet.Mask.Size()
		}
	}
	return subnet, listOfRouteTables, nil
}
