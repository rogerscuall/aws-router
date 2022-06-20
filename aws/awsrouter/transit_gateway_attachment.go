package awsrouter

import (
	"fmt"

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
