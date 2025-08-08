package store

import "log"

// only allow 32 MB of ram allocation per registry
const BUF_SIZE = 32 * 1024 * 1024

type FileInfo struct {
	Name string
	Size int
	Type string
}

type Registry struct {
	Key            string
	Secret         string
	FileInfo       *FileInfo
	Buffer         []byte
	FreeBufferChan chan struct{}
	ChunkReadyChan chan struct{}
}

func newRegistry(key string, secret string, fileInfo *FileInfo) *Registry {
	return &Registry{
		Key:            key,
		Secret:         secret,
		FileInfo:       fileInfo,
		Buffer:         make([]byte, 0, BUF_SIZE),
		FreeBufferChan: make(chan struct{}, 1),
		ChunkReadyChan: make(chan struct{}, 1),
	}
}

// clear current chunk from the buffer and returns it so that we can send it to the client
func (r *Registry) DequeueChunk() []byte {
	// wait for chunk to be ready for distribution
	<-r.ChunkReadyChan
	tmp := r.Buffer
	r.Buffer = r.Buffer[:0]
	r.FreeBufferChan <- struct{}{}
	return tmp
}

func (r *Registry) WriteChunks(chunk []byte) {
	if len(r.Buffer) >= BUF_SIZE {
		// wait for free buf channel
		log.Println("Buffer chunk full, waiting for release")
		<-r.FreeBufferChan
	}
	r.Buffer = append(r.Buffer, chunk...)
	// signal that chunk is ready
	r.ChunkReadyChan <- struct{}{}
}

func (r *Registry) BuildChunks() []int {
	// divide the file size with the allowed buf to get the equally same size chunks
	chunkSize := r.FileInfo.Size / BUF_SIZE
	// and get the remainder since there is always some case where the distribution is not equal
	remainder := r.FileInfo.Size % BUF_SIZE

	chunks := []int{}

	for i := 0; i < chunkSize; i++ {
		chunks = append(chunks, BUF_SIZE)
	}

	if remainder > 0 {
		chunks = append(chunks, remainder)
	}

	return chunks
}
