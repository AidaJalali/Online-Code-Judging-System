# Clean Code Rules for Online Judging System (Go)

This document outlines the coding standards and principles to follow for consistency, readability, and maintainability in this project.

## 1. Formatting & Naming

*   **Rule 1.1:** Always run `go fmt` or `goimports` on your code before committing. *Consistency is key.*
*   **Rule 1.2:** Use clear, descriptive names for variables, functions, types, and packages. (`userService` is better than `us`, `getUserByID` is better than `getU`).
*   **Rule 1.3:** Follow Go conventions for naming (e.g., `CamelCase` for exported identifiers, `camelCase` for internal). Acronyms like `HTTP`, `ID`, `URL` should usually be all caps (`ServeHTTP`, `UserID`, `BaseURL`).

## 2. Simplicity & Focus

*   **Rule 2.1:** Functions/methods should do one thing well. Aim for functions under 20-30 lines; refactor if significantly longer.
*   **Rule 2.2:** Minimize nesting depth (if/for). If nesting goes beyond 2-3 levels, consider refactoring using helper functions or different control structures.
*   **Rule 2.3:** Avoid unnecessary complexity or "clever" code that is hard to understand. Prefer straightforward implementations.

## 3. Error Handling

*   **Rule 3.1:** Never ignore errors. Check `err` after any operation that returns one (`if err != nil { ... }`).
*   **Rule 3.2:** Handle errors immediately or return them clearly. Add context to errors using `fmt.Errorf("... %w", err)` when returning them up the stack.
*   **Rule 3.3:** Use `errors.Is` and `errors.As` for specific error checking instead of string comparison.

## 4. Package Design & Dependencies

*   **Rule 4.1:** Keep packages focused. Group related functionality (e.g., all database interaction for users in `internal/database/user_repo.go`).
*   **Rule 4.2:** Only export identifiers (start with Uppercase) that are essential for other packages to use. Keep internal details private (lowercase).
*   **Rule 4.3:** Avoid circular package dependencies. If found, restructure your code (often by introducing interfaces or moving logic).
*   **Rule 4.4:** Keep `main` packages (`cmd/*`) thin. They should primarily parse flags/config and wire together components from `internal`.

## 5. Interfaces

*   **Rule 5.1:** Define interfaces based on *required behavior* (what a consumer needs), not based on the implementation.
*   **Rule 5.2:** Accept interfaces as parameters in functions/methods where you need flexibility or decoupling (especially for dependencies like database repositories). Return concrete structs.

## 6. Concurrency

*   **Rule 6.1:** Use goroutines only when true concurrency is needed (e.g., handling multiple requests, background tasks like the runner).
*   **Rule 6.2:** Protect access to shared mutable state using `sync.Mutex` or prefer communication via channels.
*   **Rule 6.3:** Always test concurrent code with the `-race` flag (`go test -race ./...`).

## 7. Testing

*   **Rule 7.1:** Write unit tests for core logic, especially in the `internal` packages.
*   **Rule 7.2:** Test the exported functions/methods of your packages.
*   **Rule 7.3:** Use table-driven tests for testing functions with multiple input/output scenarios.

## 8. Dependencies & Globals

*   **Rule 8.1:** Minimize global variables. Pass dependencies explicitly (e.g., pass a database connection pool or logger to structs/functions that need them).
*   **Rule 8.2:** Define dependencies using interfaces where possible (see Rule 5.2).

## 9. Comments

*   **Rule 9.1:** Write comments to explain the *why*, not the *what*, unless the code is unavoidably complex.
*   **Rule 9.2:** Add `godoc` comments to all exported identifiers (packages, types, functions, constants, variables).

## 10. Documentation (README)

*   **Rule 10.1:** Keep the main `README.md` file accurate and up-to-date.
*   **Rule 10.2:** After implementing significant changes (new features, structural changes, command updates, setup modifications), update the relevant sections of the `README.md` to reflect these changes. 