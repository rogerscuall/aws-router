package awsrouter

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func Test_newTgwAttachment(t *testing.T) {
	type args struct {
		att types.TransitGatewayRouteAttachment
	}
	tests := []struct {
		name string
		args args
		want *TgwAttachment
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newTgwAttachment(tt.args.att); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTgwAttachment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAttachmentsFromTgwRoute(t *testing.T) {
	type args struct {
		route types.TransitGatewayRoute
	}
	tests := []struct {
		name string
		args args
		want []*TgwAttachment
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAttachmentsFromTgwRoute(tt.args.route); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAttachmentsFromTgwRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDirectlyConnectedAttachmentFromTgwRoute(t *testing.T) {
	type args struct {
		rts []TgwRouteTable
	}
	tests := []struct {
		name string
		args args
		want []*TgwAttachment
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDirectlyConnectedAttachmentFromTgwRoute(tt.args.rts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDirectlyConnectedAttachmentFromTgwRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAttPath_isAttachmentInPath(t *testing.T) {
	type fields struct {
		Path          []*TgwAttachment
		MapPath       map[string]struct{}
		SrcRouteTable TgwRouteTable
		DstRouteTable TgwRouteTable
		Tgw           *Tgw
	}
	type args struct {
		ID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attPath := AttPath{
				Path:          tt.fields.Path,
				MapPath:       tt.fields.MapPath,
				SrcRouteTable: tt.fields.SrcRouteTable,
				DstRouteTable: tt.fields.DstRouteTable,
				Tgw:           tt.fields.Tgw,
			}
			if got := attPath.isAttachmentInPath(tt.args.ID); got != tt.want {
				t.Errorf("AttPath.isAttachmentInPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAttPath_addAttachmentToPath(t *testing.T) {
	type fields struct {
		Path    []*TgwAttachment
		MapPath map[string]struct{}
		// SrcRouteTable TgwRouteTable
		// DstRouteTable TgwRouteTable
		// Tgw           *Tgw
	}
	type args struct {
		att *TgwAttachment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "UniquePath", fields: fields{
			Path: []*TgwAttachment{
				{ID: "1234"},
				{ID: "5678"},
			},
			MapPath: map[string]struct{}{
				"1234": {},
				"5678": {},
			},
		},
			args:    args{att: &TgwAttachment{ID: "9012"}},
			wantErr: false,
		},
		{name: "DuplicatePath", fields: fields{
			Path: []*TgwAttachment{
				{ID: "1234"},
				{ID: "5678"},
			},
			MapPath: map[string]struct{}{
				"1234": {},
			},
		},
			args:    args{att: &TgwAttachment{ID: "1234"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attPath := AttPath{
				Path:    tt.fields.Path,
				MapPath: tt.fields.MapPath,
			}
			if err := attPath.addAttachmentToPath(tt.args.att); (err != nil) != tt.wantErr {
				t.Errorf("AttPath.addAttachmentToPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAttPath_Walk(t *testing.T) {
	type fields struct {
		Path          []*TgwAttachment
		MapPath       map[string]struct{}
		SrcRouteTable TgwRouteTable
		DstRouteTable TgwRouteTable
		Tgw           *Tgw
	}
	type args struct {
		ctx context.Context
		api AwsRouter
		src net.IP
		dst net.IP
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attPath := &AttPath{
				Path:          tt.fields.Path,
				MapPath:       tt.fields.MapPath,
				SrcRouteTable: tt.fields.SrcRouteTable,
				DstRouteTable: tt.fields.DstRouteTable,
				Tgw:           tt.fields.Tgw,
			}
			if err := attPath.Walk(tt.args.ctx, tt.args.api, tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("AttPath.Walk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAttPath_String(t *testing.T) {
	type fields struct {
		Path          []*TgwAttachment
		MapPath       map[string]struct{}
		SrcRouteTable TgwRouteTable
		DstRouteTable TgwRouteTable
		Tgw           *Tgw
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attPath := AttPath{
				Path:          tt.fields.Path,
				MapPath:       tt.fields.MapPath,
				SrcRouteTable: tt.fields.SrcRouteTable,
				DstRouteTable: tt.fields.DstRouteTable,
				Tgw:           tt.fields.Tgw,
			}
			if got := attPath.String(); got != tt.want {
				t.Errorf("AttPath.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
