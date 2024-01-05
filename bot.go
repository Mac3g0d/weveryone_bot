package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	participants    = make(map[int64][]string)
	username        string
	lastCommandCall = make(map[int64]time.Time)
)

func init() {
	log.Printf("start load envs")
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("get error when trying load envs %s", err)
	}

}

func main() {
	AdminTelegramId, _ := strconv.ParseInt(os.Getenv("ADMIN_TELEGRAM_ID"), 10, 64)
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(tgbotapi.UpdateConfig{
		Timeout: u.Timeout,
	})

	for update := range updates {
		if update.Message == nil {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ParseMode = tgbotapi.ModeHTML
		switch update.Message.Command() {

		case "add":
			if update.Message.From.ID == AdminTelegramId {
				msg.Text = GetAddCommandText(update.Message.CommandArguments(), update.Message.Chat.ID)
			} else {
				msg.Text = "У вас нет прав для добавления в спискок"
			}

		case "addme":
			msg.Text = GetAddMeCommandText(update.Message.From.UserName, update.Message.Chat.ID)

		case "del":
			if update.Message.From.ID == AdminTelegramId {
				msg.Text = GetDelCommandText(update.Message.CommandArguments(), update.Message.Chat.ID)
			} else {
				msg.Text = "У вас нет прав для удаления из списка"
			}

		case "delme":
			msg.Text = GetDelMeCommandText(update.Message.From.UserName, update.Message.Chat.ID)

		case "list":
			msg.Text = GetListCommandText(update.Message.Chat.ID)

		case "all", "everyone":
			msg.Text = GetAllCommandText(update.Message.Chat.ID)

		case "help":
			msg.Text = GetHelpCommandText()
			msg.ParseMode = tgbotapi.ModeMarkdown

		}
		if msg.Text != "" {
			sentMessage, err := bot.Send(msg)
			if err != nil {
				log.Fatal(sentMessage, err)
			}
		}

	}

}

func contains(arr []string, item string) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}
	return false
}

func DelParticipants(arr []string, ChatID int64) {
	for _, name := range arr {
		name = strings.ReplaceAll(name, "@", "")
		if contains(participants[ChatID], name) {
			for i, user := range participants[ChatID] {
				if user == name {
					participants[ChatID] = append(participants[ChatID][:i], participants[ChatID][i+1:]...)
					break
				}
			}
		}
	}
}

func AddParticipants(arr []string, ChatID int64) {
	for _, name := range arr {
		name = strings.ReplaceAll(name, "@", "")
		if !contains(participants[ChatID], name) {
			participants[ChatID] = append(participants[ChatID], name)
		}
	}
}

func GetAddMeCommandText(username string, ChatID int64) string {
	if !contains(participants[ChatID], username) {
		participants[ChatID] = append(participants[ChatID], username)
		return "Пользователь " + username + " добавлен в список."
	} else {
		return "Пользователь " + username + " уже есть в списке."
	}
}

func GetDelMeCommandText(username string, ChatID int64) string {
	if contains(participants[ChatID], username) {
		for i, user := range participants[ChatID] {
			if user == username {
				participants[ChatID] = append(participants[ChatID][:i], participants[ChatID][i+1:]...)
				break
			}
		}
		return "Пользователь " + username + " удален из списка."
	} else {
		return "Пользователя " + username + " нет в списке."
	}
}

func GetAddCommandText(arguments string, ChatID int64) string {
	usernames := strings.Split(arguments, " ")
	if len(arguments) == 0 {
		return "Вы не указали пользователей которых нужно добавить в список."
	}

	if len(usernames) < 2 {
		username = usernames[0]
	}

	if len(usernames) > 2 {
		AddParticipants(usernames, ChatID)
		return "Пользователи :\n" + strings.Join(usernames, "\n") + "\nдобавленны в список."
	} else {
		if !contains(participants[ChatID], username) {
			participants[ChatID] = append(participants[ChatID], username)
			return "Добавлен пользователь " + username + " в список."
		} else {
			return "Пользователь " + username + " уже есть в списке."
		}
	}
}

func GetDelCommandText(arguments string, ChatID int64) string {
	usernames := strings.Split(arguments, " ")

	if len(arguments) == 0 {
		return "Вы не указали пользователей которых нужно удалить из список."
	}

	if len(usernames) < 2 {
		username = usernames[0]
	}

	if len(usernames) > 2 {
		DelParticipants(usernames, ChatID)
		return "Пользователи :\n" + strings.Join(usernames, "\n") + "\nудалены из списка."
	} else {
		if contains(participants[ChatID], username) {
			for i, user := range participants[ChatID] {
				if user == username {
					participants[ChatID] = append(participants[ChatID][:i], participants[ChatID][i+1:]...)
					break
				}
			}
			return "Пользователь " + username + " удален из списка."
		} else {
			return "Пользователя " + username + " нет в списке."
		}
	}

}

func GetListCommandText(ChatID int64) string {
	var sb strings.Builder
	sb.WriteString("Список участников:\n")
	sb.WriteString(strings.Join(participants[ChatID], "\n"))
	return sb.String()
}

func GetAllCommandText(ChatID int64) string {
	var sb strings.Builder
	if lastCallTime, ok := lastCommandCall[ChatID]; ok && time.Now().Sub(lastCallTime) < time.Minute {
		sb.WriteString("Эта команда уже была вызвана минутой ранее")
	} else {
		lastCommandCall[ChatID] = time.Now()
		sb.WriteString("Подсосы, общий сбор!\n" + "@" + strings.Join(participants[ChatID], " @") + " ")
	}
	return sb.String()

}

func GetHelpCommandText() string {
	var sb strings.Builder
	sb.WriteString(`Доступные комманды:
/addme - добавить себя в список для _@all_
/add *username* - добавить участника в список по никнейму
/add *username1* *username2* ... *usernameN* добавить несколько участников в список по никнейму
/delme - удалить себя из списка
/del *username* - удалить участника из списока по никнейму
/del *username1* *username2* ... *usernameN* удалить несколько участников из списока по никнейму
/list - показать список всех участников
/all, /everyone - тегнуть в чате всех кто в списке`)

	return sb.String()
}
