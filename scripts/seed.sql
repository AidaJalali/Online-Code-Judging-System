-- Insert sample admin user
INSERT INTO users (username, email, password_hash, role) VALUES
('admin', 'admin@example.com', '$2a$10$X7J3QZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZ', 'admin');

-- Insert sample regular users
INSERT INTO users (username, email, password_hash, role) VALUES
('user1', 'user1@example.com', '$2a$10$X7J3QZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZ', 'regular'),
('user2', 'user2@example.com', '$2a$10$X7J3QZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZq3YQZ', 'regular');

-- Insert sample questions
INSERT INTO questions (title, statement, time_limit_ms, memory_limit_mb, status, owner_id) VALUES
('Hello World', 'Write a program that prints "Hello, World!"', 1000, 256, 'published', 1),
('Sum of Two Numbers', 'Write a program that takes two numbers as input and prints their sum', 1000, 256, 'published', 1),
('Factorial', 'Write a program that calculates the factorial of a given number', 1000, 256, 'draft', 2);

-- Insert test cases for Hello World
INSERT INTO test_cases (question_id, input, expected_output, is_sample) VALUES
(1, '', 'Hello, World!', true);

-- Insert test cases for Sum of Two Numbers
INSERT INTO test_cases (question_id, input, expected_output, is_sample) VALUES
(2, '5 7', '12', true),
(2, '10 20', '30', false),
(2, '-5 5', '0', false);

-- Insert test cases for Factorial
INSERT INTO test_cases (question_id, input, expected_output, is_sample) VALUES
(3, '5', '120', true),
(3, '0', '1', false),
(3, '10', '3628800', false);

-- Insert sample submissions
INSERT INTO submissions (user_id, question_id, code, status, result, execution_time_ms, memory_usage_mb) VALUES
(2, 1, 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello, World!")\n}', 'completed', 'ok', 10, 5),
(2, 2, 'package main\n\nimport "fmt"\n\nfunc main() {\n    var a, b int\n    fmt.Scan(&a, &b)\n    fmt.Println(a + b)\n}', 'completed', 'ok', 15, 6),
(3, 1, 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello, World!")\n}', 'completed', 'ok', 12, 5),
(3, 2, 'package main\n\nimport "fmt"\n\nfunc main() {\n    var a, b int\n    fmt.Scan(&a, &b)\n    fmt.Println(a - b)\n}', 'completed', 'wrong_answer', 14, 6); 