package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type NestStruct struct {
	NestData  string `json:"nestdata"`
	NestData2 string `json:"othernestdata"`
}

func (ns NestStruct) String() string {
	return fmt.Sprintf("nestData: %s, otherNestData: %s", ns.NestData, ns.NestData2)
}

type TestStruct struct {
	Data     string     `json:"data"`
	NestData NestStruct `json:"envelope"`
}

func (ts TestStruct) String() string {
	return fmt.Sprintf("data: %s, nestData: %s", ts.Data, ts.NestData.String())
}

func main() {
	http.HandleFunc("/endpoints", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		testResponse := TestStruct{
			Data: "Tommy",
			NestData: NestStruct{
				NestData:  "Spaven",
				NestData2: "Whatever mate",
			},
		}
		fmt.Printf("\nGot a request!")
		fmt.Printf("\nSending response %s!", testResponse)
		json.NewEncoder(w).Encode(testResponse)
	})

	fmt.Printf("\nListening on port 3031")
	log.Fatal(http.ListenAndServe(":3031", nil))
}
