package auth

import (
	"context"
	"fmt"
	"strings"

	oidc "github.com/coreos/go-oidc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns a gRPC interceptor that validates JWTs from Keycloak.
func UnaryServerInterceptor(provider *oidc.Provider, clientID string) (grpc.UnaryServerInterceptor, error) {
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		hhandler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("missing metadata")
		}

		authHeaders := md["authorization"]
		if len(authHeaders) == 0 {
			return nil, fmt.Errorf("authorization token not provided")
		}
		tok := strings.TrimPrefix(authHeaders[0], "Bearer ")
		if _, err := verifier.Verify(ctx, tok); err != nil {
			return nil, fmt.Errorf("invalid token: %w", err)
		}

		return handler(ctx, req)
	}, nil
}

// PerRPCCredentials attaches the Bearer token to outgoing RPCs.
func PerRPCCredentials(token string) grpc.PerRPCCredentials {
	return oauthToken{token: token}
}

type oauthToken struct{ token string }

func (a oauthToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + a.token}, nil
}

func (a oauthToken) RequireTransportSecurity() bool { return false }
