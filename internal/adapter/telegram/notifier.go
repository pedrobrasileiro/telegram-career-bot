package telegram

import "gopkg.in/telebot.v3"

// Notifier implementa port.Notifier enviando pro chat via bot.Send —
// usado pelos jobs assíncronos (scan/eval/ask) pra avisar o resultado
// depois de o handler original já ter respondido.
type Notifier struct {
	Bot *telebot.Bot
}

func (n Notifier) Notify(chatID int64, message string) error {
	_, err := n.Bot.Send(&telebot.User{ID: chatID}, message)
	return err
}

func (n Notifier) NotifyHTML(chatID int64, message string) error {
	_, err := n.Bot.Send(&telebot.User{ID: chatID}, message, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
	return err
}
