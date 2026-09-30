package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Ошибка подключения:", err)
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Введите сообщение: ")
	msg, _ := reader.ReadString('\n')
	_, err = conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Ошибка отправки:", err)
		return
	}
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка чтения ответа:", err)
		return
	}
	fmt.Println("Ответ сервера:", response)
}
