package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func (a *Authenticator) isNeedAuth(method string) bool {
	for _, noAuthMethod := range a.skipAuthMethods {
		if strings.HasPrefix(method, noAuthMethod) {
			return false
		}
	}
	return true
}

func (a *Authenticator) UnaryAuthenticate(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if !a.isNeedAuth(info.FullMethod) {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		a.logger.Warn("missing metadata")
		return nil, errMissingMetadata
	}

	token := grpcExtractToken(md["authorization"])
	if token == "" {
		a.logger.Warn("missing token in metadata")
		return nil, errInvalidToken
	}

	ctx, err := a.authorize(ctx, token)
	if err != nil {
		a.logger.Warn("failed to authorize token", "error", err)
		return nil, err
	}

	return handler(ctx, req)
}

func (a *Authenticator) StreamAuthenticate(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !a.isNeedAuth(info.FullMethod) {
		return handler(srv, ss)
	}

	ctx := ss.Context()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errMissingMetadata
	}

	token := grpcExtractToken(md["authorization"])
	if token == "" {
		return errInvalidToken
	}

	ctx, err := a.authorize(ctx, token)
	if err != nil {
		return err
	}

	return handler(srv, newWrappedStream(ctx, ss))
}

func grpcExtractToken(authorization []string) string {
	if len(authorization) < 1 {
		return ""
	}

	return strings.TrimPrefix(authorization[0], "Bearer ")
}

type wrappedStreamContext struct {
	ctx context.Context
	grpc.ServerStream
}

func newWrappedStream(ctx context.Context, s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStreamContext{ctx: ctx, ServerStream: s}
}
