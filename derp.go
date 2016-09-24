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

// The Details struct represents a record's details struct, with the same keys as the JSON response from OVH.
type Details struct {
	Exists    bool   `json:"exists,omitempty"`
	ActualIP  string `json:"actualip,omitempty"`
	Target    string `json:"target,omitempty"`
	Zone      string `json:"zone,omitempty"`
	FieldType string `json:"fieldType,omitempty"`
	SubDomain string `json:"subDomain,omitempty"`
	ID        int    `json:"id,omitempty"`
}

// Init method initializes the new instanced Details struct and sets the corresponding attributes.
func (d *Details) Init(client *ovh.Client, details Details) {
	d.GetRecordID(client, details)
	if d.Exists {
		d.GetDetails(client, details)
	}
	d.GetActualIP()
}

// hasChanged returns true if your actual IP and the IP the record is pointing at has changed. Returns false otherwise.
func hasChanged(actualIP string, targetIP string) bool {
	if strings.Compare(actualIP, targetIP) == 0 {
		return false
	}
	return true
}

// GetActualIP gets your actual IP address from ipinfo.io and sets ActualIP Details attribute.
func (d *Details) GetActualIP() {
	resp, err := http.Get("http://ipinfo.io/ip")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("Error reading file contents: ", err)
	}
	d.ActualIP = strings.TrimSpace(string(body))
	return
}

// GetRecordID gets the record identifier from OVH and sets the attribute Exists to true in case the record already exists. False otherwise.
func (d *Details) GetRecordID(client *ovh.Client, details Details) {
	var recordID []int
	err := client.Get("/domain/zone/"+details.Zone+"/record?fieldType="+details.FieldType+"&subDomain="+details.SubDomain, &recordID)
	if err != nil {
		log.Fatalln("Error querying details:", err)
		return
	}
	if len(recordID) > 0 {
		d.ID = recordID[0]
		d.Exists = true
	} else {
		d.Exists = false
	}
	return
}

// GetDetails gets all the record details from OVH. It returns the instanced Details struct with all the attributes from OVH already set.
func (d *Details) GetDetails(client *ovh.Client, details Details) {
	err := client.Get("/domain/zone/"+details.Zone+"/record/"+strconv.Itoa(d.ID), &d)
	if err != nil {
		log.Fatalln("Failed to get details: ", err)
	}
	return
}

// CreateRecord creates the record entry on your DNS zone.
func (d *Details) CreateRecord(client *ovh.Client, details Details) {
	fmt.Println(details)
	update := &Details{FieldType: d.FieldType, SubDomain: d.SubDomain, Target: d.ActualIP}
	fmt.Println(*update)
	err := client.Post("/domain/zone/"+d.Zone+"/record", update, nil)
	if err != nil {
		log.Fatalln("Error creating record: ", err)
	}
}

// UpdateRecord does the following:
// - checks if the record already exists
// - updates the new entry for your record if your IP changed
// - creates the new entry if it didn't exist.
func (d *Details) UpdateRecord(client *ovh.Client, details Details) {
	if d.Exists == true && d.ActualIP != d.Target {
		update := &Details{Target: details.ActualIP, SubDomain: details.SubDomain}
		err := client.Put("/domain/zone/"+details.Zone+"/record/"+strconv.Itoa(details.ID), update, nil)
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
