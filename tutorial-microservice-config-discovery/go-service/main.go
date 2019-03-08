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
	"net/http"

	"github.com/kumuluz/kumuluzee-go-config/config"
	"github.com/kumuluz/kumuluzee-go-discovery/discovery"

	"github.com/gin-gonic/gin"
)

var mockDB []Customer
var conf config.Util
var disc discovery.Util

func main() {
	// initialize functions
	initDB()
	initConfig()
	initDiscovery()

	// register service to service registry
	disc.RegisterService(discovery.RegisterOptions{})

	router := gin.Default()

	// Registers middleware function, which for each request checks our external configuration and
	// if 'maintenance' key is set to true, it will return error saying service is unavailable,
	// otherwise it will call next handler.
	// To test, while running go to http://localhost:8500 and change key
	// 'environments/dev/services/node-service/1.0.0/config/rest-config/maintenance' to 'true' and
	// then try to perform a request. To enable it again, just change the key to 'false'
	router.Use(func(c *gin.Context) {
		maintenanceMode, _ := conf.GetBool("rest-config.maintenance")
		if maintenanceMode {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, ErrorResponse{
				http.StatusServiceUnavailable,
				"Service is undergoing maintenance, check back in a minute.",
			})
		} else {
			c.Next()
		}
	})

	// prepare routes and map them to handlers
	v1c := router.Group("/v1/customers")
	{
		// GET /v1/customers/
		v1c.GET("/", getCustomers)
		// GET /v1/customers/:id
		v1c.GET("/:id", getCustomerByID)
		// GET /v1/customers/:id/orders/
		v1c.GET("/:id/orders", getOrdersByCustomerID)
		// POST /v1/customers/
		v1c.POST("/", createCustomer)
		// GET /v1/customers/:id/neworder
		v1c.GET("/:id/neworder", createOrder)
	}

	// run REST API server
	router.Run(":9000")
}

func initDB() {
	mockDB = make([]Customer, 0)
	mockDB = append(mockDB,
		Customer{100, "John", "Carlile", "john.ca@mail.com", "053347863"},
		Customer{101, "Ann", "Lockwood", "lockwood_ann@mail.com", "023773123"},
		Customer{102, "Elizabeth", "Mathews", "eli23@mail.com", "043343403"},
		Customer{103, "Isaac", "Anderson", "isaac.anderson@mail.com", "018743831"},
		Customer{104, "Barret", "Peyton", "barretp@mail.com", "063343148"},
		Customer{105, "Terry", "Cokes", "terry_cokes@mail.com", "053339123"},
	)
}

func initConfig() {
	conf = config.NewUtil(config.Options{
		Extension:  "consul",
		ConfigPath: "config.yaml",
	})
}

func initDiscovery() {
	disc = discovery.New(discovery.Options{
		Extension:  "consul",
		ConfigPath: "config.yaml",
	})
}
