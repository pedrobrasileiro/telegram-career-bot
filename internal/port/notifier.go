package port

// Notifier envia uma mensagem assíncrona pra um chat, usado pelos usecases
// de job (scan/evaluate/ask) que rodam em background e avisam o resultado
// depois de o handler original já ter respondido.
type Notifier interface {
	Notify(chatID int64, message string) error
	NotifyHTML(chatID int64, message string) error
}
