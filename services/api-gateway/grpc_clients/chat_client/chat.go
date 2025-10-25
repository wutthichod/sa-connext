package chat_client

import (
	"context"

	"github.com/wutthichod/sa-connext/shared/config"
	pb "github.com/wutthichod/sa-connext/shared/proto/chat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChatServiceClient struct {
	Client pb.ChatServiceClient
	conn   *grpc.ClientConn
}

func NewChatServiceClient(config config.Config) (*ChatServiceClient, error) {

	conn, err := grpc.NewClient(config.App().Chat, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewChatServiceClient(conn)
	return &ChatServiceClient{
		Client: client,
		conn:   conn,
	}, nil
}

func (c *ChatServiceClient) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func (c *ChatServiceClient) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	return c.Client.CreateChat(ctx, req)
}

func (c *ChatServiceClient) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	return c.Client.SendMessage(ctx, req)
}
