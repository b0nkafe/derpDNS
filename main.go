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
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/ovh/go-ovh/ovh"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s config_file", os.Args[0])
	}

	config, err := loadConfig(os.Args[1])
	if err != nil {
		log.Fatalln("Error parsing config file: ", err)
	}

	c, err := ovh.NewClient(config.OVH.Endpoint, config.OVH.ApplicationKey, config.OVH.ApplicationSecret, config.OVH.ConsumerKey)
	if err != nil {
		log.Fatalln("Error initializing OVH Client: ", err)
	}

	var record = Details{
		Zone:      config.Record.Zone,
		SubDomain: config.Record.SubDomain,
		FieldType: config.Record.RecordType,
	}

	record.Init(c, record)
	record.UpdateRecord(c, record)

}

// The Config struct type is the top level element for a Config type. It has two struct attributes that map the specific options for OVH and the DNS record itself
type Config struct {
	OVH    OVHConfig    `json:"ovh"`
	Record RecordConfig `json:"record"`
}

// The OVHConfig type maps the configuration for your OVH application
type OVHConfig struct {
	Endpoint          string `json:"endpoint"`
	ApplicationSecret string `json:"application_secret"`
	ApplicationKey    string `json:"application_key"`
	ConsumerKey       string `json:"consumer_key"`
}

// The RecordConfig type just maps the DNS record configuration
type RecordConfig struct {
	SubDomain  string `json:"subDomain"`
	Zone       string `json:"zone"`
	RecordType string `json:"recordType"`
}

// loadConfig parses and loads the config. It returns a Config object type and nil as error in case everything went OK. It should return an empty instance of a Config object with an error in case something fails
func loadConfig(file string) (Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	configContent, err := ioutil.ReadAll(f)
	if err != nil {
		return Config{}, err
	}

	config := new(Config)
	err = json.Unmarshal(configContent, &config)
	if err != nil {
		return Config{}, err
	}

	if config.Record.SubDomain == "" || config.Record.Zone == "" || config.Record.RecordType == "" || config.OVH.ApplicationKey == "" || config.OVH.ApplicationSecret == "" || config.OVH.ConsumerKey == "" || config.OVH.Endpoint == "" {
		var err = errors.New("derp, check config file because somehting's missing")
		return Config{}, err
	}

	return *config, nil
}
