package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var pokemon = flag.String(
	"pokemon",
	"",
	"path to pokemon csv",
)

func main() {
	flag.Parse()

	file, err := os.Open(*pokemon)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		row := strings.Split(string(scanner.Text()), ",")
		if len(row) != 3 {
			panic("Invalid Row")
		}

		experience, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			panic("Invalid weight")
		}

		weight := float64(1) / experience

		fmt.Printf("%s,%s,%s\n", row[0], row[1], strconv.FormatFloat(weight, 'f', -1, 64))
	}
}
