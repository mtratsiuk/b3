package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func LoadDotEnv(path string) error {
	dotenv, err := os.ReadFile(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("failed to read dotenv file: %v", err)
	}

	for idx, line := range strings.Split(strings.TrimSpace(string(dotenv)), "\n") {
		// Ignore comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		k, v, ok := strings.Cut(line, "=")

		if !ok {
			return fmt.Errorf("failed to read dotenv file at line %v, expected `KEY=value`", idx)
		}

		if _, defined := os.LookupEnv(k); !defined {
			if err := os.Setenv(k, v); err != nil {
				return fmt.Errorf("failed to set env value from dotenv file for key %v, %v", k, err)
			}
		}
	}

	return nil
}
