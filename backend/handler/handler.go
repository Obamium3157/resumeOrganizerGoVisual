package handler

import (
	"changeme/backend/disk"
	"changeme/backend/mail"
	"fmt"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"time"
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

func ValidateCredentials(email, password string) error {
	log.Println("Подключение к почте...")
	if _, err := mail.Connect(email, password); err != nil {
		log.Println("Подключение неуспешно: " + err.Error())
		return err
	}

	return nil
}

func ValidateToken(token string) error {
	const destination = "https://cloud-api.yandex.net/v1/disk"

	log.Printf("Отправлен запрос на: %v", destination)

	req, err := http.NewRequest("GET", destination, nil)
	if err != nil {
		return fmt.Errorf("не удалось создать запрос: %w", err)
	}
	req.Header.Set("Authorization", "OAuth "+token)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	log.Println("Отправка запроса с проверкой токена...")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка при запросе к API Яндекс.Диска: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Ошибка при вызове body.Close(): %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("некорректный токен (status %d): %s)", resp.StatusCode, string(body))
	}

	log.Println("Успешно!")

	return nil
}

func SaveToken(token string) error {
	envMap := map[string]string{
		AuthorizationToken: token,
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

	emit("Подключение к IMAP")

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

	emit("Чтение писем")

	if err := mail.ProcessEmails(imapClient, diskSession, emit); err != nil {
		emit(fmt.Sprintf("Ошибка обработки почты: %v", err))
		return fmt.Errorf("ошибка обработки почты: %v", err)
	}

	return nil
}
