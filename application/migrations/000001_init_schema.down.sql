-- Drop triggers first
DROP TRIGGER IF EXISTS update_submissions_updated_at ON submissions;
DROP TRIGGER IF EXISTS update_test_cases_updated_at ON test_cases;
DROP TRIGGER IF EXISTS update_questions_updated_at ON questions;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS test_cases;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS users;

-- Drop enum types
DROP TYPE IF EXISTS submission_result;
DROP TYPE IF EXISTS submission_status;
DROP TYPE IF EXISTS question_status;
DROP TYPE IF EXISTS user_role; 