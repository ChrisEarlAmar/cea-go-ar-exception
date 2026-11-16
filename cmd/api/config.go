package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	go_ora "github.com/sijms/go-ora/v2"
)

func oracleDSNFromEnv() (string, error) {
	sid, err := requiredEnv("ORACLE_SID")
	if err != nil {
		return "", err
	}
	host, err := requiredEnv("ORACLE_HOST")
	if err != nil {
		return "", err
	}
	username, err := requiredEnv("ORACLE_USERNAME")
	if err != nil {
		return "", err
	}
	password, err := requiredEnv("ORACLE_PASSWORD")
	if err != nil {
		return "", err
	}

	port := 1521
	if rawPort := strings.TrimSpace(os.Getenv("ORACLE_PORT")); rawPort != "" {
		port, err = strconv.Atoi(rawPort)
		if err != nil || port < 1 || port > 65535 {
			return "", fmt.Errorf("ORACLE_PORT must be a number from 1 to 65535")
		}
	}

	return go_ora.BuildUrl(host, port, "", username, password, map[string]string{
		"SID": sid,
	}), nil
}

func requiredEnv(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}
