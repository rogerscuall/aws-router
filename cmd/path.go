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

	"github.com/rogerscuall/aws-router/aws/awsrouter"
	"github.com/spf13/cobra"
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
		ctx := context.TODO()

		fmt.Println("path called")
		fmt.Println("args:", args)
		srcIPAddress := net.ParseIP(args[0])
		if srcIPAddress == nil {
			app.ErrorLog.Println("invalid source IP address:", args[0])
		}
		dstIPAddress := net.ParseIP(args[1])
		if dstIPAddress == nil {
			app.ErrorLog.Println("invalid destination IP address:", args[1])
		}
		tgws, err := app.UpdateRouting(ctx)
		if err != nil {
			app.ErrorLog.Println("error updating routing:", err)
		}
		for _, tgw := range tgws {
			fmt.Printf("Transit Gateway Name: %s\n", tgw.Name)
			if len(tgw.RouteTables) > 0 {
				tgw.UpdateTgwRouteTablesAttachments(context.TODO(), app.RouterClient)
				tgwPath := awsrouter.NewAttPath()
				tgwPath.Tgw = tgw
				tgwPath.Walk(context.TODO(), app.RouterClient, srcIPAddress, dstIPAddress)
				fmt.Println("Path:", tgwPath.String())
			} else {
				fmt.Println("No Route Tables found")
			}
		}
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
