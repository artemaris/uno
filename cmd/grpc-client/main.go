package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"uno/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Парсим флаги командной строки
	serverAddr := flag.String("addr", "localhost:9090", "gRPC server address")
	flag.Parse()

	// Проверяем переменную окружения
	if envAddr := os.Getenv("GRPC_SERVER_ADDR"); envAddr != "" {
		*serverAddr = envAddr
	}

	// Подключаемся к gRPC серверу
	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", *serverAddr, err)
	}
	defer conn.Close()

	// Создаем клиент
	client := proto.NewShortenerServiceClient(conn)

	// Тестируем ShortenURL
	fmt.Println("Testing ShortenURL...")
	shortenResp, err := client.ShortenURL(context.Background(), &proto.ShortenURLRequest{
		Url: "https://example.com",
	})
	if err != nil {
		log.Fatalf("ShortenURL failed: %v", err)
	}
	fmt.Printf("ShortenURL response: %s\n", shortenResp.Result)

	// Тестируем Ping
	fmt.Println("\nTesting Ping...")
	pingResp, err := client.Ping(context.Background(), &proto.PingRequest{})
	if err != nil {
		log.Fatalf("Ping failed: %v", err)
	}
	fmt.Printf("Ping response: %+v\n", pingResp)

	// Тестируем GetStats
	fmt.Println("\nTesting GetStats...")
	statsResp, err := client.GetStats(context.Background(), &proto.GetStatsRequest{})
	if err != nil {
		log.Fatalf("GetStats failed: %v", err)
	}
	fmt.Printf("Stats response: URLs=%d, Users=%d\n", statsResp.Urls, statsResp.Users)

	// Тестируем BatchShortenURL
	fmt.Println("\nTesting BatchShortenURL...")
	batchResp, err := client.BatchShortenURL(context.Background(), &proto.BatchShortenURLRequest{
		Items: []*proto.BatchURLItem{
			{
				CorrelationId: "1",
				OriginalUrl:   "https://google.com",
			},
			{
				CorrelationId: "2",
				OriginalUrl:   "https://github.com",
			},
		},
	})
	if err != nil {
		log.Fatalf("BatchShortenURL failed: %v", err)
	}
	fmt.Printf("BatchShortenURL response: %d items\n", len(batchResp.Items))
	for _, item := range batchResp.Items {
		fmt.Printf("  %s: %s\n", item.CorrelationId, item.ShortUrl)
	}

	fmt.Println("\nAll tests completed successfully!")
}
