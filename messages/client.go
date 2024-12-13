package messages

type ClientRegister struct {
	Username string
	Password string
}

type ClientLogin struct {
	Username string
	Password string
}
