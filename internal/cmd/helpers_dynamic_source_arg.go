package cmd

import (
	"errors"
	"os"
)

var (
	ErrNoValueProvided = errors.New("no value provided")
)

func ResolveDynamicSourceArg(directValue, filePath, envVarName string) (string, error) {
	if directValue != "" {
		return directValue, nil
	}

	if filePath != "" {
		return ReadFileValue(filePath)
	}

	if envVarName != "" {
		envValue := os.Getenv(envVarName)
		if envValue != "" {
			return envValue, nil
		}
	}

	return "", ErrNoValueProvided
}
