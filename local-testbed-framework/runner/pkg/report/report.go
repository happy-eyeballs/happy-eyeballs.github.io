package report

import (
	"fmt"
	"github.com/rs/zerolog"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/resultdatabase"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/pkg/utils"
	"path"
	"strconv"
	"strings"
)

type Report interface {
	GenerateReport(logger zerolog.Logger, artifactsDirectoryPath string, resultDatabase resultdatabase.ResultDatabase, runId int)
}

type report struct {
}

func NewReport() Report {
	return &report{}
}

func (r *report) GenerateReport(logger zerolog.Logger, artifactsDirectoryPath string, resultDatabase resultdatabase.ResultDatabase, runId int) {
	logger.Info().Msg("Starting report generation")

	report := fmt.Sprintf("# Report (%d)\n\n", runId)

	testCases, err := resultDatabase.GetTestCases(runId)
	if err != nil {
		logger.Err(err).Msg("Error getting test cases from database")
		return
	}

	clients, err := resultDatabase.GetClients(runId)
	if err != nil {
		logger.Err(err).Msg("Error getting clients from database")
		return
	}

	for _, testCase := range testCases {
		report += fmt.Sprintf("## %s\n", testCase)

		measurements, err := resultDatabase.GetMeasurementsOfTestCase(runId, testCase)
		if err != nil {
			logger.Err(err).Msg("Error getting measurements of test case from database")
			continue
		}

		for _, measurement := range measurements {
			report += fmt.Sprintf("### %s\n", measurement)

			iterationVariables, err := resultDatabase.GetIterationVariablesOfMeasurement(runId, testCase, measurement)
			if err != nil {
				logger.Err(err).Msg("Error getting iteration variables of measurement from database")
				continue
			}

			tableHeader := []string{"Client"}
			for _, iterationVariable := range iterationVariables {
				tableHeader = append(tableHeader, strconv.Itoa(iterationVariable))
			}

			report += fmt.Sprintf("| %s |\n", strings.Join(tableHeader, " | "))
			report += fmt.Sprintf("|%s\n", strings.Repeat("---|", len(tableHeader)))

			for _, client := range clients {
				values := make([]string, 0, len(iterationVariables))
				for _, iterationVariable := range iterationVariables {
					value, err := resultDatabase.GetValue(runId, testCase, measurement, client, iterationVariable)
					if err != nil {
						logger.Err(err).Msg("Error getting value from database")
					}

					if value == nil {
						values = append(values, "")
					} else {
						values = append(values, *value)
					}
				}

				report += fmt.Sprintf("| %s | %s |\n", client, strings.Join(values, " | "))
			}
		}
	}

	writeReportFile(logger, path.Join(artifactsDirectoryPath, "report.md"), report)

	logger.Info().Msg("Report generation finished")
}

func writeReportFile(logger zerolog.Logger, filePath string, data string) {
	reportFile, err := utils.CreateFile(filePath)
	if err != nil {
		logger.Err(err).Msg("Error creating report file")
		return
	}

	_, err = reportFile.Write([]byte(data))
	if err != nil {
		logger.Err(err).Msg("Error writing report file")
		return
	}

	err = reportFile.Close()
	if err != nil {
		logger.Err(err).Msg("Error closing report file")
		return
	}
}
