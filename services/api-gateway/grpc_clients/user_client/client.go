package user_client

import (
	"context"
	"os"

	pb "github.com/wutthichod/sa-connext/shared/proto/User"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	userServiceAddr = "localhost:9093"
)

type UserServiceClient struct {
	Client pb.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserServiceClient() (*UserServiceClient, error) {
	UserServiceUrl := os.Getenv("User_SERVICE_URL")
	if UserServiceUrl == "" {
		UserServiceUrl = userServiceAddr
	}

	conn, err := grpc.NewClient(UserServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func (ds *UserServiceClient) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	return ds.Client.CreateUser(ctx, req)
}
