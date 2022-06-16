package draw

import (
	"fmt"
	"image/color"
	"io/fs"

	"gitlab.presidio.com/rgomez/aws-router/aws/awsrouter"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

// type Router interface {
// 	UpdateRouteTables()
// }

const (
	//route height
	routeHeight = 100
	//route width
	routeWidth = 300
	//route margin
	routeMargin = 50
	// tgw width
	tgwWidth = 2800
	// tgw height
	tgwHeight     = 200
	tgwRtIdLength = 26
)

// DrawTgwRouteTable draws a TGW route table
func DrawTgwRouteTable(dc *gg.Context, x, y float64, name string, face font.Face) {
	// Draw the rectangle of teh route table
	dc.DrawRectangle(x, y, float64(routeWidth), float64(routeHeight))
	dc.SetColor(color.Black)
	dc.SetFontFace(face)
	dc.DrawStringAnchored(name, float64(x)+float64(routeWidth)*0.5, float64(y)+0.5*float64(routeHeight), 0.5, 0.5)
	dc.Stroke()
}

// DrawTgw draws a TGW
func DrawTgw(dc *gg.Context, x, y float64, tgw awsrouter.Tgw) {
	// find the number of route tables
	numRt := len(tgw.RouteTables)

	// find the width of the TGW
	tgwWidth := numRt*(routeWidth+routeMargin) + routeMargin
	// Draw the rectangle of teh route table
	dc.DrawRectangle(x, y, float64(tgwWidth), float64(tgwHeight))
	dc.SetColor(color.Black)
	dc.Stroke()
	dc.DrawString(tgw.Name, x, y)
}

// CreateTgwRouteTable creates a context for an AWS TGW object
func CreateTgwContext(tgw awsrouter.Tgw) *gg.Context {
	// find the number of route tables
	numRt := len(tgw.RouteTables)

	// find the width of the TGW
	tgwWidth := numRt*(routeWidth+routeMargin) + routeMargin

	dc := gg.NewContext(tgwWidth+200, 1000)
	c := color.Color(color.White)
	dc.SetColor(c)
	// Draw a rectangle
	dc.Clear()
	return dc
}

// DrawTgwFull draws a TGW and all the route tables
func DrawTgwFull(tgw awsrouter.Tgw, folder fs.FileInfo) error {
	// if the TGW has no Route Tables return an error
	if len(tgw.RouteTables) == 0 {
		return fmt.Errorf("TGW has no Route Tables")
	}
	// if folder is nil return an error or  the folder does not exist return an error
	if folder == nil || !folder.IsDir() {
		return fmt.Errorf("folder is nil")
	}
	
	// initial x and y coordinates
	x, y := float64(100), float64(100)
	dc := CreateTgwContext(tgw)
	// set the font
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return err
	}
	face := truetype.NewFace(font, &truetype.Options{
		Size: 40,
	})
	dc.SetFontFace(face)
	DrawTgw(dc, x, y, tgw)
	face = truetype.NewFace(font, &truetype.Options{
		Size: 15,
	})
	dc.SetFontFace(face)
	for i, rt := range tgw.RouteTables {
		dc.Push()
		x := float64(100 + routeMargin + i*(routeWidth+routeMargin))
		y := float64(100 + routeMargin)
		DrawTgwRouteTable(dc, x, y, rt.Name, face)
		dc.Pop()
	}
	fileName := fmt.Sprintf("%s/%s.png", folder.Name(), tgw.Name)
	dc.SavePNG(fileName)
	return nil
}
