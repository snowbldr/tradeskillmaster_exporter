package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 4 {
		panic("You must provide the tsm data file, realm, and output directory as arguments")
	}
	tsmDataFile := os.Args[1]
	realm := os.Args[2]
	outputDir := os.Args[3]

	file, err := os.Open(tsmDataFile)
	if err != nil {
		panic("Cannot open tsm data file " + tsmDataFile + "\nCause: " + err.Error())
	}
	info, err := file.Stat()
	if err != nil {
		panic("Cannot stat tsm data file " + tsmDataFile + "\nCause: " + err.Error())
	}

	if !exists(outputDir) {
		panic("Output dir: " + outputDir + " does not exist")
	}
	scanner := bufio.NewScanner(file)
	size := int(info.Size())
	scanner.Buffer(make([]byte, 0, size), size)
	line := ""
	for scanner.Scan() {
		line = scanner.Text()
		if strings.Contains(strings.ToLower(line), strings.ToLower(realm)) {
			break
		}
	}
	if line == "" {
		panic("Realm " + realm + " not found in tsm data file " + tsmDataFile)
	}

	timeRegex := regexp.MustCompile(".*downloadTime=([0-9]+).*")
	fieldsRegex := regexp.MustCompile(".*fields=\\{(\".*\")}.*")
	dataRegex := regexp.MustCompile(".*data=\\{(\\{.*})}.*")

	downloadTime := timeRegex.ReplaceAllString(line, "$1")
	outputFile := filepath.Join(outputDir, realm+"_"+downloadTime+".json")
	if exists(outputFile) {
		println("Current data already exported to: " + outputFile)
	}

	out, err := os.Create(outputFile)
	if err != nil {
		panic("Failed to create output file\nCause: " + err.Error())
	}
	defer out.Close()
	fields := fieldsRegex.ReplaceAllString(line, "$1")
	data := dataRegex.ReplaceAllString(line, "$1")
	//remove extra } on end of data
	data = data[:len(data)-1]
	data = strings.ReplaceAll(strings.ReplaceAll(data, "}", "]"), "{", "[")
	_, err = out.WriteString("{\"downloadTime\":" + downloadTime + ",\"fields\":[" + fields + "],\"data\":[" + data + "]}")
	if err != nil {
		panic("Failed to write data\nCause: " + err.Error())
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
