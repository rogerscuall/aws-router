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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.presidio.com/rgomez/aws-router/adapters/db"
	"gitlab.presidio.com/rgomez/aws-router/ports"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Extracts all routing information from AWS and save it to a DB for later use",
	Long:  `Once the DB is populated, we can find information about routing`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		defer func() {
			if err != nil {
				cobra.CheckErr(err)
			}
		}()
		dbNamePrefix := viper.GetString("db_name")
		dbNameTgw := fmt.Sprintf("%s_tgw", dbNamePrefix)
		dbNameTgwRouteTable := fmt.Sprintf("%s_tgw_route_table", dbNamePrefix)
		var dbAdapterTgw, dbAdapterTgwRouteTable ports.DbPort
		dbAdapterTgw, err = db.NewAdapter(dbNameTgw)
		dbAdapterTgwRouteTable, err = db.NewAdapter(dbNameTgwRouteTable)
		defer dbAdapterTgw.CloseDbConnection()
		defer dbAdapterTgwRouteTable.CloseDbConnection()
		fmt.Println("Downloading routing information from AWS")
		ctx := context.TODO()
		tgws, err := app.UpdateRouting(ctx)
		fmt.Println("Saving routing information to DB")
		for _, tgw := range tgws {
			err = dbAdapterTgw.SetVal(tgw.ID, tgw.Bytes())
			for _, rt := range tgw.RouteTables {
				err = dbAdapterTgwRouteTable.SetVal(rt.ID, rt.Bytes())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
