package route53

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jaehong21/hibiscus/internal/aws/route53"
)

type fetchHostedZoneMsg struct {
	Rows []table.Row
	Err  error
}

func fetchHostedZone() tea.Cmd {
	return func() tea.Msg {
		hostedZones, err := route53.ListHostedZones()
		if err != nil {
			return fetchHostedZoneMsg{Err: err}
		}

		return fetchHostedZoneMsg{Rows: getHostedZoneRows(&hostedZones)}
	}
}

type filterHostedZoneMsg struct {
	Rows []table.Row
	Err  error
}

func filterHostedZone(query string) tea.Cmd {
	return func() tea.Msg {
		hostedZones, err := route53.ListHostedZones()
		if err != nil {
			return filterHostedZoneMsg{Err: err}
		}

		var filtered []types.HostedZone
		for _, hostedZone := range hostedZones {
			if strings.Contains(strings.ToLower(*hostedZone.Name), strings.ToLower(query)) {
				filtered = append(filtered, hostedZone)
			}
		}

		return filterHostedZoneMsg{Rows: getHostedZoneRows(&filtered)}
	}
}

func getHostedZoneRows(hostedZones *[]types.HostedZone) []table.Row {
	rows := []table.Row{}
	for _, hostedZone := range *hostedZones {
		rows = append(rows, table.Row{
			*hostedZone.Name,
			fmt.Sprintf("%d", *hostedZone.ResourceRecordSetCount),
			*hostedZone.Id,
		})
	}

	return rows
}

type fetchRecordsMsg struct {
	Rows []table.Row
	Err  error
}

func fetchRecords(hostedZoneID *string) tea.Cmd {
	return func() tea.Msg {
		records, err := route53.ListRecords(hostedZoneID)
		if err != nil {
			return fetchRecordsMsg{Err: err}
		}

		return fetchRecordsMsg{Rows: getRecordRows(&records)}
	}
}

type filterRecordsMsg struct {
	Rows []table.Row
	Err  error
}

func filterRecords(hostedZoneID *string, query string) tea.Cmd {
	return func() tea.Msg {
		records, err := route53.ListRecords(hostedZoneID)
		if err != nil {
			return filterRecordsMsg{Err: err}
		}

		var filtered []types.ResourceRecordSet
		for _, record := range records {
			var values string
			for _, rr := range record.ResourceRecords {
				values = *rr.Value + " "
			}

			if strings.Contains(strings.ToLower(*record.Name), strings.ToLower(query)) { // Name
				filtered = append(filtered, record)
			} else if strings.Contains(strings.ToLower(values), strings.ToLower(query)) { // Value
				filtered = append(filtered, record)
			}
		}

		return filterRecordsMsg{Rows: getRecordRows(&filtered)}
	}
}

func getRecordRows(records *[]types.ResourceRecordSet) []table.Row {
	var rows []table.Row

	for _, record := range *records {
		var ttl string
		var weight string

		if record.TTL == nil {
			ttl = ""
		} else {
			ttl = fmt.Sprintf("%d", *record.TTL)
		}

		if record.Weight == nil {
			weight = "-"
		} else {
			weight = fmt.Sprintf("%d", *record.Weight)
		}

		// Handle normal resource records
		if len(record.ResourceRecords) > 0 {
			for _, rr := range record.ResourceRecords {
				rows = append(rows, table.Row{
					*record.Name,
					string(record.Type),
					*rr.Value,
					ttl,
					weight,
				})
			}
		} else if record.AliasTarget != nil {
			// Handle alias records (CloudFront, ELB, etc.)
			aliasType := "Alias"

			// Identify the type of alias
			if route53.IsCloudFrontAlias(record.AliasTarget.HostedZoneId) {
				aliasType = "CloudFront"
			} else if route53.IsELBAlias(record.AliasTarget.HostedZoneId) {
				aliasType = "ELB"
			}

			rows = append(rows, table.Row{
				*record.Name,
				string(record.Type),
				fmt.Sprintf("%s (%s) -> %s", aliasType, *record.AliasTarget.HostedZoneId, *record.AliasTarget.DNSName),
				"", // No TTL for alias records
				"", // No weight for alias records
			})
		}

		// NOTE: add blank row
		// rows = append(rows, table.Row{})
	}

	return rows
}
