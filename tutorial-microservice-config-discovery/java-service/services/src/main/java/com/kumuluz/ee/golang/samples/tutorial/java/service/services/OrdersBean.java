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

package com.kumuluz.ee.golang.samples.tutorial.java.service.services;

import com.kumuluz.ee.golang.samples.tutorial.java.service.persistence.exceptions.JavaServiceException;
import com.kumuluz.ee.golang.samples.tutorial.java.service.persistence.models.Order;
import com.kumuluz.ee.rest.beans.QueryParameters;
import com.kumuluz.ee.rest.utils.JPAUtils;

import javax.enterprise.context.ApplicationScoped;
import javax.persistence.EntityManager;
import javax.persistence.PersistenceContext;
import javax.transaction.Transactional;
import java.util.List;

@ApplicationScoped
public class OrdersBean {
	
	@PersistenceContext(unitName = "db-jpa-unit")
	private EntityManager entityManager;

	public List<Order> getOrders(QueryParameters query) {
		List<Order> orders = JPAUtils.queryEntities(entityManager, Order.class, query);
		return orders;
	}
	
	public Order getOrderById(long orderId) {
		Order order = entityManager.find(Order.class, orderId);
		if (order == null) {
			throw new JavaServiceException("Order not found!", 404);
		}
		return order;
	}
	
	@Transactional
	public void createOrder(Order order) {
		entityManager.persist(order);
	}
}
