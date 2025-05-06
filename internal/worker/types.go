package worker

type job struct {
	data  []byte
	index uint32
}

type result struct {
	index uint32
	data  []byte
	size  int
	err   error
}
