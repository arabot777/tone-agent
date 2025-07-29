package rpccode

type ErrInfo struct {
	Title     string `json:"title"`
	Msg       string `json:"msg"`
	Reference string `json:"reference"`
}

// Register(code, message, title, reference)
func Register(code int, fields ...string) {
	errors.Register(code, fields...)
}

// Render(code)
func Render(code int) ErrInfo {
	return errors.Render(code)
}

func RenderCoder(err error) (int, ErrInfo) {
	return errors.ParseErrInfo(err)
}
