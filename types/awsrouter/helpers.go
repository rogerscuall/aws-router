package awsrouter

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func GetAllRoutes(ctx context.Context, client *ec2.Client) error {
	inputTgwRouteTable := TgwRouteTableInputFilter([]string{})
	resultTgwRouteTable, err := GetTgwRouteTables(context.TODO(), client, inputTgwRouteTable)
	if err != nil {
		return fmt.Errorf("error getting tgw route tables: %w", err)
	}
	var wg sync.WaitGroup
	var routeTableMap = make(map[string][]types.TransitGatewayRoute)
	for _, tgwRouteTable := range resultTgwRouteTable.TransitGatewayRouteTables {
		wg.Add(1)
		go func(routeTable types.TransitGatewayRouteTable, wg *sync.WaitGroup) {
			defer wg.Done()
			var mapKey string
			// Convert into an optional function pass as a parameter to adorn the map key
			f := func() (mapKey string) {
				mapKey = strings.Join([]string{*routeTable.TransitGatewayRouteTableId, *routeTable.TransitGatewayRouteTableId}, "->")
				for _, route := range routeTable.Tags {
					if *route.Key == "Name" {
						return strings.Join([]string{*route.Value, *routeTable.TransitGatewayRouteTableId}, "->")
					}
				}
				return mapKey
			}
			mapKey = f()
			inputTgwSearchRoutes := TgwSearchRoutesInputFilter(*routeTable.TransitGatewayRouteTableId)
			resultTgwSearchRoutes, err := GetTgwRoutes(context.TODO(), client, inputTgwSearchRoutes)
			if err != nil {
				panic("error, " + err.Error())
			}

			routeTableMap[mapKey] = resultTgwSearchRoutes.Routes
		}(tgwRouteTable, &wg)
	}
	wg.Wait()
	for k, v := range routeTableMap {
		fmt.Println("Route table:", k)
		for _, route := range v {
			fmt.Printf("\t-: %v", *route.DestinationCidrBlock)
			fmt.Printf("=> %v", route.State)
			for _, transitGatewayAttachment := range route.TransitGatewayAttachments {
				fmt.Printf("=> %v\n", *transitGatewayAttachment.TransitGatewayAttachmentId)
			}
		}
	}
	return nil
}

// getNamesFromTags returns the name tags if exist, if not it will signal with an error.
func GetNamesFromTags(tags []types.Tag) (string, error) {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			return *tag.Value, nil
		}
	}
	return "", fmt.Errorf("no name tag found")
}
