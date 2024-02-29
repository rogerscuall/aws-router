package awsrouter

import (
	"encoding/csv"
	"fmt"
	"io/fs"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// getNamesFromTags returns the name tags if exist, if not it will signal with an error.
func GetNamesFromTags(tags []types.Tag) (string, error) {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			return *tag.Value, nil
		}
	}
	return "", fmt.Errorf("no name tag found")
}

// ExportTgwRoutesExcel creates a Excel with all the routes in all Tgw Route Tables.
// Each sheet on the Excel is a Tgw Route Table, each route is a route.
func ExportTgwRoutesExcel(tgws []*Tgw, folder fs.FileInfo) error {
	if !folder.IsDir() {
		return fmt.Errorf("folder %s is not a directory", folder.Name())
	}
	folderName := folder.Name()
	for _, tgw := range tgws {
		f := excelize.NewFile()
		for _, tgwRouteTable := range tgw.RouteTables {
			sheet := f.NewSheet(tgwRouteTable.Name)
			for i, route := range tgwRouteTable.Routes {
				// Only for the header
				if i == 0 {
					// TODO: create an optional function to create a header and row
					f.SetCellValue(tgwRouteTable.Name, "A1", "Destination")
					f.SetCellValue(tgwRouteTable.Name, "B1", "State")
					f.SetCellValue(tgwRouteTable.Name, "C1", "RouteType")
					f.SetCellValue(tgwRouteTable.Name, "D1", "PrefixList")
					f.SetCellValue(tgwRouteTable.Name, "E1", "AttachmentName")
				}
				state := fmt.Sprint(route.State)
				routeType := fmt.Sprint(route.Type)
				for _, attachment := range tgwRouteTable.Attachments {
					fmt.Println("attachment id:", attachment.ID)
					fmt.Println("attachment name:", attachment.Name)
				}
				var attachmentName = "-"
				if len(route.TransitGatewayAttachments) != 0 {
					fmt.Println("att len:", len(route.TransitGatewayAttachments))
					attachmentID := fmt.Sprint(*route.TransitGatewayAttachments[0].TransitGatewayAttachmentId)
					attachmentName = tgwRouteTable.GetAttachmentName(attachmentID)
					if attachmentName == "" {
						attachmentName = attachmentID
					}
				}
				var prefixListId string
				if route.PrefixListId == nil {
					prefixListId = "-"
				} else {
					prefixListId = *route.PrefixListId
				}
				row := []string{
					*route.DestinationCidrBlock,
					state,
					routeType,
					prefixListId,
					attachmentName,
				}
				f.SetSheetRow(tgwRouteTable.Name, "A"+fmt.Sprint(i+2), &row)
			}
			f.SetActiveSheet(sheet)
		}
		fileName := fmt.Sprintf("%s/%s.xlsx", folderName, tgw.Name)
		if err := f.SaveAs(fileName); err != nil {
			return fmt.Errorf("error saving excel: %w", err)
		}
	}
	return nil
}

// ExportRouteTableRoutesCsv creates a CSV with all the routes in one Tgw Route Table.
func ExportRouteTableRoutesCsv(w *csv.Writer, tgwrt TgwRouteTable) error {
	defer w.Flush()
	w.Write([]string{"Destination CIDR Block", "State", "Type"})
	for _, route := range tgwrt.Routes {
		state := fmt.Sprint(route.State)
		routeType := fmt.Sprint(route.Type)
		err := w.Write([]string{*route.DestinationCidrBlock, state, routeType})
		if err != nil {
			return fmt.Errorf("error writing to csv: %w", err)
		}
	}
	return nil
}

//
