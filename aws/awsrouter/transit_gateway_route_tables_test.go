package awsrouter

import (
	"net"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
)

func TestTgwRouteTable_Bytes(t *testing.T) {
	type fields struct {
		ID   string
		Name string
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		//
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TgwRouteTable{
				ID:   tt.fields.ID,
				Name: tt.fields.Name,
			}
			if got := tr.Bytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TgwRouteTable.Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTgwRouteTable_BestRouteToIP(t *testing.T) {
	type fields struct {
		ID     string
		Name   string
		Data   types.TransitGatewayRouteTable
		Routes []types.TransitGatewayRoute
	}
	type args struct {
		ipAddress net.IP
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    types.TransitGatewayRoute
		wantErr bool
	}{
		{
			name: "Multiple Routes",
			fields: fields{
				ID:     "tgw-rtb-123456789",
				Name:   "tgw-rtb-123456789",
				Data:   types.TransitGatewayRouteTable{},
				Routes: listSearchTransitGatewayRoutesOutput.Routes,
			},
			args: args{
				ipAddress: net.ParseIP("10.0.2.1"),
			},
			want:    listSearchTransitGatewayRoutesOutput.Routes[2],
			wantErr: false,
		},
		{
			name: "No Routes in Route Table",
			fields: fields{
				ID:     "tgw-rtb-123456789",
				Name:   "tgw-rtb-123456789",
				Data:   types.TransitGatewayRouteTable{},
				Routes: []types.TransitGatewayRoute{},
			},
			args: args{
				ipAddress: net.ParseIP("10.0.2.1"),
			},
			want:    types.TransitGatewayRoute{},
			wantErr: false,
		},
		{
			name: "No matching Routes",
			fields: fields{
				ID:     "tgw-rtb-123456789",
				Name:   "tgw-rtb-123456789",
				Data:   types.TransitGatewayRouteTable{},
				Routes: listSearchTransitGatewayRoutesOutput.Routes,
			},
			args: args{
				ipAddress: net.ParseIP("192.8.1.1"),
			},
			want:    types.TransitGatewayRoute{},
			wantErr: false,
		},
		{
			name: "Bad IP Address",
			fields: fields{
				ID:     "tgw-rtb-123456789",
				Name:   "tgw-rtb-123456789",
				Data:   types.TransitGatewayRouteTable{},
				Routes: listSearchTransitGatewayRoutesOutput.Routes,
			},
			args: args{
				ipAddress: net.ParseIP("123"),
			},
			want:    types.TransitGatewayRoute{},
			wantErr: false,
		},
		{
			name: "Bad Destination CIDR",
			fields: fields{
				ID:   "tgw-rtb-123456789",
				Name: "tgw-rtb-123456789",
				Data: types.TransitGatewayRouteTable{},
				Routes: []types.TransitGatewayRoute{
					{
						DestinationCidrBlock: aws.String("123"),
					},
				},
			},
			args: args{
				ipAddress: net.ParseIP("123"),
			},
			want:    types.TransitGatewayRoute{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := TgwRouteTable{
				ID:     tt.fields.ID,
				Name:   tt.fields.Name,
				Data:   tt.fields.Data,
				Routes: tt.fields.Routes,
			}
			got, err := tr.BestRouteToIP(tt.args.ipAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("TgwRouteTable.BestRouteToIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TgwRouteTable.BestRouteToIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newTgwRouteTable(t *testing.T) {
	type args struct {
		t types.TransitGatewayRouteTable
	}
	tests := []struct {
		name string
		args args
		want *TgwRouteTable
	}{
		{
			name: "With Tags",
			args: args{
				t: listDescribeTransitGatewayRouteTablesOutput.TransitGatewayRouteTables[0],
			},
			want: &TgwRouteTable{
				ID:   "rtb-0d7f9b0a",
				Name: "testA",
				Data: listDescribeTransitGatewayRouteTablesOutput.TransitGatewayRouteTables[0],
			},
		},
		{
			name: "Without Tags",
			args: args{
				t: types.TransitGatewayRouteTable{
					TransitGatewayRouteTableId: aws.String("rtb-0d7f9b0d"),
					TransitGatewayId:           aws.String("tgw-0d7f9b0x"),
				},
			},
			want: &TgwRouteTable{
				ID:   "rtb-0d7f9b0d",
				Name: "rtb-0d7f9b0d",
				Data: types.TransitGatewayRouteTable{
					TransitGatewayRouteTableId: aws.String("rtb-0d7f9b0d"),
					TransitGatewayId:           aws.String("tgw-0d7f9b0x"),
				},
			},
		},
		{
			name: "No Transit Gateway ID",
			args: args{
				t: types.TransitGatewayRouteTable{},
			},
			want: &TgwRouteTable{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newTgwRouteTable(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTgwRouteTable() = %v, want %v", got, tt.want)
			}
		})
	}
}
