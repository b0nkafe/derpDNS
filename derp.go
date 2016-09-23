/*
==============================================================================

Copyright (c) 2016 Bruno Heras <srm@b0nk.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

==============================================================================
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ovh/go-ovh/ovh"
)

type Details struct {
	Exists    bool   `json:"exists,omitempty"`
	ActualIP  string `json:"actualip,omitempty"`
	Target    string `json:"target",omitempty`
	Zone      string `json:"zone,omitempty"`
	FieldType string `json:"fieldType,omitempty"`
	SubDomain string `json:"subDomain,omitempty"`
	Id        int    `json:"id,omitempty"`
}

func (d *Details) Init(client *ovh.Client, details Details) {
	d.GetRecordID(client, details)
	if d.Exists {
		d.GetDetails(client, details)
	}
	d.GetActualIP()
}

func hasChanged(actualIP string, targetIP string) bool {
	if strings.Compare(actualIP, targetIP) == 0 {
		return false
	}
	return true
}

func (d *Details) GetActualIP() {
	resp, err := http.Get("http://ipinfo.io/ip")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	d.ActualIP = strings.TrimSpace(string(body))
	return
}

func (d *Details) GetRecordID(client *ovh.Client, details Details) {
	var recordID []int
	err := client.Get("/domain/zone/"+details.Zone+"/record?fieldType="+details.FieldType+"&subDomain="+details.SubDomain, &recordID)
	if err != nil {
		log.Fatalln("Error querying details:", err)
		return
	}
	if len(recordID) > 0 {
		d.Id = recordID[0]
		d.Exists = true
	} else {
		d.Exists = false
	}
	return
}

func (d *Details) GetDetails(client *ovh.Client, details Details) {
	err := client.Get("/domain/zone/"+details.Zone+"/record/"+strconv.Itoa(d.Id), &d)
	if err != nil {
		log.Fatalln("Failed to get details: ", err)
	}
	return
}

func (d *Details) CreateRecord(client *ovh.Client, details Details) {
	fmt.Println(details)
	update := &Details{FieldType: d.FieldType, SubDomain: d.SubDomain, Target: d.ActualIP}
	fmt.Println(*update)
	err := client.Post("/domain/zone/"+d.Zone+"/record", update, nil)
	if err != nil {
		log.Fatalln("Error creating record: ", err)
	}
}

func (d *Details) UpdateRecord(client *ovh.Client, details Details) {
	if d.Exists == true && d.ActualIP != d.Target {
		update := &Details{Target: details.ActualIP, SubDomain: details.SubDomain}
		err := client.Put("/domain/zone/"+details.Zone+"/record/"+strconv.Itoa(details.Id), update, nil)
		if err != nil {
			log.Fatalln("Error updating target: ", err)
		}
		log.Printf("Changes detected for %s.%s. Renewing... %s => %s.\n", d.SubDomain, d.Zone, d.ActualIP, details.Target)
	} else if d.ActualIP == d.Target {
		log.Printf("No changes for %s.%s.\n", d.SubDomain, d.Zone)
	} else {
		d.CreateRecord(client, details)
		log.Printf("Created new record: %s.%s => type %s pointing to %s.\n", d.SubDomain, d.Zone, d.FieldType, d.ActualIP)
	}
	return
}
