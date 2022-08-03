package awsrouter

// TgwPath	 represent the path through the Transit Gateway, from the source to the destination.
type TgwPath struct {
	SourceAtt             []*TgwAttachment
	DestinationAtt        []*TgwAttachment
	SourceRouteTable      TgwRouteTable
	DestinationRouteTable TgwRouteTable
	TransitGatewayID      string
}

// func (t *TransitGatewayPath) FindSourceAttachment(sourceIPAddress net.IP) TgwAttachment {
// 	for _, routeTable := range tgw.RouteTables {
// 		fmt.Println("routeTable:", routeTable.ID)
// 		result, err := routeTable.BestRouteToIP(sourceIPAddress)
// 		if err != nil {
// 			cobra.CheckErr(err)
// 		}
// 		fmt.Println("result:", *result.DestinationCidrBlock)
// 	}
// }
