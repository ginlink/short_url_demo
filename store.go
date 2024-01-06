package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	save chan record
}

type record struct {
	Key, URL string
}

const cap = 1000

func NewURLStore(filename string) *URLStore {
	s := &URLStore{
		urls: make(map[string]string),
		save: make(chan record, cap),
	}

	if err := s.load(filename); err != nil {
		log.Fatal("Error loading file", err)
	}

	go s.consumeRecord(filename)

	return s
}

func (s *URLStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.urls[key]
}

func (s *URLStore) Set(key string, url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查存在性
	// 不允许写入覆盖
	if _, exist := s.urls[key]; exist {
		return false
	}

	s.urls[key] = url
	return true
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.urls)
}

// 根据url生成短连接，并写入文件
func (s *URLStore) Put(url string) string {
	for {
		key := genKey(s.Count())

		if ok := s.Set(key, url); ok {
			// TODO 这里为什么是小写？
			// 定义是大写开头
			s.save <- record{key, url}

			return key
		}

	}
}

func (s *URLStore) load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error open file", err)
		return err
	}
	defer f.Close()

	// 循环从json文件中读取数据
	// 写入map中
	d := json.NewDecoder(f)
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(r.Key, r.URL)
		}
	}
	if err == io.EOF {
		return nil
	}

	log.Println("Error decoding record", err)
	return err
}

func (s *URLStore) consumeRecord(filename string) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Error opening file", err)
	}
	defer f.Close()

	e := json.NewEncoder(f)

	for {
		r := <-s.save
		if err := e.Encode(r); err != nil {
			fmt.Println("Error saving to file", err)
		}
	}
}
