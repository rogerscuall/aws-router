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

	"github.com/rogerscuall/aws-router/aws/awsrouter"
	"github.com/spf13/cobra"
)

// excelCmd represents the excel command
var excelCmd = &cobra.Command{
	Use:   "excel",
	Short: "Export all route tables to excel",
	Long: `Each Transit Gateway will have a separate Excel and each route table will have a separate sheet.
By default all excel are stored on the folder excel. The folder has to exist.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		ctx := context.TODO()
		defer func() {
			if err != nil {
				app.ErrorLog.Println(err)
			}
		}()
		fmt.Println("Exporting AWS routing to Excel")
		folderName := "excel"
		tgws, err := app.UpdateRouting(ctx)
		var folder os.FileInfo
		if folder, err = os.Stat(folderName); err != nil {
			err = os.Mkdir(folderName, 0755)
			if err != nil {
				fmt.Println("Error creating folder:", err)
				return // Exit if there's an error
			}
			fmt.Println("Folder created:", folderName)
			folder, _ = os.Stat(folderName)
		}
		err = awsrouter.ExportTgwRoutesExcel(tgws, folder)
		if err != nil {
			fmt.Println(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(excelCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// excelCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// excelCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
