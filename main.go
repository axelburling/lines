package main

import (
	"bytes"
	"encoding/json"
	"github.com/fatih/color"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Error(err error) {
	if err != nil {
		panic(err)
	}
}

type Config struct {
	Name    string   `json:"name"`
	Exclude []string `json:"exclude"`
}

type Entry struct {
	name string
	dir  bool
}

var state = map[string]Entry{}

func removePrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

func RemoveFirstChar(input string) string {
	if len(input) <= 1 {
		return ""
	}
	return input[1:]
}

func getEntries(base string, dir string, c []string) map[string]Entry {

	_, e := color.New(color.FgGreen).Println("Scanning directory:", base)
	Error(e)

	err := filepath.Walk(path.Join(base, dir), func(path string, info os.FileInfo, err error) error {
		Error(err)

		if info.Name() == "lines.json" {
			return nil
		}

		for _, v := range c {
			if strings.Contains(removePrefix(path, base), v) && len(v) > 0 {
				return nil
			}
		}

		println(RemoveFirstChar(removePrefix(removePrefix(path, base), base+"/")))
		if info.IsDir() {
			state[path] = Entry{info.Name(), true}
		} else {
			state[path] = Entry{info.Name(), false}
		}
		return nil
	})

	Error(err)

	return state
}

func countLines(file string) int {

	_, e := color.New(color.FgBlue).Println("Counting lines in file:", file)
	Error(e)

	f, err := os.Open(file)
	Error(err)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			Error(err)
		}
	}(f)

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := f.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count

		case err != nil:
			Error(err)
		}
	}
}

func makeSpicyOutput(lines int, p string) {
	_, err := color.New(color.FgYellow).Add(color.Bold).Add(color.Underline).Println("Project", p, "has", lines, "lines of code")
	if err != nil {
		Error(err)
	}

	_, err = color.New(color.FgRed).Add(color.Bold).Println("If the previous number is abnormally high, you should probably consider adding a lines.json file to your project root to exclude certain files/directories from the count.")
	if err != nil {
		Error(err)
	}

	_, err = color.New(color.FgRed).Add(color.Bold).Println("If the previous number is abnormally low, you should probably consider start coding more. LOL")

	return
}

// reads lines.json in the current directory
func readConfig() ([]string, string) {

	c := Config{}

	b, err := os.ReadFile("lines.json")

	if err != nil {
		return []string{}, "root"
	}

	err = json.Unmarshal(b, &c)

	if err != nil {
		return []string{}, "root"
	}

	return c.Exclude, c.Name
}

func main() {
	dir, err := os.Getwd()

	Error(err)

	c, p := readConfig()

	s := getEntries(dir, "", c)

	totalLines := 0

	for k, v := range s {
		if !v.dir {
			lines := countLines(k)
			totalLines += lines
		}
	}

	makeSpicyOutput(totalLines, p)
}
