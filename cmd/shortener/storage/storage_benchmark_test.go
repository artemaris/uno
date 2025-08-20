package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkInMemoryStorage_SaveGet(b *testing.B) {
	s := NewInMemoryStorage()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			id := fmt.Sprintf("id-%d", idx)
			url := fmt.Sprintf("https://example.com/%d", idx)
			s.Save(id, url, "user")
			s.Get(id)
			idx++
		}
	})
}

func BenchmarkInMemoryStorage_SaveBatch(b *testing.B) {
	s := NewInMemoryStorage()
	pairs := make(map[string]string, 100)
	for i := 0; i < 100; i++ {
		pairs[fmt.Sprintf("id-%d", i)] = fmt.Sprintf("https://example.com/%d", i)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.SaveBatch(pairs, "user")
	}
}

func BenchmarkFileStorage_SaveGet(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "db.jsonl")
	s, err := NewFileStorage(path)
	if err != nil {
		b.Fatalf("failed to create file storage: %v", err)
	}
	defer func() {
		if fs, ok := s.(*FileStorage); ok {
			fs.file.Close()
			os.Remove(path)
		}
	}()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("id-%d", i)
		url := fmt.Sprintf("https://example.com/%d", i)
		s.Save(id, url, "user")
		s.Get(id)
	}
}
