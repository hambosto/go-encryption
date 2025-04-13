package worker

type Job struct {
	Data  []byte
	Index uint32
}

type Result struct {
	Index uint32
	Data  []byte
	Size  int
	Error error
}
