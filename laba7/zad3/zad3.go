package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	In, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Ошибка запуска сервера", err)
		return
	}
	fmt.Println("TCP сервер слушает порт 8080")
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nПолучен сигнал завершения, останавливаем сервер...")
		cancel()
		In.Close()
	}()
	for {
		conn, err := In.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				wg.Wait()
				fmt.Println("Все соединения закрыты, сервер остановлен")
				return
			default:
				fmt.Println("Ошибка принятия соединения:", err)
				continue
			}
		}
		wg.Add(1)
		go handleConnection(ctx, conn, &wg)
	}
}

func handleConnection(ctx context.Context, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		msg, err := reader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}
		msg = strings.TrimSpace(msg)
		fmt.Printf("Получено от %s: %s\n", conn.RemoteAddr(), msg)
		_, err = conn.Write([]byte("Сообщение получено: " + msg + "\n"))
		if err != nil {
			return
		}
	}
}
