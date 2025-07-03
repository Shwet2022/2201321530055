# 2201321530055
Full Stack Developer Assessment Submission
This repository contains the submission for the Full Stack Developer assessment. The project is a complete URL shortener application, including a backend microservice and a frontend web application.

Project Structure
The repository is organized into three main directories as per the submission guidelines:

/
|-- Logging-Middleware/       # Contains the Go-based logging middleware.

|-- Backend-Test-Submission/  # Contains the Go-based URL shortener microservice.

|-- Frontend-Test-Submission/ # Contains the React (TypeScript) frontend application.

1. Logging Middleware
Location: Logging-Middleware/

Technology: Go

Description: A simple, reusable logging middleware that outputs structured JSON logs to the console. This middleware is integrated into the main backend microservice to handle all request logging.

2. Backend: URL Shortener Microservice
Location: Backend-Test-Submission/

Technology: Go

Description: A RESTful microservice that provides the core functionality for creating, redirecting, and retrieving statistics for shortened URLs. It uses an in-memory map for data storage.

How to Run the Backend
Navigate to the Backend-Test-Submission directory:

cd Backend-Test-Submission

Initialize the Go module and fetch dependencies:

go mod init backend
go mod tidy

Run the application:

go run main.go

The server will start and listen on http://localhost:8080. All incoming requests and responses will be logged to the console in a structured JSON format.

API Endpoints
A. Create Short URL
Method: POST

Route: /shorturls

Description: Creates a new short URL.

Request Body:

{
  "url": "[https://www.example.com/a-very-long-url-to-shorten](https://www.example.com/a-very-long-url-to-shorten)",
  "validity": 60, // Optional: in minutes, defaults to 30
  "shortcode": "my-link" // Optional: custom shortcode
}

Success Response (201 Created):

{
  "shortLink": "http://localhost:8080/my-link",
  "expiry": "2025-07-03T13:30:00Z"
}

B. Redirect Short URL
Method: GET

Route: /{shortcode}

Description: Redirects to the original long URL and increments the hit counter.

Example: http://localhost:8080/my-link

Success Response: HTTP 302 Found redirection.

C. Retrieve Short URL Statistics
Method: GET

Route: /{shortcode}/stats

Description: Retrieves usage statistics for a given short URL.

Example: http://localhost:8080/my-link/stats

Success Response (200 OK):

{
  "originalUrl": "[https://www.example.com/a-very-long-url-to-shorten](https://www.example.com/a-very-long-url-to-shorten)",
  "createdAt": "2025-07-03T12:30:00Z",
  "expiresAt": "2025-07-03T13:30:00Z",
  "hits": 1
}

3. Frontend: URL Shortener Web App
Location: Frontend-Test-Submission/

Technology: React with TypeScript

Styling: Vanilla CSS

Description: A responsive, single-page web application that allows users to interact with the backend microservice. It provides a user-friendly interface for creating short URLs and viewing their statistics.

How to Run the Frontend
Navigate to the Frontend-Test-Submission directory:

cd Frontend-Test-Submission

Install the required dependencies:

npm install

Start the development server:

npm start

The application will open automatically in your web browser, typically at http://localhost:3000.

Features
URL Shortener Form: A clean interface to submit a long URL, an optional custom shortcode, and a validity period.

Statistics Viewer: A separate view to input a short URL and retrieve its usage statistics.

Responsive Design: The UI is designed to be fully functional and visually appealing on both desktop and mobile devices.

User Feedback: Provides clear loading states, success messages, and error notifications to enhance the user experience.
