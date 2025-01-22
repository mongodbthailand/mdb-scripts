package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"go.mongodb.org/atlas-sdk/v20241113004/admin"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	apiKey := getEnvOrPrompt("PUBLIC_API_KEY", reader)
	apiSecret := getEnvOrPrompt("PRIVATE_API_KEY", reader)
	groupId := getEnvOrPrompt("PROJECT_ID", reader)
	clusterName := getEnvOrPrompt("CLUSTER_NAME", reader)

	sdk, err := admin.NewClient(admin.UseDigestAuth(apiKey, apiSecret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing SDK: %v\n", err)
		return
	}

	for {
		fmt.Println("Select an action:")
		fmt.Println("0. Exit")
		fmt.Println("1. Take a snapshot")
		fmt.Println("2. Read all snapshots")
		fmt.Println("3. Delete the oldest snapshot")
		fmt.Print("Enter your choice (0/1/2/3): ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "0":
			fmt.Println("Exiting program.")
			return
		case "1":
			takeSnapshot(sdk, groupId, clusterName)
		case "2":
			readAllSnapshots(sdk, groupId, clusterName)
		case "3":
			deleteOldestShardedSnapshot(sdk, groupId, clusterName, reader)
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

func getEnvOrPrompt(envVar string, reader *bufio.Reader) string {
	value := os.Getenv(envVar)
	if value == "" {
		fmt.Printf("Enter your %s: ", envVar)
		input, _ := reader.ReadString('\n')
		value = strings.TrimSpace(input)
	}
	return value
}

func takeSnapshot(sdk *admin.APIClient, groupId, clusterName string) {
	diskBackupOnDemandSnapshotRequest := admin.NewDiskBackupOnDemandSnapshotRequest() // DiskBackupOnDemandSnapshotRequest |
	diskBackupOnDemandSnapshotRequest.SetDescription("Take Snapshot via API")
	diskBackupOnDemandSnapshotRequest.SetRetentionInDays(1)
	resp, r, err := sdk.CloudBackupsApi.TakeSnapshot(context.Background(), groupId, clusterName, diskBackupOnDemandSnapshotRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CloudBackupsApi.TakeSnapshot`: %v (%v)\n", err, r)
		apiError, ok := admin.AsError(err)
		if ok {
			fmt.Fprintf(os.Stderr, "API error obj: %v\n", apiError)
		}
		return
	}
	// response from `TakeSnapshot`: DiskBackupSnapshot
	fmt.Fprintf(os.Stdout, "Response from `CloudBackupsApi.TakeSnapshot`: %v (%v)\n", resp, r)
	fmt.Println("Snapshot taken successfully.")
}

func readAllSnapshots(sdk *admin.APIClient, groupId, clusterName string) {
	resp, r, err := sdk.CloudBackupsApi.ListShardedClusterBackups(context.Background(), groupId, clusterName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CloudBackupsApi.ListShardedClusterBackups`: %v (%v)\n", err, r)
		apiError, ok := admin.AsError(err)
		if ok {
			fmt.Fprintf(os.Stderr, "API error obj: %v\n", apiError)
		}
		return
	}

	for _, snapshot := range *resp.Results {
		fmt.Printf("Snapshot ID: %s, Status: %s, Created At: %s\n", *snapshot.Id, *snapshot.Status, snapshot.CreatedAt.String())
	}
}

func deleteOldestShardedSnapshot(sdk *admin.APIClient, groupId, clusterName string, reader *bufio.Reader) {
	resp, r, err := sdk.CloudBackupsApi.ListShardedClusterBackups(context.Background(), groupId, clusterName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CloudBackupsApi.ListShardedClusterBackups`: %v (%v)\n", err, r)
		apiError, ok := admin.AsError(err)
		if ok {
			fmt.Fprintf(os.Stderr, "API error obj: %v\n", apiError)
		}
		return
	}

	if resp.Results == nil || len(*resp.Results) == 0 {
		fmt.Println("No snapshots found.")
		return
	}

	// Sort the snapshots by createdAt
	sort.Slice(*resp.Results, func(i, j int) bool {
		createdAtI := (*resp.Results)[i].CreatedAt
		createdAtJ := (*resp.Results)[j].CreatedAt
		return createdAtI.Before(*createdAtJ)
	})

	// Select the oldest snapshot
	oldestSnapshot := (*resp.Results)[0]

	// Display the oldest snapshot details
	if oldestSnapshot.Id == nil || oldestSnapshot.Status == nil || oldestSnapshot.CreatedAt == nil {
		fmt.Println("Invalid snapshot data.")
		return
	}

	fmt.Printf("Oldest Snapshot ID: %s, Status: %s, Created At: %s\n", *oldestSnapshot.Id, *oldestSnapshot.Status, oldestSnapshot.CreatedAt.String())

	// Prompt for confirmation to delete the oldest snapshot
	fmt.Print("Do you want to delete this snapshot? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "y" {
		fmt.Println("Deletion cancelled.")
		return
	}

	// Select the oldest snapshot
	if resp.Results != nil && len(*resp.Results) > 0 {
		resp, r, err := sdk.CloudBackupsApi.DeleteShardedClusterBackup(context.Background(), groupId, clusterName, *oldestSnapshot.Id).Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error when calling `CloudBackupsApi.DeleteShardedClusterBackup`: %v (%v)\n", err, r)
			apiError, ok := admin.AsError(err)
			if ok {
				fmt.Fprintf(os.Stderr, "API error obj: %v\n", apiError)
			}
			return
		}
		// response from `DeleteShardedClusterBackup`: any
		fmt.Fprintf(os.Stdout, "Response from `CloudBackupsApi.DeleteShardedClusterBackup`: %v (%v)\n", resp, r)
		fmt.Println("Oldest snapshot deleted successfully.")
	} else {
		fmt.Println("No snapshots found.")
	}
}
