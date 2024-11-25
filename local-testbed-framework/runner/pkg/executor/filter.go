package executor

import (
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/client"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/testcase"
	"golang.org/x/exp/slices"
)

func FilterTestCases(testCases []testcase.TestCase, testCaseFilter []string) []testcase.TestCase {
	if testCaseFilter == nil || len(testCaseFilter) == 0 || testCaseFilter[0] == "*" {
		return testCases
	}

	filteredTestCases := make([]testcase.TestCase, 0, len(testCaseFilter)+1)

	for _, testCase := range testCases {
		if slices.Contains(testCaseFilter, testCase.Name()) || testCase.Name() == testcase.MetaTestCaseName {
			filteredTestCases = append(filteredTestCases, testCase)
		}
	}

	return filteredTestCases
}

func FilterClients(clients []client.Client, clientFilter []string) []client.Client {
	if clientFilter == nil || len(clientFilter) == 0 || clientFilter[0] == "*" {
		return clients
	}

	filteredClients := make([]client.Client, 0, len(clientFilter))

	for _, currentClient := range clients {
		if slices.Contains(clientFilter, currentClient.DisplayName()) || slices.Contains(clientFilter, currentClient.DirectoryName()) {
			filteredClients = append(filteredClients, currentClient)
		}
	}

	return filteredClients
}
