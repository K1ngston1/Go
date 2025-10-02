package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

func replaceRuns(s string) string {
	rs := []rune(s)
	var b strings.Builder
	for i := 0; i < len(rs); {
		j := i + 1
		for j < len(rs) && rs[j] == rs[i] {
			j++
		}
		if j-i >= 2 {
			b.WriteRune('+')
		} else {
			b.WriteRune(rs[i])
		}
		i = j
	}
	return b.String()
}

func shuffleWord(word string) string {
	rs := []rune(word)
	rand.Shuffle(len(rs), func(i, j int) {
		rs[i], rs[j] = rs[j], rs[i]
	})
	return string(rs)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	data, err := ioutil.ReadFile("/home/k1ngst0n/GolandProjects/Go/Laba3/input.txt")
	if err != nil {
		fmt.Println("Помилка читання input.txt:", err)
		return
	}

	lines := strings.Split(string(data), "\n")
	var outLines []string

	for _, line := range lines {
		parts := strings.FieldsFunc(line, func(r rune) bool {
			return r == ' ' || r == ','
		})

		var transformed []string
		for _, w := range parts {
			if w == "" {
				continue
			}
			t := replaceRuns(w)
			t = shuffleWord(t)
			transformed = append(transformed, t)
		}

		outLines = append(outLines, strings.Join(transformed, "-"))
	}
	err = ioutil.WriteFile("output.txt", []byte(strings.Join(outLines, "\n")), 0644)
	if err != nil {
		fmt.Println("Помилка запису output.txt:", err)
		return
	}

	fmt.Println("Готово! Результат у файлі output.txt")
}
