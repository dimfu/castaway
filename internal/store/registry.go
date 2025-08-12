package store

import (
	"log"
	"time"
)

type FileInfo struct {
	Name string
	Size int
	Type string
}

type Registry struct {
	Store    *Store
	Key      string
	FileInfo *FileInfo

	// to temporarily store file buffer/blob chunk from the client
	Buffer []byte

	// channel as a signals so that the transmission is synchronized correctly
	freeBufferChan chan struct{}
	chunkReadyChan chan struct{}

	// to indicate that the download is ready and make my life easier on the front end
	Ready bool

	Used          bool
	TimeUsedSince time.Time
}

func newRegistry(store *Store, key string, fileInfo *FileInfo) *Registry {
	r := &Registry{
		Store:          store,
		Key:            key,
		FileInfo:       fileInfo,
		Buffer:         make([]byte, 0, store.UsingChunkSize),
		freeBufferChan: make(chan struct{}, 1),
		chunkReadyChan: make(chan struct{}, 1),
		Ready:          false,
		Used:           false,
		TimeUsedSince:  time.Now(),
	}

	// check if whether the registry is inactive or not
	go r.inactiveChecker()

	// initially the buffer is free for writing
	r.freeBufferChan <- struct{}{}
	return r
}

// clear current chunk from the buffer and returns it so that we can send it to the client
func (r *Registry) DequeueChunk() []byte {
	<-r.chunkReadyChan
	r.Used = true
	r.TimeUsedSince = time.Now()
	tmp := make([]byte, len(r.Buffer))
	copy(tmp, r.Buffer)
	r.Buffer = r.Buffer[:0]
	// signal to writer that the r.buffer is ready
	r.freeBufferChan <- struct{}{}
	return tmp
}

// write chunks to registry buffer, and wait for the client to download the buffered chunk
// so that the chunk buffer can be cleared for the next chunk
func (r *Registry) WriteChunks(chunk []byte) {
	<-r.freeBufferChan // wait for "buffer free"
	if len(chunk) > cap(r.Buffer) {
		// clamp to buffer size to avoid overflowing the buffer
		chunk = chunk[:cap(r.Buffer)]
	}
	r.Buffer = r.Buffer[:len(chunk)]
	copy(r.Buffer, chunk)
	r.chunkReadyChan <- struct{}{} // signal "data ready"
}

func (r *Registry) BuildChunks() []int {
	// divide the file size with the allowed buf to get the equally same size chunks
	chunkSize := r.FileInfo.Size / cap(r.Buffer)
	// and get the remainder since there is always some case where the distribution is not equal
	remainder := r.FileInfo.Size % cap(r.Buffer)

	chunks := []int{}

	for i := 0; i < chunkSize; i++ {
		chunks = append(chunks, cap(r.Buffer))
	}

	if remainder > 0 {
		chunks = append(chunks, remainder)
	}

	if len(chunks) > 0 {
		r.Ready = true
	}

	return chunks
}

func (r *Registry) inactiveChecker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if r.Used && time.Now().After(r.TimeUsedSince.Add(30*time.Second)) {
				log.Printf("Registry %s has been removed due to inactivity", r.Key)
				r.Store.ClearRegistry(r.Key)
				return
			}
		}
	}
}
