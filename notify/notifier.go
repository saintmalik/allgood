package notify

type Notifier interface {
    Notify(message string) error
}

// type SlackNotifier struct {
//     WebhookURL string
// }

// func (s *SlackNotifier) Notify(message string) error {
//     return nil
// }

// type TelegramNotifier struct {
//     BotToken string
//     ChatID   string
// }

// func (t *TelegramNotifier) Notify(message string) error {
//     return nil
// }