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
	"strings"

	"github.com/rs/xid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.presidio.com/rgomez/aws-router/adapters/db"
	"gitlab.presidio.com/rgomez/aws-router/ports"
)

var dbName string

// processCmd represents the process command
var processCmd = &cobra.Command{
	Use:   "process",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbName = viper.GetString("db_name")
		var dbAdapter ports.DbPort
		var err error
		dbAdapter, err = db.NewAdapter(dbName)
		defer dbAdapter.CloseDbConnection()

		id := xid.New().String()[13:]
		taskName := strings.Join(args, " ")
		task := tasks.New(id, taskName)
		bTask := tasks.ToBytes(task)
		err = dbAdapter.SetVal(id, bTask)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(processCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// processCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// processCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
