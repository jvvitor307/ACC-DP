package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"golang.org/x/term"

	"acc-dp/user"
)

func backendURL() string {
	u := os.Getenv("BACKEND_URL")
	if u == "" {
		return "http://localhost:18088"
	}
	return u
}

func storagePath() string {
	p := os.Getenv("USER_STORAGE_PATH")
	if p == "" {
		return "./data/users.json"
	}
	return p
}

func prompt(scanner *bufio.Scanner, label string) string {
	fmt.Printf("%s: ", label)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func readPassword() (string, error) {
	fmt.Print("Senha: ")
	pwdBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(pwdBytes), nil
}

func registerUser(scanner *bufio.Scanner) {
	email := prompt(scanner, "Email")
	displayName := prompt(scanner, "Nome de exibicao")
	password, err := readPassword()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler senha: %v\n", err)
		return
	}

	if email == "" || displayName == "" || password == "" {
		fmt.Println("Email, nome e senha sao obrigatorios.")
		return
	}

	client := user.NewClient(backendURL())
	result, err := client.Register(context.Background(), user.RegisterInput{
		Email:       email,
		DisplayName: displayName,
		Password:    password,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao registrar: %v\n", err)
		return
	}

	store, err := user.NewStore(storagePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir store: %v\n", err)
		return
	}

	u := user.User{
		ID:          result.UserID,
		Email:       result.Email,
		DisplayName: result.DisplayName,
		CreatedAt:   time.Now(),
	}

	if err := store.UpsertUser(u); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao salvar usuario: %v\n", err)
		return
	}

	if err := store.SetActive(u.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao definir ativo: %v\n", err)
		return
	}

	fmt.Printf("Usuario registrado e definido como ativo.\n  ID:    %s\n  Email: %s\n  Nome:  %s\n", u.ID, u.Email, u.DisplayName)
}

func loginUser(scanner *bufio.Scanner) {
	email := prompt(scanner, "Email")
	if email == "" {
		fmt.Println("Email e obrigatorio.")
		return
	}

	password, err := readPassword()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler senha: %v\n", err)
		return
	}

	hostname, _ := os.Hostname()

	client := user.NewClient(backendURL())
	result, err := client.Login(context.Background(), user.LoginInput{
		Email:     email,
		Password:  password,
		MachineID: hostname,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer login: %v\n", err)
		return
	}

	store, err := user.NewStore(storagePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir store: %v\n", err)
		return
	}

	u := user.User{
		ID:        result.UserID,
		Email:     email,
		CreatedAt: time.Now(),
	}

	for _, cached := range store.List() {
		if cached.ID == result.UserID {
			u.Email = cached.Email
			u.DisplayName = cached.DisplayName
			u.CreatedAt = cached.CreatedAt
			break
		}
	}

	if err := store.UpsertUser(u); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao salvar usuario: %v\n", err)
		return
	}

	if err := store.SetActive(u.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao definir ativo: %v\n", err)
		return
	}

	fmt.Printf("Logado como %s (%s).\n", u.DisplayName, u.Email)
}

func switchUser(scanner *bufio.Scanner) {
	loginUser(scanner)
}

func listUsers() {
	store, err := user.NewStore(storagePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir store: %v\n", err)
		return
	}

	users := store.List()
	if len(users) == 0 {
		fmt.Println("Nenhum usuario registrado. Use a opcao 1 para registrar.")
		return
	}

	activeUser, _ := store.Active()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ATIVO\tID\tEMAIL\tNOME")
	for _, u := range users {
		marker := ""
		if u.ID == activeUser.ID {
			marker = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", marker, u.ID, u.Email, u.DisplayName)
	}
	w.Flush()
}

func whoami() {
	store, err := user.NewStore(storagePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao abrir store: %v\n", err)
		return
	}

	u, err := store.Active()
	if err != nil {
		fmt.Println("Nenhum usuario ativo. Faca login primeiro (opcao 2).")
		return
	}

	fmt.Printf("ID:     %s\nEmail:  %s\nNome:   %s\n", u.ID, u.Email, u.DisplayName)
}
