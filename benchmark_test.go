package stash

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkGet(b *testing.B) {
	c := New[string, string]()
	c.Set("key", "value", 0)
	for range b.N {
		c.Get("key")
	}
}

func BenchmarkSet(b *testing.B) {
	c := New[string, string]()
	for i := range b.N {
		c.Set(fmt.Sprintf("key%d", i), "value", 0)
	}
}

func BenchmarkDelete(b *testing.B) {
	c := New[string, string]()
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("key%d", i), "value", 0)
	}
	b.ResetTimer()
	for i := range b.N {
		c.Delete(fmt.Sprintf("key%d", i))
	}
}

func BenchmarkConcurrentGet(b *testing.B) {
	for _, goroutines := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("goroutines-%d", goroutines), func(b *testing.B) {
			c := New[string, string]()
			c.Set("key", "value", 0)
			var wg sync.WaitGroup
			each := b.N / goroutines
			if each == 0 {
				each = 1
			}
			b.ResetTimer()
			for g := 0; g < goroutines; g++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for range each {
						c.Get("key")
					}
				}()
			}
			wg.Wait()
		})
	}
}

func BenchmarkConcurrentSet(b *testing.B) {
	for _, goroutines := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("goroutines-%d", goroutines), func(b *testing.B) {
			c := New[string, string]()
			var wg sync.WaitGroup
			each := b.N / goroutines
			if each == 0 {
				each = 1
			}
			b.ResetTimer()
			for g := 0; g < goroutines; g++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for i := range each {
						c.Set(fmt.Sprintf("key-%d-%d", id, i), "value", 0)
					}
				}(g)
			}
			wg.Wait()
		})
	}
}

func BenchmarkShardedGet(b *testing.B) {
	for _, goroutines := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("goroutines-%d", goroutines), func(b *testing.B) {
			sc := NewSharded[string, string](16)
			sc.Set("key", "value", 0)
			var wg sync.WaitGroup
			each := b.N / goroutines
			if each == 0 {
				each = 1
			}
			b.ResetTimer()
			for g := 0; g < goroutines; g++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for range each {
						sc.Get("key")
					}
				}()
			}
			wg.Wait()
		})
	}
}

func BenchmarkShardedSet(b *testing.B) {
	for _, goroutines := range []int{1, 4, 16, 64} {
		b.Run(fmt.Sprintf("goroutines-%d", goroutines), func(b *testing.B) {
			sc := NewSharded[string, string](16)
			var wg sync.WaitGroup
			each := b.N / goroutines
			if each == 0 {
				each = 1
			}
			b.ResetTimer()
			for g := 0; g < goroutines; g++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for i := range each {
						sc.Set(fmt.Sprintf("key-%d-%d", id, i), "value", 0)
					}
				}(g)
			}
			wg.Wait()
		})
	}
}
