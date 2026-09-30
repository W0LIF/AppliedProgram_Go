package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	In, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
		return
	}
	defer In.Close()
	fmt.Println("TCP сервер слушает порт 8080")
	for {
		conn, err := In.Accept()
		if err != nil {
			fmt.Println("Ошибка при приёме соединения:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		msg = strings.TrimSpace(msg)
		fmt.Printf("Получно от %s: %s\n", conn.RemoteAddr(), msg)
		_, err = conn.Write([]byte("Сообщение получено: " + msg + "\n"))
		if err != nil {
			return
		}
	}
}
