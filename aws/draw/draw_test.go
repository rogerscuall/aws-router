 package draw

import (
	"io/fs"
	"reflect"
	"testing"

	"github.com/fogleman/gg"
	"gitlab.presidio.com/rgomez/aws-router/aws/awsrouter"
	"golang.org/x/image/font"
)

func TestDrawTgwRouteTable(t *testing.T) {
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateTgwContext(tt.args.tgw); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateTgwContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDrawTgwFull(t *testing.T) {
	type args struct {
		tgw    awsrouter.Tgw
		folder fs.FileInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DrawTgwFull(tt.args.tgw, tt.args.folder); (err != nil) != tt.wantErr {
				t.Errorf("DrawTgwFull() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
