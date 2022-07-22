package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"
)

type threadSafePrintliner struct {
	l sync.Mutex
	w io.Writer
}

func newThreadSafePrintliner(w io.Writer) *threadSafePrintliner {
	return &threadSafePrintliner{w: w}
}

func (p *threadSafePrintliner) println(s string) {
	p.l.Lock()
	fmt.Fprintln(p.w, s)
	p.l.Unlock()
}

func readQuery(r io.Reader) string {
	s, _ := ioutil.ReadAll(r) // N.B. not interested in this error; might as well return an empty string
	return strings.TrimSpace(strings.Replace(string(s), "\n", " ", -1))
}

func trimEmpty(s []string) []string {
	var r = make([]string, 0)
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
