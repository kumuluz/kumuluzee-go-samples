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
	"fmt"
	"net/http"
	"strconv"

	"github.com/dghubble/sling"
	"github.com/gin-gonic/gin"
	"github.com/kumuluz/kumuluzee-go-discovery/discovery"
)

// returns array of customers with 200 OK code
func getCustomers(c *gin.Context) {
	c.JSON(http.StatusOK, mockDB)
	return
}

// returns user object with 200 OK code if found
// and 404 NOT FOUND code if such user doesn't exists
func getCustomerByID(c *gin.Context) {
	sid := c.Param("id")
	id, err := strconv.ParseInt(sid, 0, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			http.StatusBadRequest,
			fmt.Sprintf("ID conversion to integer failed with error: %s", err.Error()),
		})
		return
	}

	for _, e := range mockDB {
		if e.ID == id {
			c.JSON(http.StatusOK, e)
			return
		}
	}

	c.JSON(http.StatusNotFound, ErrorResponse{
		http.StatusNotFound,
		fmt.Sprintf("Customer with id %d not found.", id),
	})
	return
}

// this endpoint calls our Java service to reterieve all orders for a given customer ID
func getOrdersByCustomerID(c *gin.Context) {

	// discover Java service...
	ordAddress, err := disc.DiscoverService(discovery.DiscoverOptions{
		Value:       "java-service",
		Environment: "dev",
		Version:     "1.0.0",
		AccessType:  "direct",
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	cID := c.Param("id")
	orders := new([]OrderResponse)

	_, err = sling.New().Get(ordAddress).Path("/v1/orders?where=customerId:EQ:" + cID).ReceiveSuccess(orders)
	if err != nil {
		// Java service returned something other than code 2xx
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, orders)
	return
}

// creates a new customer from POST body
func createCustomer(c *gin.Context) {
	var customer Customer

	err := c.ShouldBindJSON(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			http.StatusBadRequest,
			fmt.Sprintf("Could not create customer from JSON."),
		})
	}

	// give it an ID
	customer.ID = mockDB[len(mockDB)-1].ID + 1

	// add it to "database"
	mockDB = append(mockDB, customer)

	c.JSON(http.StatusCreated, customer)
	return
}

// this endpoint is an example of POSTing json data to our Java service
// it generates new Order Request and calls our Java service to create it
// Returns order with 201 CREATED code if successful.
func createOrder(c *gin.Context) {
	// prepare a new order
	sid := c.Param("id")
	id, err := strconv.ParseInt(sid, 0, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			http.StatusBadRequest,
			fmt.Sprintf("ID conversion to integer failed with error: %s", err.Error()),
		})
		return
	}

	ord := OrderRequest{
		CustomerID:  id,
		Title:       "New order",
		Description: "This is a new order.",
	}

	// discover Java service to post order to
	ordAddress, err := disc.DiscoverService(discovery.DiscoverOptions{
		Value:       "java-service",
		Environment: "dev",
		Version:     "1.0.0",
		AccessType:  "direct",
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// pointer to OrderResponse, where request's response will be stored
	ordResp := &OrderResponse{}

	// perform POST request
	fmt.Println(ordAddress)
	_, err = sling.New().Post(ordAddress).Path("/v1/orders").BodyJSON(ord).ReceiveSuccess(ordResp)
	if err != nil {
		// Java service returned something other than code 2xx
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, ordResp)
	return
}
