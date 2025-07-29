package mail

import (
	"changeme/backend/disk"
	"fmt"
	"github.com/emersion/go-imap"
	"io"
	"log"
	"net/mail"
	"regexp"
	"strings"

	imapClient "github.com/emersion/go-imap/client"
)

const (
	imapAddress      string = "imap.yandex.ru:993"
	amountOfChannels int    = 10
	inbox            string = "INBOX"
	allMessages      string = "1:*"
	selectReadOnly   bool   = false
	firstSender      int    = 0
	tildaHostName    string = "tilda.ws"
)

func Connect(email, password string) (*imapClient.Client, error) {
	c, err := imapClient.DialTLS(imapAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("IMAP dial: %v", err)
	}

	if err := c.Login(email, password); err != nil {
		return nil, fmt.Errorf("login failed: %v", err)
	}
	log.Println("IMAP подключение успешно")
	return c, nil
}

func ProcessEmails(c *imapClient.Client, diskSession *disk.Session, emit func(string)) error {
	if err := selectMailBox(c, inbox); err != nil {
		return err
	}

	messages, section, err := fetchMessages(c)
	if err != nil {
		return err
	}

	for msg := range messages {
		if err := processMessage(msg, section, diskSession, emit); err != nil {
			emit(fmt.Sprintf("Ошибка при обработке письма: %v", err))
			log.Printf("Ошибка при обработке письма: %v", err)
		}
	}

	emit("Все письма обработаны!")
	log.Println("Все письма обработаны!")

	return nil
}

func selectMailBox(c *imapClient.Client, name string) error {
	_, err := c.Select(name, selectReadOnly)
	if err != nil {
		return fmt.Errorf("ошибка при выборе %s: %v", name, err)
	}

	return nil
}

func fetchMessages(c *imapClient.Client) (<-chan *imap.Message, *imap.BodySectionName, error) {
	seqSet := new(imap.SeqSet)
	if err := seqSet.Add(allMessages); err != nil {
		return nil, nil, fmt.Errorf("ошибка seqSet.Add: %v", err)
	}

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchBodyStructure, section.FetchItem()}

	messages := make(chan *imap.Message, amountOfChannels)
	go func() {
		if err := c.Fetch(seqSet, items, messages); err != nil {
			log.Fatalf("Ошибка client.Fetch: %v", err)
		}
	}()

	return messages, section, nil
}

func processMessage(msg *imap.Message, section *imap.BodySectionName, diskSession *disk.Session, emit func(string)) error {
	if msg.Envelope == nil {
		return nil
	}

	hostName := msg.Envelope.From[firstSender].HostName
	if !strings.Contains(hostName, tildaHostName) {
		return nil
	}

	body := msg.GetBody(section)
	if body == nil {
		return nil
	}

	fields, err := parseEmailBody(body)
	if err != nil {
		return err
	}

	link := fields["file_0"]
	if link == "" {
		return nil
	}

	filename, success := extractFilename(link)
	if !success {
		return nil
	}

	jobTitle := fields["job_title"]
	if jobTitle == "" {
		return nil
	}

	year := msg.Envelope.Date.Year()
	return handleResume(filename, jobTitle, year, diskSession, emit)
}

func parseEmailBody(body io.Reader) (map[string]string, error) {
	m, err := mail.ReadMessage(body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения MIME: %w", err)
	}

	bodyBytes, err := io.ReadAll(m.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела: %w", err)
	}

	return parseFields(string(bodyBytes)), nil
}

func parseFields(body string) map[string]string {
	re := regexp.MustCompile(`(?i)(file_0|job_title):\s*(.+?)(?:<br>|$)`)
	matches := re.FindAllStringSubmatch(body, -1)
	fields := make(map[string]string)
	for _, match := range matches {
		fields[strings.ToLower(match[1])] = strings.TrimSpace(match[2])
	}
	return fields
}

func extractFilename(input string) (string, bool) {
	re := regexp.MustCompile(`tilda\w+\.(zip|rar|7z)`)
	match := re.FindString(input)
	if match != "" {
		return match, true
	}

	return "", false
}

func handleResume(filename string, jobTitle string, year int, session *disk.Session, emit func(string)) error {
	searchFolder := "" // Ищем в руте
	f, err := disk.FindFile(searchFolder, filename, session)
	if err != nil || f == "" {
		emit(fmt.Sprintf("Файл не найден для: %s", filename))
		return fmt.Errorf("файл не найден для: %s", filename)
	}
	finalPath := fmt.Sprintf("%s/%d", jobTitle, year)

	if err := disk.CreateSeriesOfFolders(finalPath, session.Token); err != nil {
		emit(fmt.Sprintf("Не удалось создать структуру папок: %v", err))
		return fmt.Errorf("не удалось создать структуру папок: %v", err)
	}

	newPath := fmt.Sprintf("%s/%s", finalPath, f)
	if err := disk.MoveFile(f, newPath, session.Token); err != nil {
		emit(fmt.Sprintf("Ошибка при перемещении файла: %v", err))
		return fmt.Errorf("ошибка при перемещении файла: %v", err)
	}

	emit(fmt.Sprintf("Файл успешно перемещён: %s -> %s\n", f, newPath))
	log.Printf("Файл успешно перемещён: %s -> %s\n", f, newPath)
	return nil
}
