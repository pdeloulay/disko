#!/bin/bash

# Simple script to test board API endpoints manually
# This assumes the server is running on localhost:8080

BASE_URL="http://localhost:8080/api"
AUTH_TOKEN="test_token_123"  # This would be a real Clerk JWT token in production

echo "Testing Board Management API Endpoints"
echo "======================================"

# Test 1: Create a board
echo "1. Creating a new board..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/boards" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d '{
    "name": "Test Board",
    "description": "This is a test board created via API",
    "visibleColumns": ["parking", "now", "next", "later"],
    "visibleFields": ["oneLiner", "description", "valueStatement"]
  }')

echo "Create Response: $CREATE_RESPONSE"
echo ""

# Extract board ID from response (requires jq)
if command -v jq &> /dev/null; then
  BOARD_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')
  echo "Created Board ID: $BOARD_ID"
else
  echo "jq not found. Please install jq to extract board ID automatically."
  echo "Manually extract the board ID from the response above for the next tests."
  read -p "Enter the board ID: " BOARD_ID
fi
echo ""

# Test 2: Get all boards
echo "2. Getting all boards..."
GET_BOARDS_RESPONSE=$(curl -s -X GET "$BASE_URL/boards" \
  -H "Authorization: Bearer $AUTH_TOKEN")

echo "Get Boards Response: $GET_BOARDS_RESPONSE"
echo ""

# Test 3: Update the board (only if we have a board ID)
if [ "$BOARD_ID" != "null" ] && [ -n "$BOARD_ID" ]; then
  echo "3. Updating the board..."
  UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL/boards/$BOARD_ID" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    -d '{
      "name": "Updated Test Board",
      "description": "This board has been updated",
      "visibleColumns": ["parking", "now", "release"]
    }')

  echo "Update Response: $UPDATE_RESPONSE"
  echo ""

  # Test 4: Delete the board
  echo "4. Deleting the board..."
  DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/boards/$BOARD_ID" \
    -H "Authorization: Bearer $AUTH_TOKEN")

  echo "Delete Response: $DELETE_RESPONSE"
  echo ""
else
  echo "Skipping update and delete tests - no valid board ID"
fi

echo "Testing complete!"
echo ""
echo "Note: These tests will fail with authentication errors unless:"
echo "1. The server is running with proper Clerk configuration"
echo "2. A valid Clerk JWT token is provided"
echo "3. MongoDB is connected and accessible"