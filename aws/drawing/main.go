package main

import (
	"fmt"
	"image/color"

	"gitlab.presidio.com/rgomez/aws-router/aws/awsrouter"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

const (
	//route height
	routeHeight = 20
	//route width
	routeWidth = 20
	//route margin
	routeMargin = 50
	// tgw width
	tgwWidth = 2800
	// tgw height
	tgwHeight = 800
)

func main() {
	tgw := awsrouter.Tgw{
		ID:   "tgw-12345",
		Name: "tgw-12345",
		RouteTables: []*awsrouter.TgwRouteTable{
			{
				ID:   "rt-12345",
				Name: "rt-12345",
			},
			{
				ID:   "rt-123456",
				Name: "rt-123456",
			},
			{
				ID:   "rt-1234567",
				Name: "rt-1234567",
			},
			{
				ID:   "rt-12345678",
				Name: "rt-12345678",
			},
		},
	}
	//fmt.Println(tgw)
	// find the number of route tables on the TGW.
	// number of route tables
	numRt := len(tgw.RouteTables)

	// find the ratio of the width of the TGW to the number of route tables
	// ratio of width to number of route tables
	// find the size of the image
	// image width
	imgWidth := (routeWidth * numRt) + (routeMargin * (numRt + 1))
	fmt.Println("imgWidth:", imgWidth)
	// // image height
	//imgHeight := routeHeight + 2*routeMargin

	dc := gg.NewContext(3000, 1000)
	c := color.Color(color.White)
	dc.SetColor(c)
	// Draw a rectangle
	dc.Clear()

	// set the font
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic("")
	}
	face := truetype.NewFace(font, &truetype.Options{
		Size: 40,
	})
	dc.SetFontFace(face)
	
	// Draw a TGW
	// color for the rectangle
	dc.DrawRectangle(100, 100, float64(tgwWidth), float64(tgwHeight))
	dc.SetColor(color.Black)
	dc.Stroke()
	dc.DrawString(tgw.Name, 100, 100)
	// Save the image to a file

	// find width for each route table
	// width of each route table
	rtWidth := (tgwWidth - (numRt+1)*routeMargin) / numRt
	// find height for each route table
	// height of each route table

	rtHeight := (tgwHeight - 2*routeMargin) / numRt
	for i, rt := range tgw.RouteTables {
		dc.Push()
		x := 100 + routeMargin + i*(rtWidth+routeMargin)
		y := 100 + routeMargin
		// dc.Scale(float64(routeWidth), float64(routeHeight))
		dc.DrawRectangle(float64(x), float64(y), float64(rtWidth), float64(rtHeight))
		dc.SetColor(color.Black)
		dc.Stroke()
		dc.DrawString(rt.Name, float64(x), float64(y))
		dc.Pop()
	}
	dc.SavePNG("out.png")
}
