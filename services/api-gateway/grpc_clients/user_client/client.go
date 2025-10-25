package user_client

import (
	"context"

	"github.com/wutthichod/sa-connext/shared/config"
	pb "github.com/wutthichod/sa-connext/shared/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserServiceClient struct {
	Client pb.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserServiceClient(config config.Config) (*UserServiceClient, error) {

	conn, err := grpc.NewClient(config.App().User, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewUserServiceClient(conn)
	return &UserServiceClient{
		Client: client,
		conn:   conn,
	}, nil
}

func (c *UserServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}

func (s *UserServiceClient) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return s.Client.CreateUser(ctx, req)
}
