package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayMetadataKeyUserAgent = "grpcgateway-user-agent"

	grpcClientMetadataKeyUserAgent = "user-agent"

	grpcMetadataKeyClientIp = "x-forwarded-for"
)

type MetaData struct {
	UserAgent string
	ClientIp  string
}

// extractMetaData extracts metadata from the incoming context.
func (server *Server) extractMetaData(ctx context.Context) *MetaData {
	mtd := &MetaData{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// Iterate through the metadata keys.
		for key := range md {
			switch key {
			case grpcGatewayMetadataKeyUserAgent:
				// Get the first value for the "grpc-gateway-user-agent" key.
				if userAgents := md[key]; len(userAgents) > 0 {
					mtd.UserAgent = userAgents[0]
				}
			case grpcClientMetadataKeyUserAgent:
				// Get the first value for the "user-agent" key.
				if userAgents := md[key]; len(userAgents) > 0 {
					mtd.UserAgent = userAgents[0]
				}
			case grpcMetadataKeyClientIp:
				// Get the first value for the "x-forwarded-for" key.
				if clientIps := md[key]; len(clientIps) > 0 {
					mtd.ClientIp = clientIps[0]
				}
			}
		}
	}

	// Check if the peer information is available in the context.
	if p, ok := peer.FromContext(ctx); ok {
		mtd.ClientIp = p.Addr.String()
	}

	return mtd
}
