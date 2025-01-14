# go-web
[![Go](https://github.com/javitab/go-web/actions/workflows/go.yml/badge.svg)](https://github.com/javitab/go-web/actions/workflows/go.yml)

# Description
This repository acts as a template for creating a Go project involving a database/api/cli/web view based application. Includes basic interfaces for maintaing user database and interactions with LDAP for SSO.

# Configuration

This project requires two .env files: One in the project root, and another in `.devcontainer/` for development purposes only.

## Project .env file

For the general application, an .env file in the project root is required with the below information.

```bash
DB_DSN="host=localhost user=go-web password=password dbname=go-web port=5432 TimeZone=America/New_York"
SECRET_JWT_KEY="secret for JWT tokens"
HTTP_PORT=8080
HTTP_HOST="localhost:8080" # Require appropriate HTTP_HOST to prevent MITM attacks
LDAP_BIND_CREDENTIALS="base64 user:pass"
LDAP_ADDRESS="ldap://ldap.server.com"
LDAP_BASE_DN="DC=SERVER,DC=COM"
```

Can also include API key for CLI login passthrough:
```bash
CLI_API_KEY="UserWithCLILoginPermissionAPIKey"
```

## Devcontainer .env file

The below should be in an .env file in the `.devcontainer/` folder. Ensure the information in this file is consistent with the information in the `DB_DSN` field in the main project .env file above.

```bash
POSTGRES_USER=go-web
POSTGRES_PASSWORD=password
POSTGRES_DB=go-web
POSTGRES_HOSTNAME=localhost
```

# Web API Documentation (Swagger/OpenAPI)

All Web APIs are documented in the Swagger documentation endpoint:

`http://localhost:8080/swagger/index.html`

For additional information about adding annotations to API routes, see `swaggo/swag` on [github](https://github.com/swaggo/swag).


# Entrypoint

Application can be launched via cli with below summary of arguments:

Replace references to ./go-web to appropriate entrypoint for phase in development:

For development: 
```bash

# To start directly:
go run main.go <mode> <menu>

# To generate swagger docs and start web server in debug mode:
./start_debug
```

For executing a build release:
```bash
./go-web <mode> <menu>
```
Help Info:

```bash
$ go run main.go
No arguments provided

Printing Available CLI Arguments
./go-web <mode> <menu>

To run in webserver mode:
     For production mode: ./go-web
     For debug mode: ./go-web web debug

To access auth utility menu: ./go-web util auth 
   Available auth menu utilities: 
        Get LDAP User Info
        Load Groups and Security Points from file
        Create User
        LDAP User Login Test
        Soft Delete User
        Evaluate User Security Points
        Standard User Login Test
        Get User Info
        Remove User From Group
        Update Security Points
        Set LDAP User
        Change User Password
        Add User to Group
        Get Group Info

```
