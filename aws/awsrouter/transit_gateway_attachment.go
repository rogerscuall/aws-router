package awsrouter

import (
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// TgwAttachments holds the data of a Transit Gateway Attachment.
type TgwAttachment struct {
	ID         string
	ResourceID string
	Type       string
}

// newTgwAttach builds a TgwAttachment from a aws TransitGatewayRouteAttachment type.
func newTgwAttachment(att types.TransitGatewayRouteAttachment) *TgwAttachment {
	attType := fmt.Sprint(att.ResourceType)
	return &TgwAttachment{
		ID:         *att.TransitGatewayAttachmentId,
		ResourceID: *att.ResourceId,
		Type:       attType,
	}
}

// GetAttachmentsFromTgwRoute returns a list of TgwAttachments from a aws TransitGatewayRoute type.
func GetAttachmentsFromTgwRoute(route types.TransitGatewayRoute) []*TgwAttachment {
	if len(route.TransitGatewayAttachments) == 0 {
		return nil
	}
	var results []*TgwAttachment
	for _, attachment := range route.TransitGatewayAttachments {
		att := newTgwAttachment(attachment)
		results = append(results, att)
	}
	return results
}

// GetDirectlyConnectedAttachmentFromTgwRoute returns the TGW Attachment that is most likely to be directly connected.
// The rts is a list of TgwRouteTable with a single route prefix (best route prefix), basically the output of FilterRouteTableRoutesPerPrefix.
// The Route Tables in rts should have only one route, that is the most specific route to a destination.
func GetDirectlyConnectedAttachmentFromTgwRoute(rts []TgwRouteTable) []*TgwAttachment {
	var results []*TgwAttachment
	for _, rt := range rts {
		r := rt.Routes[0]
		switch r.Type {
		case "propagated":
			return GetAttachmentsFromTgwRoute(r)
		case "static":
			fmt.Println("Static route not implemented")
		default:
			fmt.Println("Default case not implemented")
		}
	}
	return results
}

type AttPath struct {
	Path          []*TgwAttachment
	SrcRouteTable TgwRouteTable
	DstRouteTable TgwRouteTable
	Tgw 		 *Tgw
}

// Walk will do a packet walk from the src to dst and updates the field Path.
// The function needs a attPath that has at least the source attachment.
func (attPath *AttPath) Walk(src, dst net.IP) error {
	srcRt, srcAtts, err := attPath.Tgw.GetDirectlyConnectedAttachment(src)
	if err != nil {
		return err
	}
	dstRt, dstAtts, err := attPath.Tgw.GetDirectlyConnectedAttachment(dst)
	if err != nil {
		return err
	}
	attPath.Path = append(attPath.Path, srcAtts[0])
	attPath.SrcRouteTable = srcRt
	attPath.DstRouteTable = dstRt
	tgwRt := srcRt
	for i := 0; i < 10; i++ {
		route, err := tgwRt.BestRouteToIP(dst)
		if err != nil {
			return err
		}
		if route.DestinationCidrBlock == nil {
			return fmt.Errorf("No route found available to walk")
		}
		att := newTgwAttachment(route.TransitGatewayAttachments[0])
		attPath.Path = append(attPath.Path, att)
		if att.ID == dstAtts[0].ResourceID {
			// We reach the destination attachment
			break
		}
	}
	return nil
}
