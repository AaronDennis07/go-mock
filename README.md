# GoMock

A lightweight, file-based JSON REST API server for mocking and prototyping.

## Features

- Simple file-based data storage
- RESTful CRUD operations
- Zero-configuration setup
- Lightweight and fast
- Supports multiple collections

## Installation

```bash
# Install via go get
go install github.com/yourusername/go-mock@latest

# Or clone and build
git clone https://github.com/yourusername/go-mock.git
cd go-mock
go build
```

## Quick Start

1. Create a `db.json` file:
```json
{
  "posts": [
    {"id": 1, "title": "First Post"}
  ],
  "users": [
    {"id": 1, "name": "John Doe"}
  ]
}
```

2. Run the server:
```bash
go-mock -db=db.json -port=3000
```

## Usage

### Command-line Flags
- `-db`: JSON database file path (default: `db.json`)
- `-port`: Server port (default: `3000`)

### API Endpoints

- `GET /{collection}`: List all items
- `GET /{collection}/{id}`: Get specific item
- `POST /{collection}`: Create new item
- `PUT /{collection}/{id}`: Update existing item
- `DELETE /{collection}/{id}`: Delete item

### Examples

```bash
# Get all posts
curl http://localhost:3000/posts

# Create a new post
curl -X POST http://localhost:3000/posts \
     -H "Content-Type: application/json" \
     -d '{"title":"New Post"}'

# Update a post
curl -X PUT http://localhost:3000/posts/1 \
     -H "Content-Type: application/json" \
     -d '{"title":"Updated Post"}'

# Delete a post
curl -X DELETE http://localhost:3000/posts/1
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License