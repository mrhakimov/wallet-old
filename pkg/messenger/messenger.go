package messenger

type Messenger interface {
	Send(message string) bool
	Recieve() (message string, ok bool)
}

type Telegram struct {
}

func (t *Telegram) Send(message string) bool {
	return true
}

func (t *Telegram) Recieve() (message string, ok bool) {
	return "", true
}
