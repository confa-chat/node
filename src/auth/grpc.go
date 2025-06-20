package auth

import (
	"context"
	"strings"

	"github.com/confa-chat/node/src/proto/confa"
	_ "github.com/confa-chat/node/src/proto/confa"
	"github.com/confa-chat/node/src/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	_ "google.golang.org/protobuf/types/descriptorpb"
)

func getOptionSkipAuth(mdDescriptor protoreflect.MethodDescriptor) bool {
	val := mdDescriptor.Options().ProtoReflect().Get(confa.E_SkipAuth.TypeDescriptor())
	return val.Bool()
}

func (a *Authenticator) isNeedAuth(method string) bool {
	methodFullName := protoreflect.FullName(strings.ReplaceAll(strings.TrimPrefix(method, "/"), "/", "."))
	desc, _ := protoregistry.GlobalFiles.FindDescriptorByName(methodFullName)

	if getOptionSkipAuth(desc.(protoreflect.MethodDescriptor)) {
		return false
	}

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

	user, err := a.authorize(ctx, token)
	if err != nil {
		a.logger.Warn("failed to authorize token", "error", err)
		return nil, err
	}

	ctx = ctxWithUser(ctx, user)

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

	user, err := a.authorize(ctx, token)
	if err != nil {
		return err
	}

	return handler(srv, newWrappedStream(ss, user))
}

func grpcExtractToken(authorization []string) string {
	if len(authorization) < 1 {
		return ""
	}

	return strings.TrimPrefix(authorization[0], "Bearer ")
}

type wrappedStreamContext struct {
	user store.User
	grpc.ServerStream
}

func newWrappedStream(s grpc.ServerStream, user store.User) grpc.ServerStream {
	return &wrappedStreamContext{ServerStream: s, user: user}
}

func (w *wrappedStreamContext) Context() context.Context {
	ctx := w.ServerStream.Context()
	ctx = ctxWithUser(ctx, w.user)
	return ctx
}
