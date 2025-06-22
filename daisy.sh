#!/bin/bash

API_URL="http://localhost:8080/api/entries"

# Create a new entry
create_entry() {
    read -p "Rating (1-10): " rating
    read -p "Note (optional): " note
    read -p "Tags (comma-separated): " tags

    curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -d "{\"rating\": $rating, \"note\": \"$note\", \"tags\": \"$tags\"}" | jq
}

# Get all entries
list_entries() {
    curl -s "$API_URL" | jq
}

# Get a specific entry by ID
get_entry() {
    read -p "Entry ID: " id
    curl -s "$API_URL/$id" | jq
}

# Update an entry
update_entry() {
    read -p "Entry ID to update: " id
    read -p "New Rating (1-10): " rating
    read -p "New Note: " note
    read -p "New Tags (comma-separated): " tags

    curl -s -X PUT "$API_URL/$id" \
        -H "Content-Type: application/json" \
        -d "{\"rating\": $rating, \"note\": \"$note\", \"tags\": \"$tags\"}" | jq
}

# Delete an entry
delete_entry() {
    read -p "Entry ID to delete: " id
    curl -s -X DELETE "$API_URL/$id"
    echo "Deleted entry ID $id."
}

# Show help
show_help() {
    echo "Usage: $0 [command]"
    echo "Commands:"
    echo "  create     Create a new wellness entry"
    echo "  list       List all entries"
    echo "  get        Get a specific entry by ID"
    echo "  update     Update an existing entry"
    echo "  delete     Delete an entry by ID"
    echo "  help       Show this help message"
}

# Main dispatcher
case "$1" in
    create) create_entry ;;
    list) list_entries ;;
    get) get_entry ;;
    update) update_entry ;;
    delete) delete_entry ;;
    help | *) show_help ;;
esac
