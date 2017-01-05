package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const replaceMark = "$foo"

func main() {
	input := readLine()
	res := expand(string(input), replaceStr)
	//fmt.Printf("%[1]*[2]s", 20, "#")
	fmt.Println(strings.Repeat("#", 100))
	fmt.Println(res)
}

func readLine() string {
	bio := bufio.NewReader(os.Stdin)
	line, _ := bio.ReadString('\n')
	return line
}

func expand(s string, f func(string) string) string {
	return s
}

func replaceStr(s string) string {
	return "!dupa!"
}
