/*
 *  Copyright (c) 2019 Kumuluz and/or its affiliates
 *  and other contributors as indicated by the @author tags and
 *  the contributor list.
 *
 *  Licensed under the MIT License (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  https://opensource.org/licenses/MIT
 *
 *  The software is provided "AS IS", WITHOUT WARRANTY OF ANY KIND, express or
 *  implied, including but not limited to the warranties of merchantability,
 *  fitness for a particular purpose and noninfringement. in no event shall the
 *  authors or copyright holders be liable for any claim, damages or other
 *  liability, whether in an action of contract, tort or otherwise, arising from,
 *  out of or in connection with the software or the use or other dealings in the
 *  software. See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/kumuluz/kumuluzee-go-config/config"
)

type myConfig struct {
	StringProperty  string `config:"string-property,watch"`
	IntegerProperty int    `config:"integer-property"`
	BooleanProperty bool   `config:"boolean-property"`
	ObjectProperty  struct {
		SubProperty  string `config:"sub-property,watch"`
		SubProperty2 string `config:"sub-property-2"`
	} `config:"object-property"`
}

var conf myConfig

func main() {

	prefixKey := "rest-config"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// prepare a struct for marshalling into json
		data := struct {
			Value    string `json:"value"`
			Subvalue string `json:"subvalue"`
		}{
			conf.StringProperty,
			conf.ObjectProperty.SubProperty,
		}

		// generate json from data
		genjson, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(500)
		} else {
			// write generated json to ResponseWriter
			fmt.Fprint(w, string(genjson))
		}

	})

	configPath := path.Join(".", "config.yaml")

	// initialize configuration bundle
	opts := config.Options{
		Extension:  "consul",
		ConfigPath: configPath,
	}

	config.NewBundle(prefixKey, &conf, opts)

	// get port number from configuration aswell
	util := config.NewUtil(opts)
	port, ok := util.GetInt("kumuluzee.server.http.port")
	if !ok {
		log.Printf("Error reading port from configuration")
		port = 9000
	}

	// run server
	log.Printf("Starting server on port %d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))

}
