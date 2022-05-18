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

// func main() {

// 	defer func() {
// 		if err := recover(); err != nil {
// 			fmt.Println(err)
// 		}
// 	}()
// 	cfg, err := config.LoadDefaultConfig(context.TODO())
// 	if err != nil {
// 		panic("configuration error, " + err.Error())
// 	}
// 	client := ec2.NewFromConfig(cfg)
// 	var wg sync.WaitGroup

// 	// Get all the TGW IDs
// 	input := awsrouter.TgwInputFilter([]string{})
// 	result, err := awsrouter.GetTgw(context.TODO(), client, input)
// 	for _, tgw := range result.TransitGateways {
// 		fmt.Println(*tgw.TransitGatewayId)
// 	}

// 	// Get a list of all the TGW IDs
// 	var tgwIDList []string
// 	for _, tgw := range result.TransitGateways {
// 		tgwIDList = append(tgwIDList, *tgw.TransitGatewayId)
// 	}
// 	fmt.Println("List of TGW ID", tgwIDList)

// 	// Get all the Route Tables for all the TGWs
// 	inputTgwRouteTable := awsrouter.TgwRouteTableInputFilter(tgwIDList)
// 	resultTgwRouteTable, err := awsrouter.GetTgwRouteTables(context.TODO(), client, inputTgwRouteTable)

// 	// // a test
// 	// inputTgwRouteTable1 := awsrouter.TgwRouteTableInputFilter([]string{})
// 	// resultTgwRouteTable1, err := awsrouter.GetTgwRouteTables(context.TODO(), client, inputTgwRouteTable1)
// 	// fmt.Println("This is a test")
// 	// for _, tgwRouteTable := range resultTgwRouteTable1.TransitGatewayRouteTables {

// 	// 	fmt.Println(*tgwRouteTable.TransitGatewayRouteTableId)
// 	// }
// 	fmt.Println("This is a test")
// 	// Get a list of all the TGW Route Tables IDs.
// 	var TgwRouteTableList []string
// 	for _, tgwRouteTable := range resultTgwRouteTable.TransitGatewayRouteTables {
// 		TgwRouteTableList = append(TgwRouteTableList, *tgwRouteTable.TransitGatewayRouteTableId)
// 	}
// 	fmt.Println("All TGW Route tables", TgwRouteTableList)

// 	var routeTableMap = make(map[string][]types.TransitGatewayRoute)
// 	for _, tgwRouteTable := range resultTgwRouteTable.TransitGatewayRouteTables {
// 		wg.Add(1)
// 		go func(routeTable types.TransitGatewayRouteTable, wg *sync.WaitGroup) {
// 			defer wg.Done()
// 			var mapKey string
// 			// Convert into an optional function pass as a parameter to adorn the map key
// 			f := func() (mapKey string) {
// 				mapKey = *routeTable.TransitGatewayRouteTableId
// 				for _, route := range routeTable.Tags {
// 					if *route.Key == "Name" {
// 						return *route.Value
// 					}
// 				}
// 				return mapKey
// 			}
// 			mapKey = f()
// 			inputTgwSearchRoutes := awsrouter.TgwSearchRoutesInputFilter(*routeTable.TransitGatewayRouteTableId)
// 			resultTgwSearchRoutes, err := awsrouter.GetTgwRoutes(context.TODO(), client, inputTgwSearchRoutes)
// 			if err != nil {
// 				panic("error, " + err.Error())
// 			}

// 			routeTableMap[mapKey] = resultTgwSearchRoutes.Routes
// 		}(tgwRouteTable, &wg)
// 	}
// 	wg.Wait()
// 	for k, v := range routeTableMap {
// 		fmt.Println("Route table:", k)
// 		for _, route := range v {
// 			fmt.Printf("\t-: %v", *route.DestinationCidrBlock)
// 			fmt.Printf("=> %v", route.State)
// 			for _, transitGatewayAttachment := range route.TransitGatewayAttachments {
// 				fmt.Printf("=> %v\n", *transitGatewayAttachment.TransitGatewayAttachmentId)
// 			}
// 		}
// 	}
// }

// // func getAllRoutesOnTgw(client *ec2.Client, tgwID string) (map[string][]types.TransitGatewayRoute, error) {
// // 	inputTgwRouteTable := awsrouter.TgwRouteTableInputFilter(tgwID)
// // 	resultTgwRouteTable, err := awsrouter.GetTgwRouteTables(context.TODO(), client, inputTgwRouteTable)

// // 	var wg sync.WaitGroup
// // 	var routeTableMap = make(map[string][]types.TransitGatewayRoute)

// // 	for _, tgwRouteTable := range TgwRouteTableList {
// // 		wg.Add(1)
// // 		go func(s string, wg *sync.WaitGroup) {
// // 			defer wg.Done()
// // 			inputTgwSearchRoutes := awsrouter.TgwSearchRoutesInputFilter(s)
// // 			resultTgwSearchRoutes, err := awsrouter.GetTgwRoutes(context.TODO(), client, inputTgwSearchRoutes)
// // 			if err != nil {
// // 				panic("error, " + err.Error())
// // 			}
// // 			routeTableMap[s] = resultTgwSearchRoutes.Routes
// // 		}(tgwRouteTable, &wg)
// // 	}
// // 	wg.Wait()
// // }