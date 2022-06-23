package awsrouter

import (
	"context"
	"net"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
)

var listOfRouteTables = []*TgwRouteTable{
	{
		ID:   "rtb-0d7f9b0a",
		Name: "rtb1",
		Data: listDescribeTransitGatewayRouteTablesOutput.TransitGatewayRouteTables[0],
		Routes: []types.TransitGatewayRoute{
			{
				DestinationCidrBlock: aws.String("10.0.1.0/24"),
				TransitGatewayAttachments: []types.TransitGatewayRouteAttachment{
					{
						ResourceId:   aws.String("tgw-0d7f9b0x"),
						ResourceType: "vpc",
					},
				},
				Type: "propagated",
			},
		},
	},
	{
		ID:   "rtb-0d7f9b0b",
		Name: "rtb2",
		Data: listDescribeTransitGatewayRouteTablesOutput.TransitGatewayRouteTables[1],
		Routes: []types.TransitGatewayRoute{
			{
				DestinationCidrBlock: aws.String("10.0.0.0/16"),
				TransitGatewayAttachments: []types.TransitGatewayRouteAttachment{
					{
						ResourceId:   aws.String("tgw-0d7f9b0x"),
						ResourceType: "vpc",
					},
				},
				Type: "propagated",
			},
		},
	},
	{
		ID:   "rtb-0d7f9b0c",
		Name: "rtb3",
		Data: listDescribeTransitGatewayRouteTablesOutput.TransitGatewayRouteTables[2],
		Routes: []types.TransitGatewayRoute{
			{
				DestinationCidrBlock: aws.String("0.0.0.0/0"),
				TransitGatewayAttachments: []types.TransitGatewayRouteAttachment{
					{
						ResourceId:   aws.String("tgw-0d7f9b0x"),
						ResourceType: "vpc",
					},
				},
				Type: "static",
			},
		},
	},
}

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

func TestTgwRouteTableSelectionPriority(t *testing.T) {
	type args struct {
		rts []*TgwRouteTable
		src net.IP
	}
	tests := []struct {
		name    string
		args    args
		want    *TgwRouteTable
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TgwRouteTableSelectionPriority(tt.args.rts, tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("TgwRouteTableSelectionPriority() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TgwRouteTableSelectionPriority() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestfindBestRoutePrefix(t *testing.T) {
	net10 := net.ParseIP("10.0.1.10")
	net192 := net.ParseIP("192.8.1.1")
	_, sub10, _ := net.ParseCIDR("10.0.1.0/24")
	_, subdefault, _ := net.ParseCIDR("0.0.0.0/0")
	type args struct {
		rts []*TgwRouteTable
		src net.IP
	}
	tests := []struct {
		name    string
		args    args
		want    net.IPNet
		wantErr bool
	}{
		{
			name: "No Routes",
			args: args{
				rts: []*TgwRouteTable{},
				src: net.IP{},
			},
			want:    net.IPNet{},
			wantErr: false,
		},
		{
			name: "The Most Specific Route",
			args: args{
				rts: listOfRouteTables,
				src: net10,
			},
			want:    *sub10,
			wantErr: false,
		},
		{
			name: "Default Route",
			args: args{
				rts: listOfRouteTables,
				src: net192,
			},
			want:    *subdefault,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findBestRoutePrefix(tt.args.rts, tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindBestRoutePrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindBestRoutePrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTgwRouteTable_UpdateAttachments(t *testing.T) {
	type fields struct {
		ID          string
		Name        string
		Data        types.TransitGatewayRouteTable
		Routes      []types.TransitGatewayRoute
		Attachments []TgwAttachment
	}
	type args struct {
		ctx context.Context
		api AwsRouter
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "No Attachments",
			fields: fields{
					ID:   "rtb-0d7f9b0a",
					Name: "rtb1",
					Data: listDescribeTransitGatewayRouteTablesOutput.TransitGatewayRouteTables[0],
					Routes: []types.TransitGatewayRoute{
						{
							DestinationCidrBlock: aws.String("10.0.1.0/24"),
							TransitGatewayAttachments: []types.TransitGatewayRouteAttachment{
								{
									ResourceId:   aws.String("tgw-0d7f9b0x"),
									ResourceType: "vpc",
								},
							},
							Type: "propagated",
						},
				},
			},
			args: args{
				context.TODO(),
				TgwDescriberImpl{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TgwRouteTable{
				ID:          tt.fields.ID,
				Name:        tt.fields.Name,
				Data:        tt.fields.Data,
				Routes:      tt.fields.Routes,
				Attachments: tt.fields.Attachments,
			}
			if err := tr.UpdateAttachments(tt.args.ctx, tt.args.api); (err != nil) != tt.wantErr {
				t.Errorf("TgwRouteTable.UpdateAttachments() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(tr.Attachments) < 1 {
				t.Errorf("Number of attachment is less than 1 %v, wantErr %v",len(tr.Attachments) , tt.wantErr)
			}
		})
	}
}

func Test_findBestRoutePrefix(t *testing.T) {
	type args struct {
		rts    []*TgwRouteTable
		ipAddr net.IP
	}
	tests := []struct {
		name    string
		args    args
		want    net.IPNet
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findBestRoutePrefix(tt.args.rts, tt.args.ipAddr)
			if (err != nil) != tt.wantErr {
				t.Errorf("findBestRoutePrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findBestRoutePrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterRouteTableRoutesPerPrefix(t *testing.T) {
	type args struct {
		rts    []*TgwRouteTable
		prefix net.IPNet
	}
	tests := []struct {
		name    string
		args    args
		want    []TgwRouteTable
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterRouteTableRoutesPerPrefix(tt.args.rts, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterRouteTableRoutesPerPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterRouteTableRoutesPerPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
