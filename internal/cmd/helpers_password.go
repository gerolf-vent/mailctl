package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/gerolf-vent/mailctl/internal/utils/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
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

func ReadPasswordHashed(method string, options string, fromStdin bool) (string, error) {
	password, err := ReadPassword(fromStdin)
	if err != nil {
		return "", err
	}

	passwordHash, err := PasswordHash(password, method, options)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return passwordHash, nil
}

func PasswordHash(password, method string, options string) (string, error) {
	switch strings.ToLower(method) {
	case "bcrypt":
		cost := 11 // Default cost
		if options != "" {
			var err error
			cost, err = strconv.Atoi(options)
			if err != nil {
				return "", fmt.Errorf("invalid bcrypt cost: %w", err)
			}
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
		if err != nil {
			return "", fmt.Errorf("failed to hash password: %w", err)
		}
		return string(hashedPassword), nil
	case "argon2id":
		// Default parameters
		time := uint32(1)
		memory := uint32(64 * 1024)
		threads := uint8(4)

		var optTime, optMemory uint32
		var optThreads uint8
		var err error

		if options != "" {
			optTime, optMemory, optThreads, err = argon2.ParseHashParameters(options)
			if err != nil {
				return "", fmt.Errorf("invalid argon2id parameters: %w", err)
			}
			if optTime != 0 {
				time = optTime
			}
			if optMemory != 0 {
				memory = optMemory
			}
			if optThreads != 0 {
				threads = optThreads
			}
		}

		hashedPassword, err := argon2.GenerateFromPassword([]byte(password), time, memory, threads, 32)
		if err != nil {
			return "", fmt.Errorf("failed to hash password: %w", err)
		}
		return string(hashedPassword), nil
	}

	return "", fmt.Errorf("unsupported password hashing method: %s", method)
}
