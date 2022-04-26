package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func LogInit(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

type EpicsBot struct {
	timeout time.Duration
	botToken string
	listenChannelID string
	Bot *tgbotapi.BotAPI
	TelegramBotToken string
	TelegramGroupId int64
}

func (e *EpicsBot) Run() {
	s, _ := discordgo.New("Bot " + e.botToken)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is ready")
	})
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.ChannelID == e.listenChannelID {
			e.SendToChat(m.Content)
		}
	})
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)

	err := s.Open()
	if err != nil {
		Error.Println("Cannot open the session:", err)
		os.Exit(1)
	}
	defer s.Close()


	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	Info.Println("Graceful shutdown")
}

func (e *EpicsBot) SendToChat(message string){
	msg := tgbotapi.NewMessage(e.TelegramGroupId, message)
	e.Bot.Send(msg)
	Info.Println("Message sended:", message)
}

func GetFromEnv(t string) string{
	s := os.Getenv(t)
	if s == "" {
		Error.Println("Env " + t + " is not defined.")
		os.Exit(1)
	}
	return s
}

func (e *EpicsBot) GetConfig()  {
	e.timeout = time.Second * 10
	e.botToken = GetFromEnv("TOKEN")
	e.listenChannelID = GetFromEnv("CHANNELID")
	e.TelegramBotToken = GetFromEnv("BOT_TOKEN")
	gId, err := strconv.Atoi(GetFromEnv("GROUP_ID"))
	if err != nil{
		Error.Println("Unable to parse GROUP_ID for Telegram Channel.")
		os.Exit(1)
	}
	e.TelegramGroupId = int64(gId)
	e.Bot = getBot(e.TelegramBotToken)
}

func getBot(token string) *tgbotapi.BotAPI{
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil{
		Error.Println("Bot cannot loaded.")
		return nil
	} else {
		return bot
	}
}

func init()  {
	LogInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
}

func main() {
	bot := EpicsBot{}
	bot.GetConfig()
	bot.Run()
}
