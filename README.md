 User Management API

A RESTful User Management API built with Go, Gin, and MongoDB.

 Features

- User CRUD operations (Create, Read, Update, Delete)
- JWT-based authentication
- Login endpoint
- Protected routes using authentication middleware
- File upload support
- Request validation
- Swagger API documentation

 Tech Stack

- Language: Go
- Web Framework: Gin
- Database: MongoDB
- Authentication: JWT
- API Documentation: Swagger

 Project Structure


.
├── database/
│   └── mongodb.go
├── date/
│   └── date.go
├── docs/
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── file/
│   └── upload.go
├── token/
│   └── token.go
├── project.go
├── go.mod
└── go.sum


 Getting Started

 Prerequisites

- Go
- MongoDB

 Installation

Clone the repository:


git clone https://github.com/fatimadpbrio/go-user-management-api.git
cd go-user-management-api


Install dependencies:


go mod download


 Configuration

Set the required environment variables before running the application.

Configure your MongoDB connection and JWT secret according to your local environment. Do not commit credentials or secret keys to the repository.

 Run the Application


go run project.go


 API Documentation

Swagger documentation files are available in the docs/ directory.

 Purpose

This project was developed to practice backend development with Go, RESTful API design, MongoDB integration, authentication, and file handling.

## Author

Fatima
