#!/bin/bash

# Function to prompt for user input if environment variables are not set
prompt_for_input() {
    local var_name=$1
    local var_value=${!var_name}
    if [ -z "$var_value" ]; then
        read -p "Enter your $var_name: " var_value
    fi
    echo $var_value
}

# Function to fetch the list of snapshots
fetch_snapshots() {
    curl --user "${PUBLIC_API_KEY}:${PRIVATE_API_KEY}" --digest \
        --header "Accept: application/vnd.atlas.2024-11-13+json" \
        -X GET "https://cloud.mongodb.com/api/atlas/v2/groups/${PROJECT_ID}/clusters/${CLUSTER_NAME}/backup/snapshots"
}

# Function to extract the oldest snapshot details
extract_oldest_snapshot() {
    echo $1 | jq -r '.results | sort_by(.createdAt) | .[0]'
}

# Function to delete the snapshot
delete_snapshot() {
    curl --user "${PUBLIC_API_KEY}:${PRIVATE_API_KEY}" --digest --include \
        --header "Accept: application/vnd.atlas.2024-11-13+json" \
        --request DELETE "https://cloud.mongodb.com/api/atlas/v2/groups/${PROJECT_ID}/clusters/${CLUSTER_NAME}/backup/snapshots/${1}"
}

# Function to read all snapshots
read_all_snapshots() {
    curl --user "${PUBLIC_API_KEY}:${PRIVATE_API_KEY}" --digest \
        --header "Accept: application/vnd.atlas.2024-11-13+json" \
        -X GET "https://cloud.mongodb.com/api/atlas/v2/groups/${PROJECT_ID}/clusters/${CLUSTER_NAME}/backup/snapshots?pretty=true"
}

# Function to take a snapshot
take_snapshot() {
    curl --user "${PUBLIC_API_KEY}:${PRIVATE_API_KEY}" --digest --include \
        --header "Accept: application/vnd.atlas.2024-11-13+json" \
        --header "Content-Type: application/json" \
        --request POST "https://cloud.mongodb.com/api/atlas/v2/groups/${PROJECT_ID}/clusters/${CLUSTER_NAME}/backup/snapshots" \
        --data '{
         "description": "Snapshot created via API",
         "retentionInDays": 1
       }'
}

# Prompt for necessary variables if not set
PUBLIC_API_KEY=$(prompt_for_input "PUBLIC_API_KEY")
PRIVATE_API_KEY=$(prompt_for_input "PRIVATE_API_KEY")
PROJECT_ID=$(prompt_for_input "PROJECT_ID")
CLUSTER_NAME=$(prompt_for_input "CLUSTER_NAME")

# Display menu
echo "Select an action:"
echo "1. Take a snapshot"
echo "2. Read all snapshots"
echo "3. Delete the oldest snapshot"
read -p "Enter your choice (1/2/3): " choice

case $choice in
1)
    take_snapshot
    echo "Snapshot taken successfully."
    ;;
2)
    read_all_snapshots
    ;;
3)
    # Fetch the list of snapshots
    response=$(fetch_snapshots)

    # Extract the oldest snapshot details
    snapshot=$(extract_oldest_snapshot "$response")
    snapshot_id=$(echo $snapshot | jq -r '.id')
    snapshot_status=$(echo $snapshot | jq -r '.status')
    snapshot_created_at=$(echo $snapshot | jq -r '.createdAt')

    # Display the oldest snapshot details
    echo "--------------------------------"
    echo "Oldest snapshot ID: $snapshot_id"
    echo "Status: $snapshot_status"
    echo "Created At: $snapshot_created_at"
    echo "--------------------------------"

    # Check if snapshot ID is found
    if [ -z "$snapshot_id" ]; then
        echo "No snapshots found."
        exit 1
    fi

    # Prompt for confirmation to delete the oldest snapshot
    read -p "Do you want to delete ${snapshot_id}? (y/n): " confirm
    if [ "$confirm" != "y" ]; then
        echo "Deletion cancelled."
        exit 0
    fi

    # Delete the oldest snapshot
    delete_response=$(delete_snapshot "$snapshot_id")

    # Display the delete response
    echo "Delete response: $delete_response"
    ;;
*)
    echo "Invalid choice."
    ;;
esac
