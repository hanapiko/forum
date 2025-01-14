# forum
# Forum Web Application

## Project Overview

This is a web forum application built to facilitate communication, content sharing, and user interaction through posts, comments, and engagement features.

## Key Features

### 1. User Authentication
- User registration with email, username, and password
- Secure login sessions using cookies
- Password encryption
- Unique user identification

### 2. Communication
- Create posts and comments
- Associate categories with posts
- Visibility of posts and comments for all users

### 3. Interaction
- Like and dislike posts and comments (registered users only)
- Display of like/dislike counts

### 4. Filtering Capabilities
- Filter posts by:
  - Categories
  - User's created posts
  - User's liked posts

## Technical Stack

- **Backend**: Go
- **Database**: SQLite
- **Containerization**: Docker

## Prerequisites

- Go (latest version)
- Docker
- SQLite3

## Installation

1. Clone the repository
```bash
git clone https://github.com/yourusername/forum.git
cd forum
```

2. Build Docker container
```bash
docker build -t forum-app .
```

3. Run the application
```bash
docker run -p 8080:8080 forum-app
```

## Development Setup

### Dependencies
- `sqlite3`
- `bcrypt`
- `UUID`

### Running Tests
```bash
go test ./...
```

## Project Learning Objectives

- Web development basics (HTML, HTTP)
- Session and cookie management
- Docker containerization
- SQL database manipulation
- Basic encryption techniques

## Allowed Packages

- Standard Go packages
- `sqlite3`
- `bcrypt`
- `UUID`

## Restrictions

- No frontend libraries or frameworks (React, Angular, Vue, etc.)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Specify your license here (e.g., MIT License)

## Contact

Your Name - your.email@example.com

Project Link: [https://github.com/yourusername/forum](https://github.com/yourusername/forum)