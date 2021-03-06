package reports_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/ta9mi1shi1/toggl-go/reports"
)

type summaryReport struct {
	Data []struct {
		ID    int `json:"id"`
		Title struct {
			Project string `json:"project"`
			Color   string `json:"color"`
			User    string `json:"user"`
		} `json:"title"`
		Time  int `json:"time"`
		Items []struct {
			Title struct {
				Project   string `json:"project"`
				User      string `json:"user"`
				TimeEntry string `json:"time_entry"`
			} `json:"title"`
			Time int `json:"time"`
		} `json:"items"`
	} `json:"data"`
}

func TestGetSummary(t *testing.T) {
	cases := []struct {
		name             string
		httpStatus       int
		testdataFilePath string
		in               context.Context
		out              error
	}{
		{
			name:             "200 OK",
			httpStatus:       http.StatusOK,
			testdataFilePath: "testdata/summary.json",
			in:               context.Background(),
			out:              nil,
		},
		{
			name:             "401 Unauthorized",
			httpStatus:       http.StatusUnauthorized,
			testdataFilePath: "testdata/401_unauthorized.json",
			in:               context.Background(),
			out: &reports.ReportsError{
				Err: struct {
					Message string `json:"message"`
					Tip     string `json:"tip"`
					Code    int    `json:"code"`
				}{
					Message: "api token missing",
					Tip:     "You can find your API Token in your profile at https://www.toggl.com",
					Code:    http.StatusUnauthorized,
				},
			},
		},
		{
			name:             "429 Too Many Requests",
			httpStatus:       http.StatusTooManyRequests,
			testdataFilePath: "testdata/429_too_many_requests.html",
			in:               context.Background(),
			out: &reports.ReportsError{
				Err: struct {
					Message string `json:"message"`
					Tip     string `json:"tip"`
					Code    int    `json:"code"`
				}{
					Message: "Too Many Requests",
					Tip:     "Add delay between requests",
					Code:    http.StatusTooManyRequests,
				},
			},
		},
		{
			name:             "Without context",
			httpStatus:       http.StatusOK,
			testdataFilePath: "testdata/summary.json",
			in:               nil,
			out:              reports.ErrContextNotFound,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mockServer, testdata := setupMockServer(t, c.httpStatus, c.testdataFilePath)
			defer mockServer.Close()

			actualSummaryReport := new(summaryReport)
			client := reports.NewClient(apiToken, baseURL(mockServer.URL))
			err := client.GetSummary(
				c.in,
				&reports.SummaryRequestParameters{
					StandardRequestParameters: &reports.StandardRequestParameters{
						UserAgent:   userAgent,
						WorkspaceID: workspaceID,
					},
				},
				actualSummaryReport,
			)

			if err == nil {
				expectedSummaryReport := new(summaryReport)
				if err := json.Unmarshal(testdata, expectedSummaryReport); err != nil {
					t.Error(err.Error())
				}
				if !reflect.DeepEqual(actualSummaryReport, expectedSummaryReport) {
					t.Errorf("\nwant: %+v\ngot : %+v\n", expectedSummaryReport, actualSummaryReport)
				}
			} else {
				if !reflect.DeepEqual(actualSummaryReport, &summaryReport{}) {
					t.Errorf("\nwant: %+v\ngot : %+v\n", &summaryReport{}, actualSummaryReport)
				}
			}

			var reportsError reports.Error
			if errors.As(err, &reportsError) {
				if !reflect.DeepEqual(reportsError, c.out) {
					t.Errorf("\nwant: %#+v\ngot : %#+v\n", c.out, reportsError)
				}
			} else {
				if !errors.Is(err, c.out) {
					t.Errorf("\nwant: %#+v\ngot : %#+v\n", c.out, err)
				}
			}
		})
	}
}

func TestGetSummaryEncodeRequestParameters(t *testing.T) {
	expectedQueryString := url.Values{
		"user_agent":   []string{userAgent},
		"workspace_id": []string{workspaceID},
		"grouping":     []string{"projects"},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualQueryString := r.URL.Query()
		if !reflect.DeepEqual(actualQueryString, expectedQueryString) {
			t.Error("Actual query string (" + actualQueryString.Encode() + ") is not as expected.")
		}
	}))

	client := reports.NewClient(apiToken, baseURL(mockServer.URL))
	_ = client.GetSummary(
		context.Background(),
		&reports.SummaryRequestParameters{
			StandardRequestParameters: &reports.StandardRequestParameters{
				UserAgent:   userAgent,
				WorkspaceID: workspaceID,
			},
			Grouping: "projects",
		},
		new(summaryReport),
	)
}
