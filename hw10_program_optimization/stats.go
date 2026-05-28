package hw10programoptimization

import (
	"bufio"
	"io"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

type DomainStat map[string]int

var iterPool = &sync.Pool{
	New: func() interface{} {
		return jsoniter.NewIterator(jsoniter.ConfigFastest)
	},
}

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	domainSuffix := "." + domain

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		email := parseEmail(line)
		if email == "" {
			continue
		}

		if !hasSuffixFold(email, domainSuffix) {
			continue
		}

		atPos := strings.LastIndexByte(email, '@')
		if atPos == -1 {
			continue
		}

		result[strings.ToLower(email[atPos+1:])]++
	}

	return result, scanner.Err()
}

func parseEmail(line []byte) (email string) {
	iter := iterPool.Get().(*jsoniter.Iterator)
	iter.ResetBytes(line)

	iter.ReadObjectCB(func(iter *jsoniter.Iterator, field string) bool {
		if field == "Email" {
			email = iter.ReadString()
		} else {
			iter.Skip()
		}
		return true
	})

	if iter.Error == nil {
		iterPool.Put(iter)
	}
	return
}

func hasSuffixFold(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return strings.EqualFold(s[len(s)-len(suffix):], suffix)
}
