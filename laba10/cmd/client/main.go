package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"laba10/models"
)

const tokenFile = ".lab9-token"

func main() {
	baseURL := os.Getenv("API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8443"
	}
	c := NewClient(baseURL)

	if data, err := os.ReadFile(tokenFile); err == nil {
		t := strings.TrimSpace(string(data))
		if t != "" {
			c.SetToken(t)
			fmt.Println("загружен сохранённый токен")
		}
	}

	in := bufio.NewReader(os.Stdin)
	for {
		printMenu(c)
		choice := readLine(in, "> ")
		fmt.Println()

		switch choice {
		case "1":
			doList(c, in)
		case "2":
			doGet(c, in)
		case "3":
			doCreate(c, in)
		case "4":
			doUpdate(c, in)
		case "5":
			doDelete(c, in)
		case "6":
			doLogin(c, in)
		case "7":
			doRegister(c, in)
		case "8":
			doLogout(c)
		case "0":
			fmt.Println("До встречи!")
			return
		default:
			fmt.Println("Неизвестная команда")
		}
	}
}

func printMenu(c *Client) {
	state := "не авторизован"
	if c.Token() != "" {
		state = "авторизован"
	}
	fmt.Printf("\n═══ Лабораторная 9 — клиент (%s) ═══\n", state)
	fmt.Println("  1. Список пользователей")
	fmt.Println("  2. Пользователь по ID")
	fmt.Println("  3. Создать пользователя")
	fmt.Println("  4. Обновить пользователя")
	fmt.Println("  5. Удалить пользователя")
	fmt.Println("  6. Войти")
	fmt.Println("  7. Зарегистрироваться")
	fmt.Println("  8. Выйти")
	fmt.Println("  0. Выход из программы")
}

func doList(c *Client, in *bufio.Reader) {
	opts := ListOptions{}
	if s := readLine(in, "фильтр по имени (Enter — пропустить): "); s != "" {
		opts.Name = s
	}
	if s := readLine(in, "мин. возраст (Enter — пропустить): "); s != "" {
		opts.MinAge = atoiSafe(s)
	}
	if s := readLine(in, "макс. возраст (Enter — пропустить): "); s != "" {
		opts.MaxAge = atoiSafe(s)
	}
	if s := readLine(in, "limit (Enter — 20): "); s != "" {
		opts.Limit = atoiSafe(s)
	}
	if s := readLine(in, "offset (Enter — 0): "); s != "" {
		opts.Offset = atoiSafe(s)
	}

	users, err := c.ListUsers(opts)
	if err != nil {
		reportError(err)
		return
	}
	printUsers(users)
}

func doGet(c *Client, in *bufio.Reader) {
	id := readInt(in, "ID: ")
	u, err := c.GetUser(id)
	if err != nil {
		reportError(err)
		return
	}
	printUser(u)
}

func doCreate(c *Client, in *bufio.Reader) {
	u := models.User{
		Name:  readLine(in, "Имя: "),
		Email: readLine(in, "Email: "),
		Age:   readInt(in, "Возраст: "),
	}
	created, err := c.CreateUser(u)
	if err != nil {
		reportError(err)
		return
	}
	fmt.Println("✅ Создан:")
	printUser(created)
}

func doUpdate(c *Client, in *bufio.Reader) {
	id := readInt(in, "ID: ")
	u := models.User{
		Name:  readLine(in, "Новое имя: "),
		Email: readLine(in, "Новый email: "),
		Age:   readInt(in, "Новый возраст: "),
	}
	updated, err := c.UpdateUser(id, u)
	if err != nil {
		reportError(err)
		return
	}
	fmt.Println("Обновлён:")
	printUser(updated)
}

func doDelete(c *Client, in *bufio.Reader) {
	id := readInt(in, "ID: ")
	if err := c.DeleteUser(id); err != nil {
		reportError(err)
		return
	}
	fmt.Println("Удалён")
}

func doLogin(c *Client, in *bufio.Reader) {
	email := readLine(in, "Email: ")
	password := readLine(in, "Пароль: ")

	resp, err := c.Login(email, password)
	if err != nil {
		reportError(err)
		return
	}
	c.SetToken(resp.Token)
	_ = os.WriteFile(tokenFile, []byte(resp.Token), 0o600)

	fmt.Println("Вошли как:")
	printUser(resp.User)
}

func doRegister(c *Client, in *bufio.Reader) {
	req := models.RegisterRequest{
		Name:     readLine(in, "Имя: "),
		Email:    readLine(in, "Email: "),
		Age:      readInt(in, "Возраст: "),
		Password: readLine(in, "Пароль (мин. 6): "),
	}
	resp, err := c.Register(req)
	if err != nil {
		reportError(err)
		return
	}
	c.SetToken(resp.Token)
	_ = os.WriteFile(tokenFile, []byte(resp.Token), 0o600)

	fmt.Println("Зарегистрирован и вошёл:")
	printUser(resp.User)
}

func doLogout(c *Client) {
	if err := c.Logout(); err != nil {
		reportError(err)
		return
	}
	c.SetToken("")
	_ = os.Remove(tokenFile)
	fmt.Println("Вышли из сессии")
}

func atoiSafe(s string) int {
	var n int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}
