package chat_client

import (
	"context"
	"os"

	pb "github.com/wutthichod/sa-connext/shared/proto/chat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	chatServiceAddr = "localhost:9093" // default chat service address
)

type ChatServiceClient struct {
	Client pb.ChatServiceClient
	conn   *grpc.ClientConn
}

func NewChatServiceClient() (*ChatServiceClient, error) {
	chatServiceURL := os.Getenv("CHAT_SERVICE_URL")
	if chatServiceURL == "" {
		chatServiceURL = chatServiceAddr
	}

	conn, err := grpc.NewClient(chatServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
