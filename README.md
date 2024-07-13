# Queue Server Package

The `queue_server` package provides HTTP handlers for managing queues and enqueuing items into them. This package leverages the `go-queue-service/queue` package for the underlying queue data structure.

## Handlers

### CreateQueueHandler

**Endpoint**: `/create_queue`  
**HTTP Method**: POST

#### Description

Creates a new queue with the specified name.

#### Request Format

```json
{
    "name": "myQueue"
}
```

- **name**: (string) The name of the queue to be created.

#### Response

- **HTTP Status**: 201 Created if successful.
- **JSON Response Body**: 
  ```json
  {
      "message": "Queue created",
      "name": "myQueue"
  }
  ```

#### Errors

- If the HTTP method is not POST, it returns a `405 Method Not Allowed` error.
- If there is an error decoding the JSON payload, it returns a `400 Bad Request` error with details.

### EnqueueHandler

**Endpoint**: `/enqueue`  
**HTTP Method**: POST

#### Description

Enqueues an item into the specified queue.

#### Request Format

```json
{
    "queueName": "myQueue",
    "item": {
        "ID": "unique-id-123",
        "Payload": "some payload data"
    }
}
```

- **queueName**: (string) The name of the queue where the item should be enqueued.
- **item**: (object) The item to be enqueued, containing an ID (`string`) and Payload (`string`).

#### Response

- **HTTP Status**: 201 Created if successful.
- **JSON Response Body**: 
  ```json
  {
      "message": "Item enqueued",
      "id": "unique-id-123"
  }
  ```

#### Errors

- If the HTTP method is not POST, it returns a `405 Method Not Allowed` error.
- If there is an error decoding the JSON payload, it returns a `400 Bad Request` error with details.
- If `queueName` is missing or empty, it returns a `400 Bad Request` error.

## Usage

To use these handlers, make HTTP POST requests to the respective endpoints with the appropriate JSON payload as described above.

### Example Usage with curl

1. **Create a Queue**
   ```bash
   curl -X POST http://localhost:8080/create_queue -d '{"name": "myQueue"}' -H 'Content-Type: application/json'
   ```

2. **Enqueue an Item**
   ```bash
   curl -X POST http://localhost:8080/enqueue -d '{"queueName": "myQueue", "item": {"ID": "unique-id-123", "Payload": "some payload data"}}' -H 'Content-Type: application/json'
   ```

Replace `http://localhost:8080` with the actual URL of your server.