# Online Code Judging System

## Project Overview

This project aims to build a simplified online judging system similar to platforms like Quera or Codeforces. Users can view questions, submit their code (in Go), and receive feedback on their submissions.

---

## Core Features

### Authentication, Login, and Registration

-   User registration and login functionality.
-   Secure password storage using `bcrypt` for one-way encryption.

### User Roles & Access Control

-   Two user roles: **Regular User** and **Admin**.
-   Admins can publish questions and manage user roles.
-   Access control enforced on both backend (API endpoints) and frontend (UI elements).

### Profile Page

-   Displays user details (username, submission statistics).
-   Stats include total attempted questions, success rate, and solved questions.
-   Admins can change user roles via the profile page.

### Question List

-   Displays published questions sorted by publish date (newest first).
-   Implements pagination (e.g., 10 questions per page) using query parameters.

### Questions

-   Each question has an **owner** (creator).
-   Questions start as **drafts** and require **admin** approval to be published.
-   Question details:
    -   Title
    -   Statement
    -   Time Limit (milliseconds)
    -   Memory Limit (MB)
    -   Test Input
    -   Expected Output
-   Regular users can submit solutions to published questions.

### Submissions

-   Users submit **Go (Golang)** code.
-   Initial status: "Pending Review".
-   Processed by a separate **judging service**.
-   Possible results:
    -   ✅ Accepted
    -   ❌ Compilation Error
    -   ❌ Wrong Answer
    -   ❌ Memory Limit Exceeded
    -   ❌ Time Limit Exceeded
    -   ❌ Runtime Error
-   Submission Processing:
    -   When a submission is made, it is marked as "Pending"
    -   The system retrieves the question's time and memory limits
    -   The code is executed in a Docker container with the specified limits
    -   The output is compared with the expected output
    -   The submission status is updated based on the result
    -   Results include execution time and memory usage
    -   Temporary files are cleaned up after processing

### Question & Submission Pages

-   Browse published questions and view details.
-   "Submit Answer" page for code submission.
-   "My Submissions" page listing user's submission history and results.

### Create Question Page

-   **Admins and Regular Users** can create draft questions.
-   Fields: title, statement, limits, test cases.
-   Users can edit their own draft questions (deletion is not required).
-   Admins can view all questions and manage publication status.

---

## Internal API for Submission Processing

-   A separate **runner** process handles submissions via internal APIs.
-   **Runner Responsibilities:**
    1.  Receives submission data (code, test cases, constraints).
    2.  Runs the submission against test cases.
    3.  Sends the result back to the main web service.
-   **Concurrency Management:**
    -   Ensure a submission is processed by only one runner at a time.
    -   Handle runner failures/timeouts by reassigning the submission.
    -   Mark submissions that repeatedly fail as errors and remove them from the queue.
    -   Use **database transactions** (e.g., `SELECT ... FOR UPDATE`) for safe concurrent access.

---

## Project Commands & Execution

### Main Executable Commands

1.  **`serve`**
    -   Starts the HTTP web server.
    -   Reads configuration from a file (using **Viper**).
    -   Accepts flags like `--listen :8080`.

2.  **`code-runner`**
    -   Compiles and runs submitted Go code.
    -   Uses **Docker** for secure execution (sandboxing CPU, memory, network).

3.  **`create-admin`**
    -   CLI command to create a new admin user or upgrade an existing user to admin.

---

## Project Structure

A recommended structure for this project:

```
online-judge/
├── cmd/                  # Main application(s) entry points
│   ├── server/           # Entry point for the 'serve' command (web server)
│   │   └── main.go
│   ├── runner/           # Entry point for the 'code-runner' command
│   │   └── main.go
│   └── create-admin/     # Entry point for the 'create-admin' utility
│       └── main.go
│
├── internal/             # Private application logic
│   ├── auth/
│   ├── config/
│   ├── database/
│   ├── handler/
│   ├── middleware/
│   ├── models/
│   ├── queue/
│   └── runner/
│
├── web/                  # Frontend related files
│   ├── templates/
│   └── static/
│
├── configs/              # Configuration files (e.g., config.yaml)
├── migrations/           # Database migration files
├── scripts/              # Helper scripts (e.g., seeding)
│
├── .gitignore
├── go.mod
├── go.sum
├── Dockerfile            # Dockerfile for 'serve'
├── Dockerfile.runner     # Dockerfile for 'code-runner'
├── docker-compose.yml
├── CODING_RULES.md       # Project coding standards
└── README.md
```

