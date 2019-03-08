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
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"

	"github.com/kumuluz/kumuluzee-go-config/config"
	"github.com/kumuluz/kumuluzee-go-discovery/discovery"
)

var disc discovery.Util

func main() {
	// initialize discovery
	configPath := path.Join(".", "config.yaml")

	disc = discovery.New(discovery.Options{
		Extension:  "consul",
		ConfigPath: configPath,
	})

	// register this service to registry
	// note: service parameters are read from configuration file
	_, err := disc.RegisterService(discovery.RegisterOptions{})
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/lookup", func(w http.ResponseWriter, r *http.Request) {
		// define parameters of the service we are looking for
		// and call DiscoverService
		serviceURL, err := disc.DiscoverService(discovery.DiscoverOptions{
			Value:       "test-service",
			Version:     "1.0.0",
			Environment: "dev",
			AccessType:  discovery.AccessTypeDirect,
		})
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
		} else {
			// prepare a struct for marshalling into json
			data := struct {
				Service string `json:"service"`
			}{
				serviceURL,
			}

			// generate json from data
			genjson, err := json.Marshal(data)
			if err != nil {
				w.WriteHeader(500)
			} else {
				// write generated json to ResponseWriter
				fmt.Fprint(w, string(genjson))
			}
		}
	})

	// initialize configuration
	conf := config.NewUtil(config.Options{
		Extension:  "consul",
		ConfigPath: configPath,
	})

	// perform service deregistration on received interrupt or terminate signals
	deregisterOnSignal()

	// get port number from configuration
	port, ok := conf.GetInt("kumuluzee.server.http.port")
	if !ok {
		log.Printf("Error reading port from configuration")
		port = 9000
	}

	// run server
	log.Printf("Starting server on port %d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

func deregisterOnSignal() {
	// catch interrupt or terminate signals and send them to sigs channel
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// function waits for received signal - and then performs service deregistration
	go func() {
		<-sigs
		if err := disc.DeregisterService(); err != nil {
			panic(err)
		}
		// besides deregistration, you could also do any other clean-up here.
		// Make sure to call os.Exit() with status number at the end.
		os.Exit(1)
	}()
}
