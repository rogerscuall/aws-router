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
// If no route is found, the function returns nil.
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