*   **`/cmd`**: Main application entry points.
*   **`/internal`**: Core application logic, not intended for external import.
*   **`/web`**: Frontend assets (templates, static files).
*   **`/configs`**: Configuration files.
*   **`/migrations`**: Database migration scripts.
*   **`/scripts`**: Utility scripts.

---

## Group Work & Implementation Phases

### Phase 1: Basic Frontend (Minimal UI)

-   Use **Golang templating** for rendering.
-   Implement pages: Homepage, Login/Signup, Question List (Paginated), Question Details/Submission Form, User Submissions History, Profile Page.

### Phase 2: Authentication

-   Implement user login/logout.
-   Store passwords using **bcrypt**.
-   Manage sessions using **cookies or JWT tokens**.

### Phase 3: Database Design

-   Define tables: `Users`, `Questions`, `Submissions`.
-   Optimize with **indexes** and caching where applicable.

### Phase 4: Backend Implementation

-   Implement core API endpoints for:
    -   Question CRUD (Create, Read, Update - Drafts only by owner/admin) & Publishing (Admin only).
    -   Code submission.
    -   Viewing submissions and results.
    -   Admin user management.

### Phase 5: Judging Service & API Integration

-   Implement the **`code-runner`** service.
-   Integrate the main server with the runner via internal APIs.
-   *Optional:* Implement compiled-code caching.

### Phase 6: Deployment with Docker

-   Use **Docker & Docker Compose** to manage: Web service, Database, Judge runner.
-   Expose only necessary ports (e.g., web server on port 80).
-   Ensure the judge service does **not** have direct database access.

### Phase 7: Database Seeding & Load Testing

-   Seed the database with test data (users, questions, submissions).
-   Perform load testing with multiple concurrent judge runners.
-   Optimize database queries and indexes based on testing.

---

## Final Submission Requirements

-   **GitHub Repository:** Containing all source code.
-   **ZIP File:** Including:
    -   Project code.
    -   Team member details (name, student ID).
    -   Instructions for running the project (`README.md`).
    -   Load testing scripts/tools used.
    -   Link to the GitHub repository.

mkdir -p cmd/server
mkdir -p internal/database
mkdir -p configs
mkdir -p migrations

## Database Setup

### Initial Setup

1. Make sure you have PostgreSQL installed and running on your system.

2. Create a new database:
```bash
createdb online_judge
```

3. Apply the database migrations:
```bash
psql -d online_judge -f migrations/000001_init_schema.up.sql
```

4. (Optional) Seed the database with sample data:
```bash
psql -d online_judge -f scripts/seed.sql
```

### Sample Data

The seeding script (`scripts/seed.sql`) includes:

- Users:
  - Admin user: `admin` / `admin@example.com`
  - Regular users: `user1` and `user2`
  - Note: The password hashes in the seed file are placeholders. In a real environment, you should use proper password hashing.

- Questions:
  - "Hello World" (published)
  - "Sum of Two Numbers" (published)
  - "Factorial" (draft)

- Test cases for each question
- Sample submissions with different results

### Configuration

Create a `config.yaml` file in the project root with your database configuration:

```yaml
database:
  host: localhost
  port: 5432
  user: mahdi
  password: secret123
  dbname: online-judge
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
  connect_timeout: 5
```

## Secure Configuration Handling

### Configuration Setup

1. Copy the template configuration file:
```bash
cp configs/config.yaml.template config.yaml
```

2. Edit the `config.yaml` file with your specific settings:
   - Replace all placeholder values with your actual configuration
   - Never commit this file to version control
   - Keep it secure and restrict access to it

### Environment Variables (Alternative)

For additional security, you can use environment variables instead of the config file:

```bash
# Database configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=mahdi
export DB_PASSWORD=secret123
export DB_NAME=online-judge

# Server configuration
export SERVER_LISTEN=":8080"
export SERVER_SECRET_KEY="your-secret-key-here"

# Runner configuration
export RUNNER_MAX_CONCURRENT=5
export RUNNER_TIMEOUT=30s
export RUNNER_MEMORY_LIMIT_MB=256
export RUNNER_CPU_LIMIT=1
```

