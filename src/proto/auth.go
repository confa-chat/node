package proto

import (
	"context"
	"flag"
	"strings"

	"github.com/royalcat/konfa-server/pkg/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

var port = flag.Int("port", 50051, "the port to serve on")

// func main() {
// 	flag.Parse()
// 	fmt.Printf("server starting on port %d...\n", *port)

// 	cert, err := tls.LoadX509KeyPair(data.Path("x509/server_cert.pem"), data.Path("x509/server_key.pem"))
// 	if err != nil {
// 		log.Fatalf("failed to load key pair: %s", err)
// 	}
// 	opts := []grpc.ServerOption{
// 		// The following grpc.ServerOption adds an interceptor for all unary
// 		// RPCs. To configure an interceptor for streaming RPCs, see:
// 		// https://godoc.org/google.golang.org/grpc#StreamInterceptor
// 		grpc.UnaryInterceptor(ensureValidToken),
// 		// Enable TLS for all incoming connections.
// 		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
// 	}
// 	s := grpc.NewServer(opts...)
// 	pb.RegisterEchoServer(s, &ecServer{})
// 	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}
// 	if err := s.Serve(lis); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// }

// valid validates the authorization.
func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	// Perform the token validation here. For the sake of this example, the code
	// here forgoes any of the usual OAuth2 token validation and instead checks
	// for a token matching an arbitrary string.
	return token == "some-secret-token"
}

type ctxKey string

const ctxUserKey ctxKey = "user"

// Authenticate ensures a valid token exists within a request's metadata. If
// the token is missing or invalid, the interceptor blocks execution of the
// handler and returns an error. Otherwise, the interceptor invokes the unary
// handler.
func Authenticate(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// md, ok := metadata.FromIncomingContext(ctx)
	// if !ok {
	// 	return nil, errMissingMetadata
	// }
	// // The keys within metadata.MD are normalized to lowercase.
	// // See: https://godoc.org/google.golang.org/grpc/metadata#New
	// if !valid(md["authorization"]) {
	// 	return nil, errInvalidToken
	// }

	var user User

	user = User{
		ID: uuid.MustFromString("a903b474-26f4-4262-9ba7-97edaa76491f"),
	}

	ctx = context.WithValue(ctx, ctxUserKey, &user)

	// Continue execution of handler after ensuring a valid token.
	return handler(ctx, req)
}

type User struct {
	ID uuid.UUID
}

func getCtxUser(ctx context.Context) *User {
	return ctx.Value(ctxUserKey).(*User)
}
