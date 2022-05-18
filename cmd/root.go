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
	"os"

	"gitlab.presidio.com/rgomez/aws-router/types/awsrouter"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-aws-routing",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		defer func() {
			if err != nil {
				cobra.CheckErr(err)
			}
		}()

		fmt.Println("Getting all routes from all route tables on all the TGWs in the region")
		cfg, err := config.LoadDefaultConfig(context.TODO())
		client := ec2.NewFromConfig(cfg)
		tgwInputFilter := awsrouter.TgwInputFilter([]string{})
		resultTgw, err := awsrouter.GetTgw(context.TODO(), client, tgwInputFilter)
		var tgws []*awsrouter.Tgw
		for _, tgw := range resultTgw.TransitGateways {
			newTgw := &awsrouter.Tgw{
				TgwId:   *tgw.TransitGatewayId,
				TgwData: tgw,
			}
			tgws = append(tgws, newTgw)
		}
		// for _, tgw := range tgws {
		// 	fmt.Println("The TGW is: ", tgw.TgwId)
		// }

		// Get all the route tables

		for _, tgw := range tgws {
			inputTgwRouteTable := awsrouter.TgwRouteTableInputFilter([]string{tgw.TgwId})
			resultTgwRouteTable, err := awsrouter.GetTgwRouteTables(context.TODO(), client, inputTgwRouteTable)
			if err != nil {
				fmt.Println("Error getting the route tables for the TGW: ", tgw.TgwId)
				fmt.Println("Error: ", err)
				os.Exit(1)
			}
			for _, tgwRouteTable := range resultTgwRouteTable.TransitGatewayRouteTables {
				newTgwRouteTable := awsrouter.TgwRouteTable{
					TgwRouteTableId: *tgwRouteTable.TransitGatewayRouteTableId,
					TwgRouteTable:   tgwRouteTable,
				}
				tgw.TgwRouteTables = append(tgw.TgwRouteTables, &newTgwRouteTable)
			}
		}

		
		for _, tgw := range tgws {
			tgw.GetTgwRoutes(context.TODO(), client)
		}
		for _, tgw := range tgws {
			fmt.Printf("The TGW id: %v has the routes\n", tgw.TgwId)
			for _, tgwRouteTable := range tgw.TgwRouteTables {
				fmt.Println("\tThe route table id:", tgwRouteTable.TgwRouteTableId)
				for _, tgwRoute := range tgwRouteTable.TgwRoutes {
					fmt.Println("\t\tThe route:", *tgwRoute.DestinationCidrBlock)
				}
			}
		}

	},

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-aws-routing.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".go-aws-routing" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".go-aws-routing")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
