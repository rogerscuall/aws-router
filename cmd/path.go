/*
Copyright © 2022 Roger Gomez rogerscuall@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"gitlab.presidio.com/rgomez/aws-router/aws/awsrouter"
)

// pathCmd represents the path command
var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("path called")
		fmt.Println("args:", args)
		var err error
		defer func() {
			if err != nil {
				cobra.CheckErr(err)
			}
		}()
		srcIPAddress := net.ParseIP(args[0])
		if srcIPAddress == nil {
			cobra.CheckErr(fmt.Errorf("invalid source IP address: %s", args[0]))
		}
		dstIPAddress := net.ParseIP(args[1])
		if dstIPAddress == nil {
			cobra.CheckErr(fmt.Errorf("invalid destination IP address: %s", args[0]))
		}
		cfg, err := config.LoadDefaultConfig(context.TODO())
		client := ec2.NewFromConfig(cfg)
		tgwInputFilter := awsrouter.TgwInputFilter([]string{"tgw-0424b87b4942010b0"})
		tgwOutput, err := awsrouter.GetTgw(context.TODO(), client, tgwInputFilter)
		tgw := awsrouter.NewTgw(tgwOutput.TransitGateways[0])
		fmt.Println("tgw:", tgw.ID)
		tgw.UpdateRouteTables(context.TODO(), client)
		tgw.UpdateTgwRoutes(context.TODO(), client)
		// Update the attachments
		tgw.UpdateTgwRouteTablesAttachments(context.TODO(), client)
		fmt.Println("RT Name Attach:", tgw.RouteTables[0].Name)
		fmt.Println("RT Attach:", tgw.RouteTables[0].Attachments)

		// Get the directly connected attachment for the source and destination IP address
		srcRt, srcAtts, err := tgw.GetDirectlyConnectedAttachment(srcIPAddress)
		dstRt, dstAtts, err := tgw.GetDirectlyConnectedAttachment(dstIPAddress)

		fmt.Println("srcRt:", srcRt.Name)
		fmt.Println("dstRt:", dstRt.Name)
		fmt.Println("srcAtts:", srcAtts[0].ID)
		fmt.Println("dstAtts:", dstAtts[0].ID)
		// tgwPath, err := tgw.GetTgwPath(srcIPAddress, dstIPAddress)
		// fmt.Println("Source ", tgwPath.Source)
		// fmt.Println("Destination ", tgwPath.Destination)
		// fmt.Println("TransitGatewayID ", tgwPath.TransitGatewayID)
		// fmt.Println("Path ", tgwPath.Path[0].Name)
		// fmt.Println("Path ", tgwPath.Path[1].Name)
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pathCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pathCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