### Security Best Practices

1. **Never commit sensitive data**:
   - Keep `config.yaml` out of version control
   - Use `.gitignore` to prevent accidental commits
   - Consider using environment variables for sensitive data

2. **File permissions**:
   - Set appropriate file permissions (e.g., 600 for config files)
   - Restrict access to configuration files

3. **Production deployment**:
   - Use different configuration files for development and production
   - Consider using a secrets management service in production
   - Use environment variables or secure vaults for sensitive data

4. **Password security**:
   - Use strong, unique passwords
   - Consider using password managers
   - Rotate passwords regularly

5. **Database security**:
   - Use SSL/TLS for database connections in production
   - Implement proper access controls
   - Regularly backup your database

### Database Configuration

The application is configured to connect to a local PostgreSQL database running in a Docker container. The connection settings are defined in `configs/config.yaml`:

```
database:
  host: localhost
  port: 5432
  user: mahdi
  password: secret123
  dbname: online-judge
  sslmode: disable
```

- **host**: `localhost` (use `host.docker.internal` if running the app inside another Docker container)
- **port**: `5432` (default Postgres port)
- **user**: `mahdi` (as provided)
- **password**: `secret123` (as provided)
- **dbname**: `online-judge` (as provided)
- **sslmode**: `disable` (recommended for local development)

**Example Docker command to run Postgres:**
```
sudo docker run --name online-judge -e POSTGRES_PASSWORD=secret123 -e POSTGRES_DB=online-judge -e POSTGRES_USER=mahdi -p 5432:5432 -d postgres:15
```

**Troubleshooting:**
- If you get connection errors, ensure the Postgres container is running (`sudo docker ps`).
- Make sure the credentials in `config.yaml` match your Docker environment variables.
- If running the app inside a Docker container, use `host.docker.internal` for the `host` value.
- The database `online-judge` must exist (created by the `POSTGRES_DB` env variable above).

## Code Runner Architecture (Isolated Docker Execution)

### Overview
Each code submission is executed in a dedicated Docker container (a "code runner") to ensure security, resource isolation, and scalability. The main application communicates with the code runner via an internal API, sending the submitted code, test case input, and resource limits (memory, time).

### Workflow
1. **Submission:**
   - User submits code via the question answer form.
   - The main application collects the code, test case input, expected output, and the question's memory/time limits.

2. **Runner Initialization:**
   - The main app launches a new Docker container for each submission using the Docker client.
   - The container is started with:
     - **CPU limit:** 1 core
     - **Memory limit:** as specified by the question
     - **Time limit:** enforced by the runner logic
     - **No network access**
     - **No access to the application database or host file system**

3. **Code Execution:**
   - The code and input are sent to the container (e.g., via mounted files or stdin).
   - The runner compiles (if needed) and executes the code inside the container.
   - The runner enforces the time and memory limits, terminating the process if exceeded.

4. **Result Collection:**
   - The runner captures the program's output and exit status.
   - The output is sent back to the main application via an internal API response.

5. **Validation:**
   - The main application compares the runner's output to the expected output for the test case.
   - The result (OK, Wrong Answer, Time Limit Exceeded, Memory Limit Exceeded, etc.) is recorded and shown to the user.

### Security & Isolation
- **Each code runner is a short-lived Docker container.**
- **No network, DB, or host FS access** is allowed for the container.
- **Resource limits** (CPU, memory, time) are strictly enforced.
- Containers are terminated after execution to prevent lingering processes.

### Example Docker Run Command
```
docker run --rm \
  --cpus=1 \
  --memory=128m \
  --network=none \
  -v /tmp/code:/code:ro \
  code-runner-image:latest
```

- Replace `128m` with the question's memory limit.
- The code and input can be provided via a mounted volume or sent to the container at runtime.

### Internal API
- The main app and code runner communicate via an internal API (e.g., HTTP on localhost, or via Docker exec/stdin/stdout).
- The runner returns the program output, exit code, and resource usage.

### Extensibility
- This architecture allows for easy scaling (multiple runners in parallel) and supports additional languages by building new runner images.
