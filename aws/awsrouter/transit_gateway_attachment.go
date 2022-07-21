package awsrouter

import (
	"context"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
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

// getAttachmentsFromTgwRoute returns a list of TgwAttachments from a aws TransitGatewayRoute type.
func getAttachmentsFromTgwRoute(route types.TransitGatewayRoute) []*TgwAttachment {
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

// getDirectlyConnectedAttachmentFromTgwRoute returns the TGW Attachment that is most likely to be directly connected.
// The rts is a list of TgwRouteTable with a single route prefix (best route prefix), basically the output of FilterRouteTableRoutesPerPrefix.
// The Route Tables in rts should have only one route, that is the most specific route to a destination.
func getDirectlyConnectedAttachmentFromTgwRoute(rts []TgwRouteTable) []*TgwAttachment {
	var results []*TgwAttachment
	for _, rt := range rts {
		r := rt.Routes[0]
		switch r.Type {
		case "propagated":
			return getAttachmentsFromTgwRoute(r)
		case "static":
			fmt.Println("Static route not implemented")
		default:
			fmt.Println("Default case not implemented")
		}
	}
	return results
}

// AttPath is a list of TgwAttachments that represent the path from a source to a destination.
// The first element is the source attachment, the last element is the destination attachment.
// No two elements are the same.
type AttPath struct {
	Path          []*TgwAttachment
	MapPath       map[string]struct{}
	SrcRouteTable TgwRouteTable
	DstRouteTable TgwRouteTable
	Tgw           *Tgw
}

// NewAttPath builds a AttPath from a aws TransitGatewayRoute type.
func NewAttPath() *AttPath {
	return &AttPath{
		Path:          make([]*TgwAttachment, 0),
		MapPath:       make(map[string]struct{}),
		SrcRouteTable: TgwRouteTable{},
		DstRouteTable: TgwRouteTable{},
		Tgw:           &Tgw{},
	}
}

// isAttachmentInPath returns true if the attachment is in the path.
func (attPath AttPath) isAttachmentInPath(ID string) bool {
	_, ok := attPath.MapPath[ID]
	if ok {
		return true
	}
	return false
}

// addAttachmentToPath adds an attachment to the path.
// The attachment is added only if it is not already in the path.
// If the attachment is already in the path it will throw an error.
func (attPath *AttPath) addAttachmentToPath(att *TgwAttachment) error {
	if attPath.isAttachmentInPath(att.ID) {
		return ErrTgwAttachmetInPath
	}
	attPath.Path = append(attPath.Path, att)
	attPath.MapPath[att.ID] = struct{}{}
	return nil
}

// Walk will do a packet walk from the src to dst and updates the field Path.
// The function will walk from one attachment to the next, until it reaches the dst.
// There is a limit of 10 hops. If the limit is reached, the function will return an error.
// TODO: allow the option to increase the depth of the walk, right now is 10.
func (attPath *AttPath) Walk(ctx context.Context, api AwsRouter, src, dst net.IP) error {
	srcRt, srcAtts, err := attPath.Tgw.GetDirectlyConnectedAttachment(src)
	if err != nil {
		return err
	}
	attPath.addAttachmentToPath(srcAtts[0])
	attPath.SrcRouteTable = srcRt
	tgwRt := &srcRt
	for i := 0; i < 10; i++ {
		route, err := tgwRt.BestRouteToIP(dst)
		if err != nil {
			return err
		}
		if route.DestinationCidrBlock == nil {
			return ErrTgwRouteTableRouteNotFound
		}
		nextHopAtt := newTgwAttachment(route.TransitGatewayAttachments[0])

		// Check if the next hop is already the last attachment in the path.
		// If the nextHopAtt is the last attachment in the path, then we have reached the destination.
		// This is because the BestRouteToIP in the current Route Table will will send the packet to
		// an attachment that is directly connected to the destination. Traffic entering from this attachment
		// will match the same route and will be sent back to the same attachment.
		// The best way to avoid this check is verifying if the resource after the attachment owns the CIDR block
		// for the destination.
		if len(attPath.Path) > 0 && nextHopAtt.ID == attPath.Path[len(attPath.Path)-1].ID {
			break
		}

		// Add the next hop to the path
		err = attPath.addAttachmentToPath(nextHopAtt)
		if err != nil {
			return fmt.Errorf("Attachment %s is already in the path", nextHopAtt.ID)
		}

		// Find the route table associated to the attachment
		// Create a filter for TgwAttachmentInputFilter
		filter := types.Filter{
			Name:   aws.String("resource-id"),
			Values: []string{nextHopAtt.ResourceID},
		}
		// Create a filter of type TgwAttachmentInputFilter
		input := TgwAttachmentInputFilter(filter)
		// Get the list of TgwRouteTable that match the filter
		output, err := TgwGetAttachments(ctx, api, input)
		if err != nil {
			return err
		}
		if len(output.TransitGatewayAttachments) != 1 {
			return ErrTgwRouteTableNotFound
		}
		routeTableID := *output.TransitGatewayAttachments[0].Association.TransitGatewayRouteTableId
		if routeTableID == tgwRt.ID {
			// We reach the destination attachment

			break
		}
		tgwRt, err = attPath.Tgw.GetTgwRouteTableByID(routeTableID)
	}
	return nil
}

// String for a AttPath returns a string with the path.
func (attPath AttPath) String() string {
	var result string
	for _, att := range attPath.Path {
		result += fmt.Sprintf("%s\n", att.ID)
	}
	return result
}
