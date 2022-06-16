package draw

import (
	"io/fs"
	"os"
	"reflect"
	"testing"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"gitlab.presidio.com/rgomez/aws-router/aws/awsrouter"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
)

var tgw = awsrouter.Tgw{
	Name: "TestDrawTgw",
	ID:   "TestDrawTgw",
	RouteTables: []*awsrouter.TgwRouteTable{
		{
			Name: "TestDrawTgw1",
			ID:   "tgw-rtb-042e9098a139d16f3",
		},
		{
			Name: "TestDrawTgw2",
			ID:   "tgw-rtb-042e9098a139d16f4",
		},
		{
			Name: "TestDrawTgw3",
			ID:   "tgw-rtb-042e9098a139d16f5",
		},
	},
}

func TestDrawTgwRouteTable(t *testing.T) {
	fontt, _ := truetype.Parse(goregular.TTF)
	face := truetype.NewFace(fontt, &truetype.Options{Size: 12})
	type args struct {
		dc   *gg.Context
		x    float64
		y    float64
		name string
		face font.Face
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TestDrawTgwRouteTable",
			args: args{
				dc:   gg.NewContext(tgwWidth, tgwHeight),
				x:    0,
				y:    0,
				name: "TestDrawTgwRouteTable",
				face: face,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DrawTgwRouteTable(tt.args.dc, tt.args.x, tt.args.y, tt.args.name, tt.args.face)
		})
	}
}

func TestDrawTgw(t *testing.T) {
	type args struct {
		dc  *gg.Context
		x   float64
		y   float64
		tgw awsrouter.Tgw
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TestDrawTgw",
			args: args{
				dc:  gg.NewContext(tgwWidth, tgwHeight),
				x:   0,
				y:   0,
				tgw: tgw,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DrawTgw(tt.args.dc, tt.args.x, tt.args.y, tt.args.tgw)
		})
	}
}

func TestCreateTgwContext(t *testing.T) {
	type args struct {
		tgw awsrouter.Tgw
	}
	tests := []struct {
		name string
		args args
		want *gg.Context
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateTgwContext(tt.args.tgw); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateTgwContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDrawTgwFull(t *testing.T) {
	folder, _ := os.Stat("testdata")
	type args struct {
		tgw    awsrouter.Tgw
		folder fs.FileInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "NoFolder",
			args: args{
				tgw:    tgw,
				folder: nil,
			},
			wantErr: true,
		},
		{
			name: "NoRouteTables",
			args: args{
				tgw: awsrouter.Tgw{
					Name:        "TestDrawTgw",
					ID:          "TestDrawTgw",
					RouteTables: []*awsrouter.TgwRouteTable{},
				},
				folder: nil,
			},
			wantErr: true,
		},
		{
			name: "Working",
			args: args{
				tgw:    tgw,
				folder: folder,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DrawTgwFull(tt.args.tgw, tt.args.folder); (err != nil) != tt.wantErr {
				t.Errorf("DrawTgwFull() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
