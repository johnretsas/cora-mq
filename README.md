## Queue Service

This service allows you to create and manage queues. You can create queues, enqueue items, and dequeue items from queues via HTTP endpoints.

### Requirements

- Go 1.16 or higher
- `go-queue-service/queue` package

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/go-queue-service.git
   cd go-queue-service
   ```

2. Install dependencies:
   ```sh
   go mod tidy
   ```

### Usage

#### Running the Server

1. Create a `main.go` file in the root of your project:

   ```go
   package main

   import (
       "go-queue-service/queue_server"
       "log"
       "net/http"
       "os"
   )

   func main() {
       logger := log.New(os.Stdout, "queue_server: ", log.LstdFlags)
       server := queue_server.NewQueueServer(logger)

       http.HandleFunc("/createQueue", server.CreateQueueHandler)
       http.HandleFunc("/dequeue", server.DequeueHandler)

       logger.Println("Starting server on :8080")
       if err := http.ListenAndServe(":8080", nil); err != nil {
           logger.Fatalf("Could not start server: %s\n", err)
       }
   }
   ```

2. Run the server:
   ```sh
   go run main.go
   ```

#### Endpoints

##### Create Queue

- **URL**: `/createQueue`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
      "name": "myQueue"
  }
  ```
- **Response**: `201 Created` on success

##### Dequeue Item

- **URL**: `/dequeue`
- **Method**: `POST`
- **Request Body**:
  ```json
  {
      "queueName": "myQueue"
  }
  ```
- **Response**: `200 OK` on success with the dequeued item
  ```json
  {
      "message": "Item dequeued",
      "item": {
          "ID": "itemID",
          "Payload": "itemPayload"
      }
  }
  ```

### Example

To create a queue and dequeue an item, you can use `curl` commands.

1. **Create a Queue**:
   ```sh
   curl -X POST http://localhost:8080/createQueue -H "Content-Type: application/json" -d '{"name": "myQueue"}'
   ```

2. **Dequeue an Item**:
   ```sh
   curl -X POST http://localhost:8080/dequeue -H "Content-Type: application/json" -d '{"queueName": "myQueue"}'
   ```

### Note

Ensure that you have defined the `queue` package correctly in your project with `Queue` and `QueueItem` structures and their associated methods.

### License

This project is licensed under the MIT License. See the `LICENSE` file for details.

### Contributing

If you would like to contribute, please open a pull request or issue on GitHub.

With these instructions, your `README` now includes comprehensive information on how to use the `DequeueHandler` to dequeue items from a queue.