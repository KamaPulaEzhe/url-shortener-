package storage

// import (
// 	"errors"
// 	"testing"
// 	errs "urlshortener/errs"
// )

// // func Test_SGDUrl(t *testing.T) {
// // 	s := NewStorage()
// // 	shortUrl := "abc"
// // 	expected := "https://proglib.io/p/samouchitel-po-go-dlya-nachinayushchih-chast-16-testirovanie-koda-i-ego-vidy-table-driven-podhod-parallelnye-testy-2024-09-03?ysclid=mr24u4xfjr191088381"
// // 	s.SetUrl(shortUrl, expected)
// // 	given, err := s.GetUrl(shortUrl)
// // 	if errors.Is(err, errs.ErrNotFound) {
// // 		t.Fatalf("does not exists")
// // 	}
// // 	if given != expected {
// // 		t.Fatalf("given mismathes  expected")
// // 	}

// // 	s.DeleteUrl(shortUrl)
// // 	given, err = s.GetUrl(shortUrl)
// // 	if errors.Is(err, errs.ErrNotFound) != true {
// // 		t.Fatalf("delete failed")
// // 	}

// // }

// func TestStorage(t *testing.T) {
// 	s := NewStorage()
// 	tests := []struct {
// 		name     string
// 		shortUrl string
// 		url      string
// 		wantErr  bool
// 	}{
// 		{"set and get", "abc", "https://example.com", false},
// 		{"not found", "xyz", "", true},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.url != "" {
// 				s.SetUrl(tt.shortUrl, tt.url)
// 			}
// 			got, err := s.GetUrl(tt.shortUrl)
// 			if tt.wantErr {
// 				if !errors.Is(err, errs.ErrNotFound) {
// 					t.Errorf("expected ErrNotFound, got %v", err)
// 				}
// 			} else {
// 				if err != nil {
// 					t.Fatalf("unexpected error: %v", err)
// 				}
// 				if got != tt.url {
// 					t.Errorf("got %s, want %s", got, tt.url)
// 				}
// 			}
// 		})
// 	}
// }

// // func fanIn(ch1, ch2 <-chan int) <-chan int {
// // 	res := make(chan int)
// // 	go func() {
// // 		for {
// // 			select {
// // 			case v := <-ch1:
// // 				res <- v
// // 			case v := <-ch2:
// // 				res <- v
// // 			}
// // 		}
// // 	}()

// // 	return res
// // }
