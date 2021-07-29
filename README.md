# Solution

The solution was proposed in a set of different microservices in order to comply with the technical requirements and business constraints.
Since it cannot miss any callback request (arrow 1), I opted for adding a message broker (RabbitMQ) to persist every callback so no data
is lost in case any component crashes. By dividing it into producer and consumer we have more flexibility to scale the application to deal
with high throughput (since they work concurrently) and satisfy high availability.

<p align="center">
  <img height="300" src="https://user-images.githubusercontent.com/1853854/127479375-01433585-5dfd-49e3-aef9-932666e5ccf8.png">
</p>

## Pipeline
The data flow described by the arrows in the picture above can be described as such:

1. `Tester` sends a request to `Producer` on the callback endpoint with a list of IDs
   
2. `Producer` splits the IDs into individual messages and publish them on the `Queue` asynchronously
   
3. Each `Consumer` will fetch a message from the queue with an ID
   
4. The `Consumer` then fetches the object status calling `"/objects/:id"` endpoint of the `Tester`
   
5. `Tester` responds to the request with the object status
   
6. If the status is online the `Consumer` updates the ID's `last_seen` timestamp on the `Database`
   
7. After the data is persisted to PostgreSQL, the `Consumer` acknowledges the message to RabbitMQ to be removed from the queue.
   Otherwise, if there was any error and the object status was not saved on the database, the message would
   return to the queue and be consumed again.
   
8. In order to remove old entries from the database (30 seconds or more), there is a `Routine` executed every 3 seconds
    that queries the database to remove such entries. Such frequency can be changed using the `ROUTINE_INTERVAL` parameter.
   
## Parameters

The `.env` file includes the environment variables necessary to run the services.
The `docker-compose.yml` file includes the configuration for the deployment of each individual service.
Some considerations:

* `ROUTINE_INTERVAL` sets how often the routine to delete old entries from the database should be performed (default is every 3 seconds).

* Since all containers are connected through a docker network, the hosts of the endpoints are the name of each service.
For instance, the database host is "db", the callback API host is "producer" and the object status API host is "tester".
For this matter, the API endpoint consumed by `tester_service.go` was changed from `http://localhost:9090/callback` to `http://producer:9090/callback`.

* Since the consumer depends on an API that can take up to 4 seconds to respond, I implemented it with concurrency given by the parameter `MAX_CONSUMERS`.
That means that with `MAX_CONSUMERS` equal to 100, a single container will have up to 100 workers (goroutines) running in parallel, 
each consuming a message from the queue and sending a request to the API independently. 
This way we can decrease the impact of the API's slowness.


## Logging

Logging was implemented using the `sirupsen/logrus` module. This way we get information of errors that might happen
to the application and also we can use debugging messages to monitor each service's operation setting the log level to `DebugLevel`.

## How to run

The `docker-compose.yml` includes the configuration to run all the services:
RabbitMQ, PostgreSQL, producer, 2 consumers with different `MAX_CONSUMERS` values, tester and the routine.
To run, simply execute the following command at the root of the project:
```
docker-compose up -d
```

You can connect to the database on `localhost:5432` (connection information in the `.env` file) to check the objects and the last time they were seen.
RabbitMQ also provides a user interface where you can check queue statistics, such as queue size, performance and how many consumers are connected.
You can access it on `http://localhost:15672/`. The default username and password is `guest`.
The picture below presents an example of RabbitMQ's UI.

<p align="center">
    <img height="300" src="https://user-images.githubusercontent.com/1853854/127513521-7ea198a1-4ee4-4184-8447-0326d61e9610.png">
</p>

## Improvements

Due to time constraints and to avoid high code complexity, there are a few improvements that were not implemented in this
solution but are worth to be mentioned since they would improve the application performance in a real scenario.

* **Use delayed retry policy**: when a service can't connect to the queue or the database, it will keep trying infinitely
until a connection is established. It could be more interesting to add delays between such retries and define a hard limit
to crash the application in case no connection was established after X tries.

* **Use Kubernetes**: with it, we could implement load balancing (different pods for a single service),
  automatic scaling strategies and automatic recreate pods after a crash (to ensure high availability).
  Also, the routine could be implemented as a Kubernetes' CronJob.

* **Use cloud services**: RabbitMQ could be substituted by as GCP's Pub/Sub (or similar services from other providers)
  and the same would be possible for a cloud service database instead of a container with PostgreSQL in it.
  By adopting a cloud service we would have to worry less about defining scaling strategies.
  
* **Use Redis instead of PostgreSQL**: the need for a routine to exclude old entries from the database could be avoided
  if we used a key-value database that implements TTL. Redis is an example. 

* **Testing**: this solution does not include any test suit, which is 100% important to assure code quality,
  easier maintenance and further development. I did not write test cases due to time constraints and also because
  the assessment did not require it, but I am a big follower of TDD.
  
* **Project structure**: all services were included in the same branch of the repository but ideally it could be divided
so we can apply DDD into the code organization and project structure.


