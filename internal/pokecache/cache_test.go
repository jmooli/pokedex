package pokecache

import (
	"fmt"
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		key string
		val []byte
	}{
		{
			key: "https://example.com",
			val: []byte("testdata"),
		},
		{
			key: "https://example.com/path",
			val: []byte("moretestdata"),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cache := NewCache(interval)
			cache.Add(c.key, c.val)
			val, ok := cache.Get(c.key)
			if !ok {
				t.Errorf("expected to find key")
				return
			}
			if string(val) != string(c.val) {
				t.Errorf("expected value %s, got %s", string(c.val), string(val))
				return
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond
	cache := NewCache(baseTime)
	cache.Add("https://example.com", []byte("testdata"))

	if _, ok := cache.Get("https://example.com"); !ok {
		t.Errorf("expected to find key")
		return
	}

	time.Sleep(waitTime)

	if _, ok := cache.Get("https://example.com"); ok {
		t.Errorf("expected key to be cleaned up")
		return
	}
}

func TestCleanNonExpiredEntries(t *testing.T) {
	const expiration = 50 * time.Millisecond
	cache := NewCache(expiration)
	key := "https://example.com/nonexpired"
	expectedData := []byte("nonexpired data")
	cache.Add(key, expectedData)

	time.Sleep(20 * time.Millisecond)

	cache.Clean(expiration)

	if val, ok := cache.Get(key); !ok {
		t.Errorf("expected key %s to be present", key)
	} else if string(val) != string(expectedData) {
		t.Errorf("expected value %s, got %s", string(expectedData), string(val))
	}
}
func TestCleanExpiredEntries(t *testing.T) {
	const expiration = 20 * time.Millisecond
	cache := NewCache(expiration)
	key := "https://example.com/expired"
	cache.Add(key, []byte("expired data"))

	time.Sleep(30 * time.Millisecond)

	cache.Clean(expiration)

	if _, ok := cache.Get(key); ok {
		t.Errorf("expected key %s to be expired and removed", key)
	}
}
