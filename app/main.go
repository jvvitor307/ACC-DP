package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println()
		fmt.Println("=== ACC Data Platform ===")
		fmt.Println("1. Registrar usuario")
		fmt.Println("2. Login")
		fmt.Println("3. Trocar usuario")
		fmt.Println("4. Listar usuarios")
		fmt.Println("5. Quem sou eu")
		fmt.Println("6. Iniciar producer")
		fmt.Println("0. Sair")
		fmt.Print("\nEscolha: ")

		if !scanner.Scan() {
			break
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			registerUser(scanner)
		case "2":
			loginUser(scanner)
		case "3":
			switchUser(scanner)
		case "4":
			listUsers()
		case "5":
			whoami()
		case "6":
			startProducer()
		case "0":
			fmt.Println("Saindo...")
			return
		default:
			fmt.Println("Opcao invalida.")
		}
	}
}

func startProducer() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("Iniciando producer (Ctrl+C para parar)...")
	if err := runProducer(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Producer encerrou com erro: %v\n", err)
	}
	fmt.Println("Producer parado.")
}
