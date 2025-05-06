package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stayfatal/VK-pub-sub/config"
	"github.com/stayfatal/VK-pub-sub/gen/pubsubpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stayfatal/VK-pub-sub/internal/service"
	"github.com/stayfatal/VK-pub-sub/pkg/logger"
	"github.com/stayfatal/VK-pub-sub/pkg/pubsub"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var testEnv struct {
	cfg    *config.Config
	ps     pubsub.PubSub
	client pubsubpb.PubSubClient
}

func TestMain(m *testing.M) {
	var err error
	testEnv.cfg, err = config.Load("../../.env")
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.MockInit()

	testEnv.ps = pubsub.NewPubSub()

	svc := service.NewService(testEnv.ps)

	srv := NewPubSubServer(svc, logger)

	grpcSrv := grpc.NewServer()

	pubsubpb.RegisterPubSubServer(grpcSrv, srv)

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", testEnv.cfg.Server.Port))
	if err != nil {
		logger.Error(err)
		return
	}

	go func() {
		logger.Info("server is listening", zap.String("port", testEnv.cfg.Server.Port))
		if err := grpcSrv.Serve(l); err != nil {
			logger.Error(err)
			return
		}
	}()

	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%s", testEnv.cfg.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error(err)
		return
	}
	testEnv.client = pubsubpb.NewPubSubClient(conn)

	code := m.Run()
	conn.Close()
	grpcSrv.GracefulStop()
	l.Close()
	testEnv.ps.Close(context.Background())

	os.Exit(code)
}

func TestPublish(t *testing.T) {
	var (
		testSubject = uuid.NewString()
	)
	_, err := testEnv.client.Publish(
		context.Background(),
		&pubsubpb.PublishRequest{
			Key:  testSubject,
			Data: "somedata",
		},
	)
	require.NoError(t, err)
}

func TestSubscribe(t *testing.T) {
	var (
		testSubject = uuid.NewString()
	)
	stream, err := testEnv.client.Subscribe(
		context.Background(),
		&pubsubpb.SubscribeRequest{
			Key: testSubject,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, stream)
	err = stream.CloseSend()
	require.NoError(t, err)
}

func TestPublishAndSubscribe(t *testing.T) {
	var (
		testSubject = uuid.NewString()
		testMsg     = "test message"
	)

	stream, err := testEnv.client.Subscribe(
		context.Background(),
		&pubsubpb.SubscribeRequest{Key: testSubject},
	)
	require.NoError(t, err)
	require.NotNil(t, stream)

	time.Sleep(time.Millisecond * 500)
	_, err = testEnv.client.Publish(context.Background(), &pubsubpb.PublishRequest{
		Key:  testSubject,
		Data: testMsg,
	})
	require.NoError(t, err)

	event, err := stream.Recv()
	require.NoError(t, err)
	require.NotNil(t, event)
	assert.Equal(t, testMsg, event.Data)
}

func TestPublishAfterClose(t *testing.T) {
	var (
		testSubject = uuid.NewString()
	)
	testEnv.ps.Close(context.Background())
	_, err := testEnv.client.Publish(context.Background(), &pubsubpb.PublishRequest{Key: testSubject, Data: "data"})
	require.Error(t, err)
	s, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, s.Code())
}

func TestSubscribeAfterClose(t *testing.T) {
	var (
		testSubject = uuid.NewString()
	)
	testEnv.ps.Close(context.Background())
	stream, err := testEnv.client.Subscribe(context.Background(), &pubsubpb.SubscribeRequest{Key: testSubject})
	require.NoError(t, err)

	_, err = stream.Recv()
	s, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, s.Code())
}

func TestSubscribeCloseStream(t *testing.T) {
	var (
		testSubject = uuid.NewString()
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := testEnv.client.Subscribe(ctx, &pubsubpb.SubscribeRequest{Key: testSubject})
	require.NoError(t, err)

	cancel()
	_, err = stream.Recv()
	s, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Canceled, s.Code())
}
