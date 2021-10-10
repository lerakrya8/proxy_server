package request

type Request struct {
	ID      int64
	Method  string
	Host    string
	URL     string
	Body    string
	Headers map[string][]string
}

