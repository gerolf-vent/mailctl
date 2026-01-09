package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

func ReadPassword(fromStdin bool) (string, error) {
	var password string
	if fromStdin {
		reader := bufio.NewReader(os.Stdin)
		pw, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			utils.PrintErrorWithMessage("failed to read password from stdin", err)
			return "", err
		}
		password = pw
	} else {
		fmt.Print("Password: ")
		pw, err := term.ReadPassword(os.Stdin.Fd())
		fmt.Println()
		if err != nil {
			fmt.Println("read error:", err)
			return "", err
		}
		password = string(pw)
	}

	password = strings.TrimSpace(password)
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	return password, nil
}

func ReadPasswordHashed(method string, fromStdin bool) (string, error) {
	password, err := ReadPassword(fromStdin)
	if err != nil {
		return "", err
	}

	passwordHash, err := utils.PasswordHash(password, method)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return passwordHash, nil
}
