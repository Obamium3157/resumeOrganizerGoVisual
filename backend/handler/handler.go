package handler

import (
	"changeme/backend/disk"
	"changeme/backend/mail"
	"fmt"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
	"log"
	"os"
)

const (
	Email              = "EMAIL"
	Password           = "PASSWORD"
	AuthorizationToken = "AUTHORIZATION_TOKEN"
)

type EnvError struct {
	msg string
}

func (e *EnvError) Error() string {
	return e.msg
}

func getEnv(key string, emit func(string)) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		emit(key + " не задан")
		return "", &EnvError{msg: fmt.Sprintf("%s не задан", key)}
	}
	return v, nil
}

func CheckEnvData(emit func(string)) bool {
	err := godotenv.Load()
	if err != nil {
		emit(fmt.Sprintf("Ошибка при загрузке .env файла: %v", err))
	}

	_, emailErr := getEnv(Email, emit)
	_, passwordErr := getEnv(Password, emit)

	return emailErr == nil && passwordErr == nil
}

func Start(emit func(string)) {
	err := godotenv.Load()
	if err != nil {
		emit(fmt.Sprintf("Ошибка при загрузке .env файла: %v", err))
		log.Fatal("Ошибка при загрузке .env файла")
	}

	email, err := getEnv(Email, emit)
	if err != nil {
		emit("Почта не указана в .env")
	}
	password, err := getEnv(Password, emit)
	if err != nil {
		emit("Пароль не указан в .env")
	}
	token, err := getEnv(AuthorizationToken, emit)
	if err != nil {
		emit("Токен авторизации не указан в .env")
	}

	emit("Подключение к IMAP...")

	imapClient, err := mail.Connect(email, password)
	if err != nil {
		emit(fmt.Sprintf("Ошибка подключения к почте: %v", err))
		log.Fatalf("Ошибка подключения к почте: %v", err)
	}
	emit("Подключение к IMAP успешно")
	defer func(imapClient *client.Client) {
		err := imapClient.Logout()
		if err != nil {
			emit(fmt.Sprintf("Ошибка выхода из клиента: %v", err))
			log.Fatalf("Ошибка выхода из клиента: %v", err)
		}
	}(imapClient)

	diskSession := disk.NewSession(token)

	emit("Чтение писем...")

	if err := mail.ProcessEmails(imapClient, diskSession, emit); err != nil {
		emit(fmt.Sprintf("Ошибка обработки почты: %v", err))
		log.Fatalf("Ошибка обработки почты: %v", err)
	}
}
