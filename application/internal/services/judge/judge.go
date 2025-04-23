package judge

import (
	"context"
	"online-judge/internal/models"
	"online-judge/internal/services/code_runner"
	"time"
)

// JudgeResult represents the result of judging a submission
type JudgeResult struct {
	SubmissionID int
	Status       string // "accepted", "wrong_answer", "time_limit_exceeded", "memory_limit_exceeded", "runtime_error", "compilation_error"
	TimeUsed     time.Duration
	MemoryUsed   int64
	TestCases    []TestCaseResult
}

// TestCaseResult represents the result of a single test case
type TestCaseResult struct {
	TestCaseID int
	Status     string // "passed", "failed", "error"
	Input      string
	Expected   string
	Output     string
	Error      string
	TimeUsed   time.Duration
	MemoryUsed int64
}

// JudgeService is the interface for the judging service
type JudgeService interface {
	// JudgeSubmission judges a submission against its test cases
	JudgeSubmission(ctx context.Context, submission *models.Submission) (*JudgeResult, error)

	// Start starts the judge service
	Start(ctx context.Context) error

	// Stop stops the judge service
	Stop() error
}

// Config holds the configuration for the judge service
type Config struct {
	// MaxConcurrentJudges is the maximum number of concurrent judging processes
	MaxConcurrentJudges int

	// RetryInterval is the time to wait before retrying failed submissions
	RetryInterval time.Duration

	// CodeRunner is the code runner service to use
	CodeRunner code_runner.CodeRunner
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		MaxConcurrentJudges: 4,
		RetryInterval:       5 * time.Second,
	}
}

// NewJudgeService creates a new instance of the judge service
func NewJudgeService(config *Config) (JudgeService, error) {
	if config == nil {
		config = DefaultConfig()
	}

	return &judgeService{
		config: config,
		queue:  make(chan *models.Submission, 100),
		stop:   make(chan struct{}),
	}, nil
}

// judgeService is the implementation of the JudgeService interface
type judgeService struct {
	config *Config
	queue  chan *models.Submission
	stop   chan struct{}
}

// Start implements the JudgeService interface
func (js *judgeService) Start(ctx context.Context) error {
	// Start the worker pool
	for i := 0; i < js.config.MaxConcurrentJudges; i++ {
		go js.worker(ctx)
	}

	return nil
}

// Stop implements the JudgeService interface
func (js *judgeService) Stop() error {
	close(js.stop)
	return nil
}

// JudgeSubmission implements the JudgeService interface
func (js *judgeService) JudgeSubmission(ctx context.Context, submission *models.Submission) (*JudgeResult, error) {
	// Add submission to the queue
	select {
	case js.queue <- submission:
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// worker processes submissions from the queue
func (js *judgeService) worker(ctx context.Context) {
	for {
		select {
		case submission := <-js.queue:
			// Get the question and test cases
			question, err := js.getQuestion(ctx, submission.QuestionID)
			if err != nil {
				// TODO: Handle error and retry
				continue
			}

			// Judge the submission
			result, err := js.judge(ctx, submission, question)
			if err != nil {
				// TODO: Handle error and retry
				continue
			}

			// Save the result
			if err := js.saveResult(ctx, result); err != nil {
				// TODO: Handle error and retry
				continue
			}

		case <-js.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

// judge processes a single submission
func (js *judgeService) judge(ctx context.Context, submission *models.Submission, question *models.Question) (*JudgeResult, error) {
	result := &JudgeResult{
		SubmissionID: submission.ID,
		TestCases:    make([]TestCaseResult, 0, len(question.TestCases)),
	}

	// Process each test case
	for _, testCase := range question.TestCases {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(ctx, time.Duration(question.TimeLimitMs)*time.Millisecond)
		defer cancel()

		// Run the code
		execResult, err := js.config.CodeRunner.ExecuteCode(ctx, submission.Code, code_runner.Language(submission.Language), testCase.Input)
		if err != nil {
			// Handle different types of errors
			switch {
			case ctx.Err() == context.DeadlineExceeded:
				result.Status = "time_limit_exceeded"
			case err.Error() == "memory_limit_exceeded":
				result.Status = "memory_limit_exceeded"
			default:
				result.Status = "runtime_error"
			}
			return result, nil
		}

		// Compare output with expected output
		testCaseResult := TestCaseResult{
			TestCaseID: int(testCase.ID),
			Input:      testCase.Input,
			Expected:   testCase.ExpectedOutput,
			Output:     execResult.Output,
			Error:      execResult.Error,
			TimeUsed:   execResult.TimeUsed,
			MemoryUsed: execResult.MemoryUsed,
		}

		if execResult.Output == testCase.ExpectedOutput {
			testCaseResult.Status = "passed"
		} else {
			testCaseResult.Status = "failed"
			result.Status = "wrong_answer"
		}

		result.TestCases = append(result.TestCases, testCaseResult)
		result.TimeUsed += execResult.TimeUsed
		if execResult.MemoryUsed > result.MemoryUsed {
			result.MemoryUsed = execResult.MemoryUsed
		}
	}

	// If all test cases passed, mark as accepted
	if result.Status == "" {
		result.Status = "accepted"
	}

	return result, nil
}

// getQuestion retrieves a question and its test cases
func (js *judgeService) getQuestion(ctx context.Context, questionID int) (*models.Question, error) {
	// TODO: Implement database query
	return nil, nil
}

// saveResult saves the judging result
func (js *judgeService) saveResult(ctx context.Context, result *JudgeResult) error {
	// TODO: Implement database update
	return nil
}
