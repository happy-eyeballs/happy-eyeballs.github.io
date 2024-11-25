package result

import "gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/resultdatabase"

type Result interface {
	WithRunId(runId int) Result
	WithTestCaseName(testCaseName string) Result
	WithClientName(clientName string) Result
	WithMeasurement(measurement string) Result
	WithIterationVariable(iterationVariable int) Result
	Record(value *string) error
}

type result struct {
	resultDatabase resultdatabase.ResultDatabase

	runId             int
	testCaseName      string
	clientName        string
	measurement       string
	iterationVariable int
	value             string
}

func NewResult(resultDatabase resultdatabase.ResultDatabase) Result {
	return &result{
		resultDatabase: resultDatabase,
	}
}

func (r *result) WithRunId(runId int) Result {
	resultCopy := *r
	resultCopy.runId = runId
	return &resultCopy
}

func (r *result) WithTestCaseName(testCaseName string) Result {
	resultCopy := *r
	resultCopy.testCaseName = testCaseName
	return &resultCopy
}

func (r *result) WithClientName(clientName string) Result {
	resultCopy := *r
	resultCopy.clientName = clientName
	return &resultCopy
}

func (r *result) WithMeasurement(measurement string) Result {
	resultCopy := *r
	resultCopy.measurement = measurement
	return &resultCopy
}

func (r *result) WithIterationVariable(iterationVariable int) Result {
	resultCopy := *r
	resultCopy.iterationVariable = iterationVariable
	return &resultCopy
}

func (r *result) Record(value *string) error {
	return r.resultDatabase.Insert(r.runId, r.testCaseName, r.clientName, r.measurement, r.iterationVariable, value)
}
