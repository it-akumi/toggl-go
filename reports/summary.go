package reports

import (
	"context"
)

const (
	summaryEndpoint string = "/reports/api/v2/summary"
)

// SummaryRequestParameters represents request parameters used in the summary report.
type SummaryRequestParameters struct {
	*StandardRequestParameters
	Grouping            string
	Subgrouping         string
	SubgroupingIDs      bool
	GroupedTimeEntryIDs bool
}

func (params *SummaryRequestParameters) urlEncode() string {
	values := params.StandardRequestParameters.values()

	if params.Grouping != "" {
		values.Add("grouping", params.Grouping)
	}
	if params.Subgrouping != "" {
		values.Add("subgrouping", params.Subgrouping)
	}
	if params.GroupedTimeEntryIDs {
		values.Add("grouped_time_entry_ids", "true")
	}
	if params.SubgroupingIDs {
		values.Add("subgrouping_ids", "true")
	}

	return values.Encode()
}

// GetSummary gets a summary report.
func (c *Client) GetSummary(ctx context.Context, params *SummaryRequestParameters, summaryReport interface{}) error {
	err := c.get(ctx, c.buildURL(summaryEndpoint, params), summaryReport)
	if err != nil {
		return err
	}
	return nil
}
