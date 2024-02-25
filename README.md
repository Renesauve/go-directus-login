# Go To Directus Login Flow

This project implements a login flow using Go and Directus, providing a secure way to handle user authentication and verification through email.

## Requirements

To run this project, you must have a Directus Docker environment set up.

## Environment Setup

### Directus Environment Variables

After setting up your Directus Docker environment, ensure you configure the environment variables for email functionality. Add the following variables to your Directus `.env` file, replacing placeholders with your actual values:

```
EMAIL_SMTP_HOST="smtp.example.com"
EMAIL_SMTP_PORT="587"
EMAIL_SMTP_USER="your-email@example.com"
EMAIL_SMTP_PASSWORD="yourpassword"
EMAIL_SMTP_SECURE="false" # true if using SSL
EMAIL_SMTP_IGNORE_TLS="false" # true to ignore TLS/SSL errors

EMAIL_FROM="your-email@example.com"
EMAIL_TRANSPORT="smtp"

ADMIN_ROLE_ID="your-admin-role-id"
```

### Go Application Environment Variables

For the Go application, create a .env file in the root of your Go project with the following variables:

```
DIRECTUS_URL="https://your-directus-instance.com"
DIRECTUS_EMAIL="your-directus-email@example.com"
DIRECTUS_PASSWORD="your-directus-password"
DIRECTUS_ADMIN_TOKEN="your-directus-admin-token"
DIRECTUS_USER_UUID="your-directus-user-uuid"
JWT_SECRET="your-jwt-secret"
```

## Using Air for Live Reloading

To streamline the development process with live reloading, use air. If you haven't installed air, you can do so by running:

`go get -u github.com/cosmtrek/air`

Or you can install it globally (if you're using Go Modules, which is recommended):

`go install github.com/cosmtrek/air@latest`

To run your project with air, navigate to your project directory and execute:

`air`

This will automatically reload your Go application whenever a file change is detected, making development faster and more efficient.
