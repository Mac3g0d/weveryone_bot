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

var participants []string

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	lastCommandCall := make(map[int64]time.Time)
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

		switch update.Message.Command() {

		case "add":
			AddHandler(bot, update)

		case "addme":
			AddMeHandler(bot, update)

		case "del":
			DelHandler(bot, update)

		case "delme":
			DelMeHandler(bot, update)
		case "list":
			ListHandler(bot, update)

		case "all", "everyone":
			if lastCallTime, ok := lastCommandCall[update.Message.Chat.ID]; ok && time.Now().Sub(lastCallTime) < time.Minute {
				msg := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"Вы уже вызывали эту команду в последнюю минуту")
				bot.Send(msg)
			} else {
				lastCommandCall[update.Message.Chat.ID] = time.Now()
				AllHandler(bot, update)
			}

		case "help":
			HelpHandler(bot, update)
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

func DelParticipants(arr []string) {
	for _, username := range arr {
		username = strings.ReplaceAll(username, "@", "")
		if contains(participants, username) {
			for i, user := range participants {
				if user == username {
					participants = append(participants[:i], participants[i+1:]...)
					break
				}
			}
		}

	}
}

func AddParticipants(arr []string) {
	for _, username := range arr {
		username = strings.ReplaceAll(username, "@", "")
		if !contains(participants, username) {
			participants = append(participants, username)
		}

	}
}
func AddMeHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	username := update.Message.From.UserName
	text := ""

	if !contains(participants, username) {
		participants = append(participants, username)
		text = "Пользователь " + username + " добавлен в список."
	} else {
		text = "Пользователь " + username + " уже есть в списке."
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bot.Send(msg)
}

func DelMeHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	username := update.Message.From.UserName
	text := ""

	if contains(participants, username) {
		for i, user := range participants {
			if user == username {
				participants = append(participants[:i], participants[i+1:]...)
				break
			}
		}
		text = "Пользователь " + username + " удален из списка."
	} else {
		text = "Пользователя " + username + " нет в списке."
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bot.Send(msg)
}

func AddHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	arguments := update.Message.CommandArguments()
	username := ""
	text := ""
	usernames := strings.Split(arguments, " ")

	if len(usernames) < 2 {
		username = usernames[0]
	}

	if len(usernames) > 2 {
		AddParticipants(usernames)
		text = "Пользователи :\n" + strings.Join(usernames, "\n") + "\nдобавленны в список."
	} else {
		if !contains(participants, username) {
			participants = append(participants, username)
			text = "Добавлен пользователь " + username + " в список."
		} else {
			text = "Пользователь " + username + " уже есть в списке."
		}
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bot.Send(msg)
}

func DelHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	arguments := update.Message.CommandArguments()
	username := ""
	text := ""
	usernames := strings.Split(arguments, " ")

	if len(usernames) < 2 {
		username = usernames[0]
	}

	if len(usernames) > 2 {
		DelParticipants(usernames)
		text = "Пользователи :\n" + strings.Join(usernames, "\n") + "\nудалены из списка."
	} else {
		if contains(participants, username) {
			for i, user := range participants {
				if user == username {
					participants = append(participants[:i], participants[i+1:]...)
					break
				}
			}
			text = "Пользователь " + username + " удален из списка."
		} else {
			text = "Пользователя " + username + " нет в списке."
		}
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	bot.Send(msg)
}

func ListHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	var sb strings.Builder
	sb.WriteString("Список участников:\n")
	for i, user := range participants {
		sb.WriteString(strconv.Itoa(i+1) + ". " + user + "\n")
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	bot.Send(msg)
}

func AllHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	var sb strings.Builder
	sb.WriteString("Подсосы, общий сбор!\n" + "@" + strings.Join(participants, " @") + " ")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	bot.Send(msg)
}

func HelpHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	var sb strings.Builder
	text := `Доступные комманды:
/addme - добавить себя в список для _@all_
/add *username* - добавить участника в список по никнейму
/add *username1* *username2* ... *usernameN* добавить несколько участников в список по никнейму
/delme - удалить себя из списка
/del *username* - удалить участника из списока по никнейму
/del *username1* *username2* ... *usernameN* удалить несколько участников из списока по никнейму
/list - показать список всех участников
/all, /everyone - тегнуть в чате всех кто в списке`
	sb.WriteString(text)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, sb.String())
	msg.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(msg)
}
