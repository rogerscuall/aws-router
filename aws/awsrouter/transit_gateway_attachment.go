package awsrouter

import (
	"context"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	"gitlab.presidio.com/rgomez/aws-router/ports"
)

// TgwAttachments holds the data of a Transit Gateway Attachment.
type TgwAttachment struct {
	// The ID of the attachment.
	ID string

	// The ID of the resource where this attachment terminates.
	ResourceID string

	// The type of the resource where this attachment terminates.
	// Common values are: vpc, vpn, direct-connect ...
	Type string
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
// This is a helper function that takes in a routes (AWS type) and returns a list of TgwAttachments.
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
// There can be 2 or more attachments in the path, but 2 or 3 are common values.
// No two elements are the same, because that would be a loop.
type AttPath struct {
	// The list of attachments in the path, from source to destination.
	Path []*TgwAttachment

	// A map of the attachments in the path, to avoid duplicates.
	mapPath map[string]struct{}

	// The source route table.
	SrcRouteTable TgwRouteTable

	// The destination route table.
	DstRouteTable TgwRouteTable

	// The Transit Gateway of this path.
	Tgw *Tgw
}

// NewAttPath builds a AttPath.
func NewAttPath() *AttPath {
	return &AttPath{
		Path:          make([]*TgwAttachment, 0),
		mapPath:       make(map[string]struct{}),
		SrcRouteTable: TgwRouteTable{},
		DstRouteTable: TgwRouteTable{},
		Tgw:           &Tgw{},
	}
}

// isAttachmentInPath returns true if the attachment is in the path.
func (attPath AttPath) isAttachmentInPath(ID string) bool {
	_, ok := attPath.mapPath[ID]
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
	attPath.mapPath[att.ID] = struct{}{}
	return nil
}

// Walk will do a packet walk from the src to dst and updates the field Path.
// The function will walk from one attachment to the next, until it reaches the dst.
// There is a limit of 10 hops. If the limit is reached, the function will return an error.
// TODO: allow the option to increase the depth of the walk, right now is 10.
func (attPath *AttPath) Walk(ctx context.Context, api ports.AWSRouter, src, dst net.IP) error {
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
		input := ports.TgwAttachmentInputFilter(filter)
		// Get the list of TgwRouteTable that match the filter
		output, err := ports.TgwGetAttachments(ctx, api, input)
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
	for i := 0; i < len(attPath.Path); i++ {
		result += attPath.Path[i].ID
		if i < len(attPath.Path)-1 {
			result += " -> "
		}
	}
	return result
}
