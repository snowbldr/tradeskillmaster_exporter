package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		panic("You must provide a command")
	}
	if os.Args[1] == "load" {
		if len(os.Args) < 5 {
			panic("You must provide the json data file to load, the realm, and the db password")
		}
		data := ReadJsonDataFile(os.Args[2])
		realm := os.Args[3]
		dbPassword := os.Args[4]
		InsertData(data.DownloadTime, realm, data.ToParams(), dbPassword, data.Fields)
	} else {
		if len(os.Args) < 5 {
			panic("You must provide the tsm data file, realm, output directory, and db password as arguments")
		}
		tsmDataFile := os.Args[1]
		realm := os.Args[2]
		outputDir := os.Args[3]
		dbPassword := os.Args[4]
		time, fields, data := ExtractData(tsmDataFile, realm, outputDir)
		if data != nil {
			InsertData(time, realm, data, dbPassword, fields)
		}
	}
}

func ExtractData(tsmDataFile string, realm string, outputDir string) (uint64, []string, [][]interface{}) {

	file, err := os.Open(tsmDataFile)
	if err != nil {
		panic("Cannot open tsm data file " + tsmDataFile + "\nCause: " + err.Error())
	}
	info, err := file.Stat()
	if err != nil {
		panic("Cannot stat tsm data file " + tsmDataFile + "\nCause: " + err.Error())
	}

	if !fileExists(outputDir) {
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

	time, err := strconv.ParseUint(downloadTime, 10, 64)

	if err != nil {
		panic("Invalid time: " + downloadTime)
	}

	outputFile := filepath.Join(outputDir, realm+"_"+downloadTime+".json")
	if fileExists(outputFile) {
		println("Current data already exported to: " + outputFile)
		return 0, nil, nil
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

	dataRows := strings.Split(data[1:len(data)-1], "],[")
	rows := make([][]interface{}, len(dataRows))
	for r, row := range dataRows {
		columns := strings.Split(row, ",")
		rows[r] = make([]interface{}, len(columns))
		for i, v := range columns {
			rows[r][i] = v
		}
	}

	return time, strings.Split(strings.ReplaceAll(fields, "\"", ""), ","), rows
}

type JsonData struct {
	DownloadTime uint64
	Fields       []string
	Data         [][]interface{}
}

func (j *JsonData) ToParams() [][]interface{} {
	d := make([][]interface{}, len(j.Data))
	for i, row := range j.Data {
		d[i] = make([]interface{}, len(row))
		for c, col := range row {
			d[i][c] = fmt.Sprintf("%v", col)
		}
	}
	return d
}

func ReadJsonDataFile(file string) JsonData {
	jsonFile, err := os.Open(file)
	if err != nil {
		panic("Failed to open file " + file + ": " + err.Error())
	}
	defer jsonFile.Close()
	bytes, err := io.ReadAll(jsonFile)

	var jsonData JsonData

	err = json.Unmarshal(bytes, &jsonData)

	if err != nil {
		panic("Failed to parse file " + file + ": " + err.Error())
	}

	return jsonData
}

var pricingFieldMap = map[string]string{
	"itemString":  "item_id",
	"marketValue": "market_value",
	"minBuyout":   "min_buyout",
	"historical":  "historical",
	"numAuctions": "num_auctions",
}

func InsertData(time uint64, realm string, rows [][]interface{}, dbPassword string, fields []string) {
	conn, err := pgx.Connect(context.Background(), "postgres://tsm:"+url.QueryEscape(dbPassword)+"@localhost:4473/tsm")

	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	sql := BuildInsertSql(fields)
	for r, row := range rows {
		_, err := conn.Exec(context.Background(), sql, append(row, time, realm)...)
		if err != nil {
			println("Failed to insert row " + strconv.Itoa(r) + ": " + err.Error())
		}
	}
}

func BuildInsertSql(fields []string) string {
	sql := "insert into price_history ("
	//time and realm are included already, and added at the end so the values can be appended to the params
	fieldCount := 2
	for _, field := range fields {
		if col, ok := pricingFieldMap[field]; ok {
			sql += col + ","
			fieldCount++
		} else {
			println("Unknown field: " + field)
		}
	}
	sql = sql + "time, realm) values ("
	for i := 1; i <= fieldCount; i++ {
		//little janky, but the timestamp is a number and needs to be converted to a date, would probably be better to
		//convert it on the go side, but uhhhhh, meh...
		if i == fieldCount-1 {
			sql += "to_timestamp($" + strconv.Itoa(i) + "),"
		} else {
			sql += "$" + strconv.Itoa(i) + ","
		}

	}
	return sql[0:len(sql)-1] + ")"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
