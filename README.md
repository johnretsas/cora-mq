# Queue Server App

The Queue Server App is a simple HTTP server application built in Go that allows you to manage multiple queues via HTTP requests. Each queue can store items defined by a unique identifier (`ID`) and a payload (`Payload`).

## Features

- **Create Queue**: Create new queues dynamically.
- **Enqueue Item**: Add items to a specific queue.
- **Dequeue Item**: Remove items from a specific queue.

## Technologies Used

- **Go**: Programming language used for server-side logic.
- **HTTP**: Handling HTTP requests and responses.
- **JSON**: Communication format for exchanging data.

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/your_username/queue-server-app.git
   cd queue-server-app
   ```

2. Build the application:

   ```bash
   go build
   ```

3. Run the server:

   ```bash
   ./queue-server-app
   ```

   By default, the server will run on `http://localhost:8080`.

## API Endpoints

- **Create Queue**
  - Endpoint: `POST /create-queue`
  - Body: `{"name": "queue_name"}`
  - Description: Creates a new queue with the specified name.

- **Enqueue Item**
  - Endpoint: `POST /enqueue?queue=queue_name`
  - Body: `{"ID": "item_id", "Payload": "item_payload"}`
  - Description: Adds an item to the specified queue.

- **Dequeue Item**
  - Endpoint: `GET /dequeue?queue=queue_name`
  - Description: Removes and returns an item from the specified queue.

## Example Usage

### Creating a Queue

```bash
curl -X POST -H "Content-Type: application/json" -d '{"name": "my_queue"}' http://localhost:8080/create-queue
```

### Enqueuing an Item

```bash
curl -X POST -H "Content-Type: application/json" -d '{"ID": "1", "Payload": "Hello, World!"}' http://localhost:8080/enqueue?queue=my_queue
```

### Dequeuing an Item

```bash
curl -X GET http://localhost:8080/dequeue?queue=my_queue
```

## Logging

- Server logs are written to standard output (stdout).
- Logging includes creation of queues, enqueueing and dequeuing of items, and errors encountered during requests.

## Error Handling

- HTTP status codes and JSON error messages are used for error handling.
- Common errors include invalid HTTP methods, missing parameters, and queue not found.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Adjust the URLs, example commands, and descriptions based on your actual implementation and requirements. This README provides a basic structure to help users understand how to install, use, and interact with your Queue Server App.
