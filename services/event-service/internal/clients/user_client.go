package clients

import (
	"context"

	"github.com/wutthichod/sa-connext/shared/config"
	pb "github.com/wutthichod/sa-connext/shared/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	Client pb.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserClient(config config.Config) (*UserClient, error) {
	conn, err := grpc.NewClient(config.App().User, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &UserClient{Client: pb.NewUserServiceClient(conn), conn: conn}, nil
}

func (c *UserClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}

func (c *UserClient) AddUserToEvent(ctx context.Context, req *pb.AddUserToEventRequest) (*pb.AddUserToEventResponse, error) {
	return c.Client.AddUserToEvent(ctx, req)
}
