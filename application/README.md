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
    -   ✅ OK
    -   ❌ Compile Error
    -   ❌ Wrong Answer
    -   ❌ Memory Limit Exceeded
    -   ❌ Time Limit Exceeded
    -   ❌ Runtime Error

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
  user: your_username
  password: your_password
  dbname: online_judge
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
export DB_USER=your_username
export DB_PASSWORD=your_password
export DB_NAME=online_judge

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
