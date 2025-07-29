package handler

import (
	"changeme/backend/disk"
	"changeme/backend/mail"
	"fmt"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
	"os"
)

const (
	Email              = "EMAIL"
	Password           = "PASSWORD"
	AuthorizationToken = "AUTHORIZATION_TOKEN"
)

func GetEnv(key string, emit func(string)) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		emit(key + " не задан")
		return "", fmt.Errorf("%s не задан", key)
	}
	return v, nil
}

func CheckEnvData(emit func(string)) bool {
	err := godotenv.Load()
	if err != nil {
		emit(fmt.Sprintf("Ошибка при загрузке .env файла: %v", err))
	}

	_, emailErr := GetEnv(Email, emit)
	_, passwordErr := GetEnv(Password, emit)

	return emailErr == nil && passwordErr == nil
}

func SaveCredentials(email, password string) error {
	envMap := map[string]string{
		Email:    email,
		Password: password,
	}

	_ = godotenv.Load(".env")
	existing, _ := godotenv.Read(".env")

	for k, v := range envMap {
		existing[k] = v
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}

	return godotenv.Write(existing, ".env")
}

func Start(emit func(string)) error {
	if err := godotenv.Load(); err != nil {
		emit(fmt.Sprintf("Ошибка при загрузке .env файла: %v", err))
		return fmt.Errorf("ошибка при загрузке .env файла: %v", err)
	}

	email, err := GetEnv(Email, emit)
	if err != nil {
		emit("Почта не указана в .env")
		return fmt.Errorf("почта не указана в .env")
	}
	password, err := GetEnv(Password, emit)
	if err != nil {
		emit("Пароль не указан в .env")
		return fmt.Errorf("пароль не указан в .env")
	}
	token, err := GetEnv(AuthorizationToken, emit)
	if err != nil {
		emit("Токен авторизации не указан в .env")
		return fmt.Errorf("токен авторизации не указан в .env")
	}

	emit("Подключение к IMAP...")

	imapClient, err := mail.Connect(email, password)
	if err != nil {
		emit(fmt.Sprintf("Ошибка подключения к почте: %v", err))
		return fmt.Errorf("ошибка подключения к почте: %v", err)
	}
	emit("Подключение к IMAP успешно")
	defer func(imapClient *client.Client) {
		if err := imapClient.Logout(); err != nil {
			emit(fmt.Sprintf("Ошибка выхода из клиента: %v", err))
		}
	}(imapClient)

	diskSession := disk.NewSession(token)

	emit("Чтение писем...")

	if err := mail.ProcessEmails(imapClient, diskSession, emit); err != nil {
		emit(fmt.Sprintf("Ошибка обработки почты: %v", err))
		return fmt.Errorf("ошибка обработки почты: %v", err)
	}

	return nil
}
