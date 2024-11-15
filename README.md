# AskMe Microservices

AskMe is a microservices-based backend that uses **Go**, **Gin**, **MongoDB**, and **SQLite** for different services, with **gRPC** for communication between the services and an **API Gateway** for routing and managing requests.
This project is developed to function similarly to Stack Overflow, providing a platform where users can register, log in, and interact with a community by posting questions, answering them, and engaging in discussions. The **User Service** handles user management, including registration, login, and profile updates, while the **Content Service** manages questions, answers, tags, and user feeds. With features such as question search, flagging, upvoting/downvoting answers, and administrative controls for banning/unbanning users, the platform facilitates a seamless experience for both regular users and administrators, fostering a collaborative environment for knowledge sharing.
## Project Structure

The project consists of the following services:

- **API Gateway**: Central point for routing requests to different services and handling common functionalities like authentication, rate limiting, etc.
- **User Service**: Manages user-related operations, including registration, login, and profile management using **SQLite** for storage.
- **Content Service**: Manages content operations like posting and viewing questions and answers, using **MongoDB** for storage.
- **Admin Service**: Manages administrative operations like banning/unbanning users and managing flagged content.

---

## Technologies Used

- **Go**: The main programming language used across all services.
- **Gin**: Web framework used for routing and handling HTTP requests.
- **MongoDB**: Used for content management in the **Content Service**.
- **SQLite**: Used for user data storage in the **User Service**.
- **gRPC**: For communication between services.
- **JWT**: Used for authentication and authorization.

---

## Service Details

### 1. **API Gateway**

The **API Gateway** is the entry point for all external requests and routes them to the appropriate services.

#### Key Features
- Centralized routing for all services.
- JWT authentication middleware for user routes.
- Calls gRPC services for user and content management.
  
#### Dependencies
- `Go`
- `Gin`
- `gRPC`

#### Endpoints
- **User Routes**: 
  - `POST /register`: Register a new user.
  - `POST /login`: Log in an existing user.
  - `GET /profile`: Get user profile (JWT auth required).
  - `PATCH /update-profile`: Update user profile (JWT auth required).
  
- **Content Routes**: 
  - `GET /questions/search`: Search for questions.
  - `GET /questions/tags`: Get questions by tags.
  - `POST /question`: Post a new question (JWT auth required).

### 2. **User Service**

The **User Service** handles user-related operations like registration, login, and profile management. It uses **SQLite** for storing user data.

#### Key Features
- User registration and login.
- User profile management.
- JWT-based authentication for protected routes.

#### Dependencies
- `Go`
- `Gin`
- `SQLite`
- `JWT`

#### Endpoints
- `POST /register`: Register a new user.
- `POST /login`: Log in a user.
- `GET /profile`: Get the user profile (JWT auth required).
- `PATCH /update-profile`: Update the user profile (JWT auth required).

---

### 3. **Content Service**

The **Content Service** is responsible for managing the content of the platform, such as questions and answers. It uses **MongoDB** to store and manage the content.

#### Key Features
- Post and delete questions and answers.
- Search questions by tags and keywords.
- Get user feeds.
  
#### Dependencies
- `Go`
- `Gin`
- `MongoDB`
- `gRPC`

#### Endpoints
- **Public Routes**: 
  - `GET /questions/search`: Search for questions.
  - `GET /questions/tags`: Get questions by tags.
  - `GET /question/{id}`: Get question by ID.
  - `GET /feed`: Get the user feed.

- **Authenticated Routes**:
  - `POST /question`: Post a new question (JWT auth required).
  - `DELETE /question`: Delete a question (JWT auth required).
  - `POST /answer`: Post an answer (JWT auth required).
  - `DELETE /answer`: Delete an answer (JWT auth required).

---

### 4. **Admin Service**

The **Admin Service** provides administrative capabilities like banning users, unbanning them, and managing flagged content.

#### Key Features
- Ban/unban users.
- View and manage flagged questions and answers.

#### Dependencies
- `Go`
- `Gin`
- `gRPC`

#### Endpoints
- `GET /answers/flagged`: Get all flagged answers.
- `GET /questions/flagged`: Get all flagged questions.
- `POST /user/unban`: Unban a user.
- `POST /user/ban`: Ban a user.
