# KumuluzEE Microservices in Go and Java 

Goal of this tutorial is to develop two microservices, one in Java and one in Go, which will communicate between themselves through KumuluzEE's service discovery and config.

We will develop a sample application for managing customers and their orders. Application consists of two microservices; one written in Go for managing customers and the other written in Java for managing orders.

Our Java microservice will use following KumuluzEE extensions:
- KumuluzEE Config for dynamic reconfiguration of microservices with the use of configuration servers,
- KumuluzEE Discovery for service registration and service discovery

Our Go microservice will use following KumuluzEE packages:
- `github.com/kumuluz/kumuluzee-go-config/config` for dynamic reconfiguration of microservices with the use of configuration servers,
- `github.com/kumuluz/kumuluzee-go-discovery/discovery` for service registration and service discovery

Both microservices will use Consul to store configuration and register services. With minor tweaks the tutorial will work with Etcd server as well.

First we will create Maven project that will contain our Java microservice. Since this part is covered in more detail in other samples, we will show just the important bits. Then we will create a simple REST API server with Go.

Complete source code can be found on the github repository.

## Run Consul server

You can download Consul from [their webpage](https://www.consul.io/).

Then you can run it by typing:
```
$ consul agent -dev
```
And you can access its UI by typing http://localhost:8500 into your browser.

You can also run Consul using Docker, by typing:
```
$ docker run -d --name=dev-consul --net=host consul
```

## Run local PostgreSQL instance

You can run PostgreSQL using Docker, by typing:
```
$ docker run --name orders-postgres --publish 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=orders -d postgres
```

## Create Maven project

We will create multi-module Maven project called java-service with following structure and pom.xml:
- java-service
    - api
    - services
    - persistence

```xml
<!-- root pom.xml -->
<properties>
    <java.version>1.8</java.version>
    <maven.compiler.source>1.8</maven.compiler.source>
    <maven.compiler.target>1.8</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>

    <kumuluzee.version>3.2.0</kumuluzee.version>
    <postgres.version>42.2.5</postgres.version>
    <kumuluzee-cors.version>1.0.5</kumuluzee-cors.version>
    <kumuluzee-config-consul.version>1.1.0</kumuluzee-config-consul.version>
    <kumuluzee-discovery-consul.version>1.1.0</kumuluzee-discovery-consul.version>
    <kumuluzee-rest.version>1.2.3</kumuluzee-rest.version>
</properties>

<dependencyManagement>
    <dependencies>
        <!-- KumuluzEE bom -->
        <dependency>
            <groupId>com.kumuluz.ee</groupId>
            <artifactId>kumuluzee-bom</artifactId>
            <version>${kumuluzee.version}</version>
            <type>pom</type>
            <scope>import</scope>
        </dependency>
        <dependency>
            <groupId>com.kumuluz.ee.cors</groupId>
            <artifactId>kumuluzee-cors</artifactId>
            <version>${kumuluzee-cors.version}</version>
        </dependency>
        <dependency>
            <groupId>com.kumuluz.ee.config</groupId>
            <artifactId>kumuluzee-config-consul</artifactId>
            <version>${kumuluzee-config-consul.version}</version>
        </dependency>
        <dependency>
            <groupId>com.kumuluz.ee.discovery</groupId>
            <artifactId>kumuluzee-discovery-consul</artifactId>
            <version>${kumuluzee-discovery-consul.version}</version>
        </dependency>
        <dependency>
            <groupId>com.kumuluz.ee.rest</groupId>
            <artifactId>kumuluzee-rest-core</artifactId>
            <version>${kumuluzee-rest.version}</version>
        </dependency>
        <!-- external -->
        <dependency>
            <groupId>org.postgresql</groupId>
            <artifactId>postgresql</artifactId>
            <version>${postgres.version}</version>
        </dependency>
    </dependencies>
</dependencyManagement>
```

### Persistence module

This module will store our data object and our persisted entities

```xml
<!-- persistence pom.xml -->
<dependencies>
    <dependency>
        <groupId>com.kumuluz.ee</groupId>
        <artifactId>kumuluzee-jpa-eclipselink</artifactId>
    </dependency>
    <dependency>
        <groupId>com.kumuluz.ee.rest</groupId>
        <artifactId>kumuluzee-rest-core</artifactId>
    </dependency>
    <dependency>
        <groupId>org.postgresql</groupId>
        <artifactId>postgresql</artifactId>
    </dependency>
</dependencies>
```

#### Order - persisted entity

This is Order entity, which is persisted into our local PostgreSQL database.

```java
@Entity
@Table(name = "orders")
@NamedQueries({
    @NamedQuery(name = "Order.findAllByCustomer", query = "SELECT o FROM Order o WHERE o.customerId = :customer_id")
})
public class Order {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private long id;
    
    @Column(name = "title")
    private String title;
    
    @Column(name = "description")
    private String description;
    
    @Column(name = "customer_id")
    private long customerId;

    // getters and setters ...
}
```

#### OrderRequest

This data object is used as request body to create new order.

```java 
public class OrderRequest {
    
    private long customerId;
    private String title;
    private String description;

    // getters and setters ...
}
```

#### CustomerResponse

This data object will be used later in the tutorial and it represents data that our Go service will return to our Java service.

```java
public class CustomerResponse {

    private long id;
    private String name;
    private String lastName;
    private String email;
    private String phone;
    
    // getters and setters ...
}
```

### Service module

This module provides beans that work with entity manager.

```xml
<!-- services pom.xml -->
<dependencies>
    <dependency>
        <groupId>com.kumuluz.ee</groupId>
        <artifactId>kumuluzee-jta-narayana</artifactId>
    </dependency>
    <dependency>
        <groupId>com.kumuluz.ee</groupId>
        <artifactId>kumuluzee-cdi-weld</artifactId>
    </dependency>
    <dependency>
        <groupId>com.kumuluz.ee.golang.samples.tutorial.java.service</groupId>
        <artifactId>persistence</artifactId>
        <version>${parent.version}</version>
    </dependency>
</dependencies>
```

#### OrdersBean

This bean is persisting our order entities into our database.

```java
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
```

### Api module

Api module is used to expose our application through REST Api and also to register and discover services.

```xml
<dependencies>
    <dependency>
        <groupId>com.kumuluz.ee</groupId>
        <artifactId>kumuluzee-core</artifactId>
    </dependency>
    <dependency>
        <groupId>com.kumuluz.ee</groupId>
        <artifactId>kumuluzee-servlet-jetty</artifactId>
    </dependency>
    <dependency>
        <groupId>com.kumuluz.ee.golang.samples.tutorial.java.service</groupId>
        <artifactId>services</artifactId>
        <version>${parent.version}</version>
    </dependency>
    <dependency>
        <groupId>com.kumuluz.ee</groupId>
        <artifactId>kumuluzee-jax-rs-jersey</artifactId>
    </dependency>
    <dependency>
        <groupId>com.kumuluz.ee.cors</groupId>
        <artifactId>kumuluzee-cors</artifactId>
    </dependency>
    <dependency>
        <groupId>com.kumuluz.ee.config</groupId>
        <artifactId>kumuluzee-config-consul</artifactId>
    </dependency>
    <dependency>
        <groupId>com.kumuluz.ee.discovery</groupId>
        <artifactId>kumuluzee-discovery-consul</artifactId>
    </dependency>
</dependencies>

<build>
    <plugins>
        <plugin>
            <groupId>com.kumuluz.ee</groupId>
            <artifactId>kumuluzee-maven-plugin</artifactId>
            <version>${kumuluzee.version}</version>
            <executions>
                <execution>
                    <id>package</id>
                    <goals>
                        <goal>repackage</goal>
                    </goals>
                </execution>
            </executions>
        </plugin>
    </plugins>
</build>
```

Api module also contains config.yaml which looks like this:

```yaml
kumuluzee:
  name: java-service
  version: 1.0.0
  env:
    name: dev
  server:
    base-url: http://localhost:8080
  config:
    start-retry-delay-ms: 500
    max-retry-delay-ms: 900000
    consul:
      hosts: http://localhost:8500
  discovery:
    consul:
      hosts: http://localhost:8500
    ttl: 20
    ping-interval: 15
  datasources:
    - jndi-name: jdbc/orders_database
      connection-url: jdbc:postgresql://localhost:5432/orders
      username: postgres
      password: postgres
      pool:
        max-size: 20
```

#### OrderApplication

Here in OrderApplication.java we register our service to our config server

```java
@ApplicationPath("v1")
@CrossOrigin
@RegisterService
public class OrderApplication extends Application {}
```

#### OrderResource

OrderResource.java exposes four endpoints:
- `GET  /orders/`: returns queried orders using [KumuluzEE REST](https://github.com/kumuluz/kumuluzee-rest),
- `GET  /orders/{orderId}`: returns order for given order ID,
- `GET  /orders/{orderId}/customer`: returns customer (specifically, CustomerResponse object) retrieved by calling our Go service, for given order ID,
- `POST /orders/ `: creates a new order from given JSON request body.

```java
@ApplicationScoped
@Path("orders")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class OrderResource {
    
    @Inject
    private OrdersBean ordersBean;

    @Context
    protected UriInfo uriInfo;

    @Inject
    @DiscoverService(value = "go-service", version = "1.0.0", environment = "dev")
    private Optional<WebTarget> serviceUrl;

    // get orders by query
    @GET
    public Response getOrders() {
        QueryParameters query = QueryParameters.query(uriInfo.getRequestUri().getQuery()).build();
        List<Order> orders = ordersBean.getOrders(query);
        return Response.status(Response.Status.OK).entity(orders).build();
    }

    // get order for given id
    @GET
    @Path("{orderId}")
    public Response getOrder(@PathParam("orderId") long orderId) {
        Order order = ordersBean.getOrderById(orderId);
        return Response.status(Response.Status.OK).entity(order).build();
    }

    // get customer for given order id
    @GET
    @Path("{orderId}/customer")
    public Response getCustomerFromOrder(@PathParam("orderId") long orderId) {
        Order order = ordersBean.getOrderById(orderId);

        if (!serviceUrl.isPresent()) {
            throw new JavaServiceException("Service URL not found!", 404);
        }

        WebTarget apiUrl = serviceUrl.get().path("v1/customers/" + order.getCustomerId());

        Response response = apiUrl.request().get();

        if (response.getStatus() == 200) {
            CustomerResponse customerResponse = response.readEntity(CustomerResponse.class);
            return Response.status(Response.Status.OK).entity(customerResponse).build();
        } else {
            throw new JavaServiceException("Service returned error status code: " + response.getStatus(), 500);
        }

    }
    
    // create new order
    @POST
    public Response createOrderForCustomer(OrderRequest newOrder) {
        Order order = new Order();
        order.setCustomerId(newOrder.getCustomerId());
        order.setTitle(newOrder.getTitle());
        order.setDescription(newOrder.getDescription());
        
        ordersBean.createOrder(order);
        
        return Response.status(Response.Status.CREATED).entity(order).build();
    }
}
```

## Create Go project

Now we will create our Go service. It will be built using [Gin web framework](http://github.com/gin-gonic/gin) and [Sling HTTP client library](http://github.com/dghubble/sling). 

If you have your own preferred packages for web framework and HTTP client, you can of course use them.

First, create a new directory in your Go Workspace (for example, `$GOPATH/src/go-service`) that will serve as a root folder for our go service.

Within Go workspace, perform `go get` to retrieve kumuluzee and other mentioned packages:

```
$ go get github.com/kumuluz/kumuluzee-go-config/config
$ go get github.com/kumuluz/kumuluzee-go-discovery/discovery
$ go get github.com/gin-gonic/gin
$ go get github.com/dghubble/sling
```

If you are using a dependency management tool for vendoring packages, you can of course use it. Here's an example using [dep](https://github.com/golang/dep) tool:
```
$ dep init
$ dep ensure github.com/kumuluz/kumuluzee-go-config/config
$ dep ensure github.com/kumuluz/kumuluzee-go-discovery/discovery
$ dep ensure github.com/gin-gonic/gin
$ dep ensure github.com/dghubble/sling
```

### Create necessary files

The Go service we will create is going to consist of three source files and a config file:
- *main.go*, which will contain main function and initialization functions
- *models.go*, which will contain structs for models we will operate with
- *handlers.go*, which will contain HTTP handlers of our REST API
- *config.yaml*, which will contain our microservice configuration

### Prepare required structs

First, in **models.go** file, we prepare the structs for Customer and Order request/response. These structs will be marshalled to and unmarshalled from json, therefore structs' fields must be exported, with json tags set up:

```go
type Customer struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    LastName string `json:"lastName"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
}

type OrderRequest struct {
    CustomerID  int64  `json:"customerId"`
    Title       string `json:"title"`
    Description string `json:"description"`
}

type OrderResponse struct {
    ID          int64  `json:"id"`
    CustomerID  int64  `json:"customerId"`
    Title       string `json:"title"`
    Description string `json:"description"`
}

type ErrorResponse struct {
    Status  int    `json:"status"`
    Message string `json:"message"`
}
```

Note that we will also use Customer struct to save customers in our mocked (in-memory) database.

### Create REST API server and prepare routes

Let's prepare the main function in our **main.go** file. The main function will:
1. initialize a mocked database (an array of Customer stucts),
2. initialize KumuluzEE config,
3. initialize KumuluzEE discovery and register itself to consul,
4. initialize and start REST API server.

```go

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

    // middleware that checks for maintenance mode parameter
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
```

### config.yaml file

In file **config.yaml**, write our configuration in the already known KumuluzEE format:
```yaml
kumuluzee:
  # service name
  name: go-service
  server:
    # url where our service will live
    base-url: http://localhost:9000
    http:
      port: 9000
  env:
    name: dev
  # specify hosts for discovery server
  discovery:
    consul:
      hosts: http://localhost:8500
  # specify hosts for config server
  config:
    consul:
      hosts: http://localhost:8500
# our custom configuration which will be registered in config server
rest-config:
  maintenance: false
```

### Initialize KumuluzEE config and discovery

Now we prepare initialization functions for KumuluzEE configuration and service discovery (in **main.go**).

```go
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
```

### Initialize database

We also prepare the initialization function for our database. For the sake of this example, we are simply going to store customers in an array. Usually, we would store such data into a database and then perform connection and queries to the database. There are many packages available for handling many different databases, so you can choose your preferred package to perform database operations.

```go
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
```

### Implement HTTP handlers

In file **handlers.go**, we write implementations of handlers used in router.

```go
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
```

## Run it

Now we can run both projects:

```
$ cd java-service
$ mvn clean package
$ java -jar ./api/target/api-1.0.0.jar
...
$ cd go-service
$ go build
$ ./go-service
```

And access their endpoints:
- Java service:
    - http://localhost:8080/v1/orders/ (since we used KumuluzEE REST, we can perform various queries, for example: http://localhost:8080/v1/orders?where=customerId:EQ:102)
    - http://localhost:8080/v1/orders/1
    - http://localhost:8080/v1/orders/1/customer
- Go service:
    - http://localhost:9000/v1/customers/
    - http://localhost:9000/v1/customers/102
    - http://localhost:9000/v1/customers/102/orders
    - http://localhost:9000/v1/customers/102/neworder

## Conclusion

In this tutorial we have used the KumuluzEE framework to build Java and Go services and make them communicate between themselves. We demonstrated how to register and discover services and how to read external configuration.

Source code can be found in GitHub repository.
 
