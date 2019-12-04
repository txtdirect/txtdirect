package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"go.txtdirect.org/txtdirect"
)

func main() {
	var types string
	var enabled []string

	flag.StringVar(&types, "types", "host,path,dockerv2,gometa,proxy,git", "Enable type. Separated using commas like \"host,path,git\"")
	flag.Parse()

	enabled = strings.Split(types, ",")

	config := txtdirect.Config{
		Enable: enabled,
	}

	if len(os.Args) < 2 {
		log.Fatalf("[txtdirect-validator]: A TXT record should be provided as a argument")
	}

	_, err := txtdirect.ParseRecord(os.Args[1], nil, &http.Request{}, config)
	if err != nil {
		log.Fatalf("[txtdirect-validator]: Couldn't parse the record: %s", err.Error())
	}

	log.Println("[txtdirect-validator]: The TXT record is valid.")
}
