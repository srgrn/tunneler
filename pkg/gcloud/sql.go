package gcloud

import (
	"context"

	"google.golang.org/api/option"
	"google.golang.org/api/sqladmin/v1"
)

// GetSQLInstanceIP returns the IP address of a Cloud SQL instance.
func GetSQLInstanceIP(ctx context.Context, project, instance string) (string, error) {
	sqladminService, err := sqladmin.NewService(ctx, option.WithScopes(sqladmin.SqlserviceAdminScope))
	if err != nil {
		return "", err
	}

	resp, err := sqladminService.Instances.Get(project, instance).Do()
	if err != nil {
		return "", err
	}

	for _, ipAddress := range resp.IpAddresses {
		if ipAddress.Type == "PRIMARY" {
			return ipAddress.IpAddress, nil
		}
	}

	return "", nil
}