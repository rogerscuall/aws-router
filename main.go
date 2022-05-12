package main

import (
	"context"
	"fmt"
	"go-aws-routing/types/awsrouter"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	client := ec2.NewFromConfig(cfg)

	input := awsrouter.TgwInputFilter([]string{"tgw-046b205ac9b96e5ae"})

	result, err := awsrouter.GetTgw(context.TODO(), client, input)
	tgwID := "aws.StringValue(result.TransitGateways[0].TransitGatewayId)"
	fmt.Println(tgwID)

	var tgwIDList []string
	tgwIDList = append(tgwIDList, aws.StringValue(result.TransitGateways[0].TransitGatewayId))
	fmt.Println(tgwIDList)
	inputTgwRouteTable := awsrouter.TgwRouteTableInputFilter(tgwIDList)
	resultTgwRouteTable, err := awsrouter.GetTgwRouteTables(context.TODO(), client, inputTgwRouteTable)
	var TgwRouteTableList []string
	for _, tgwRouteTable := range resultTgwRouteTable.TransitGatewayRouteTables {
		TgwRouteTableList = append(TgwRouteTableList, aws.StringValue(tgwRouteTable.TransitGatewayRouteTableId))
	}
	fmt.Println(TgwRouteTableList)
	routeFilter1 := types.Filter{
		Name: aws.String("route-search.longest-prefix-match"),
		Values: []string{
			"10.0.1.1",
		},
	}
	routeFilter2 := types.Filter{
		Name: aws.String("route-search.exact-match"),
		Values: []string{
			"10.0.1.0/24",
			"10.1.0.0/16",
			"10.0.3.0/24",
		},
	}
	inputTgwSearchRoutes := awsrouter.TgwSearchRoutesInputFilter("tgw-rtb-078ad023a6d3d63df", routeFilter1, routeFilter2)
	resultTgwSearchRoutes, err := awsrouter.GetTgwRoutes(context.TODO(), client, inputTgwSearchRoutes)
	for _, tgwSearchRoutes := range resultTgwSearchRoutes.Routes {
		fmt.Println(aws.StringValue(tgwSearchRoutes.DestinationCidrBlock))
	}
	fmt.Println("Second Filter")
	inputTgwSearchRoutes = awsrouter.TgwSearchRoutesInputFilter("tgw-rtb-078ad023a6d3d63df", routeFilter2)
	resultTgwSearchRoutes, err = awsrouter.GetTgwRoutes(context.TODO(), client, inputTgwSearchRoutes)
	for _, tgwSearchRoutes := range resultTgwSearchRoutes.Routes {
		fmt.Println(aws.StringValue(tgwSearchRoutes.DestinationCidrBlock))
	}
}