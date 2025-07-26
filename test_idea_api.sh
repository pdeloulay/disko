#!/bin/bash

# Test script for Idea Management API endpoints
# This script tests the idea management functionality

BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Check if server is running
print_info "Checking if server is running..."
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo -e "${RED}Server is not running. Please start the server first.${NC}"
    exit 1
fi
print_status 0 "Server is running"

# Note: These tests require authentication
print_info "Note: These tests require a valid authentication token"
print_info "Please ensure you have a valid Clerk token for testing"

# Test data
BOARD_NAME="Test Board for Ideas"
BOARD_DESC="Test board for idea management testing"

echo ""
print_info "=== Testing Idea Management API ==="

# Test 1: Create a test board first (needed for idea tests)
print_info "1. Creating test board..."
BOARD_RESPONSE=$(curl -s -X POST "$API_URL/boards" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TEST_TOKEN_HERE" \
  -d "{
    \"name\": \"$BOARD_NAME\",
    \"description\": \"$BOARD_DESC\"
  }")

if echo "$BOARD_RESPONSE" | grep -q '"id"'; then
    BOARD_ID=$(echo "$BOARD_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    print_status 0 "Board created successfully (ID: $BOARD_ID)"
else
    print_status 1 "Failed to create board"
    echo "Response: $BOARD_RESPONSE"
    exit 1
fi

# Test 2: Create an idea
print_info "2. Creating test idea..."
IDEA_RESPONSE=$(curl -s -X POST "$API_URL/boards/$BOARD_ID/ideas" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TEST_TOKEN_HERE" \
  -d '{
    "oneLiner": "Test Idea for API",
    "description": "This is a comprehensive test idea to verify the API functionality",
    "valueStatement": "This idea provides significant value for testing purposes",
    "riceScore": {
      "reach": 80,
      "impact": 70,
      "confidence": 4,
      "effort": 60
    },
    "column": "parking",
    "position": 1
  }')

if echo "$IDEA_RESPONSE" | grep -q '"id"'; then
    IDEA_ID=$(echo "$IDEA_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    print_status 0 "Idea created successfully (ID: $IDEA_ID)"
    
    # Verify idea ID starts with "I"
    if [[ $IDEA_ID == I* ]]; then
        print_status 0 "Idea ID format is correct (starts with 'I')"
    else
        print_status 1 "Idea ID format is incorrect (should start with 'I')"
    fi
else
    print_status 1 "Failed to create idea"
    echo "Response: $IDEA_RESPONSE"
    exit 1
fi

# Test 3: Get board ideas
print_info "3. Fetching board ideas..."
IDEAS_RESPONSE=$(curl -s -X GET "$API_URL/boards/$BOARD_ID/ideas" \
  -H "Authorization: Bearer YOUR_TEST_TOKEN_HERE")

if echo "$IDEAS_RESPONSE" | grep -q '"ideas"'; then
    IDEAS_COUNT=$(echo "$IDEAS_RESPONSE" | grep -o '"count":[0-9]*' | cut -d':' -f2)
    print_status 0 "Board ideas fetched successfully (Count: $IDEAS_COUNT)"
else
    print_status 1 "Failed to fetch board ideas"
    echo "Response: $IDEAS_RESPONSE"
fi

# Test 4: Update idea
print_info "4. Updating idea..."
UPDATE_RESPONSE=$(curl -s -X PUT "$API_URL/ideas/$IDEA_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TEST_TOKEN_HERE" \
  -d '{
    "oneLiner": "Updated Test Idea",
    "inProgress": true,
    "column": "now"
  }')

if echo "$UPDATE_RESPONSE" | grep -q '"oneLiner":"Updated Test Idea"'; then
    print_status 0 "Idea updated successfully"
else
    print_status 1 "Failed to update idea"
    echo "Response: $UPDATE_RESPONSE"
fi

# Test 5: Update idea position (drag-drop simulation)
print_info "5. Updating idea position..."
POSITION_RESPONSE=$(curl -s -X PUT "$API_URL/ideas/$IDEA_ID/position" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TEST_TOKEN_HERE" \
  -d '{
    "column": "next",
    "position": 2
  }')

if echo "$POSITION_RESPONSE" | grep -q '"column":"next"'; then
    print_status 0 "Idea position updated successfully"
else
    print_status 1 "Failed to update idea position"
    echo "Response: $POSITION_RESPONSE"
fi

# Test 6: Test RICE score validation
print_info "6. Testing RICE score validation..."
INVALID_RICE_RESPONSE=$(curl -s -X POST "$API_URL/boards/$BOARD_ID/ideas" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TEST_TOKEN_HERE" \
  -d '{
    "oneLiner": "Invalid RICE Test",
    "description": "Testing invalid RICE score",
    "valueStatement": "Testing validation",
    "riceScore": {
      "reach": 101,
      "impact": 70,
      "confidence": 3,
      "effort": 60
    }
  }')

if echo "$INVALID_RICE_RESPONSE" | grep -q '"code":"INVALID_RICE_SCORE"'; then
    print_status 0 "RICE score validation working correctly"
else
    print_status 1 "RICE score validation not working"
    echo "Response: $INVALID_RICE_RESPONSE"
fi

# Test 7: Test status update to "done" (should move to release)
print_info "7. Testing status update to 'done'..."
DONE_RESPONSE=$(curl -s -X PUT "$API_URL/ideas/$IDEA_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TEST_TOKEN_HERE" \
  -d '{
    "status": "done"
  }')

if echo "$DONE_RESPONSE" | grep -q '"column":"release"'; then
    print_status 0 "Status update to 'done' correctly moved idea to release column"
else
    print_status 1 "Status update to 'done' did not move idea to release column"
    echo "Response: $DONE_RESPONSE"
fi

# Test 8: Delete idea
print_info "8. Deleting test idea..."
DELETE_RESPONSE=$(curl -s -X DELETE "$API_URL/ideas/$IDEA_ID" \
  -H "Authorization: Bearer YOUR_TEST_TOKEN_HERE")

if echo "$DELETE_RESPONSE" | grep -q '"message":"Idea deleted successfully"'; then
    print_status 0 "Idea deleted successfully"
else
    print_status 1 "Failed to delete idea"
    echo "Response: $DELETE_RESPONSE"
fi

# Cleanup: Delete test board
print_info "9. Cleaning up test board..."
CLEANUP_RESPONSE=$(curl -s -X DELETE "$API_URL/boards/$BOARD_ID" \
  -H "Authorization: Bearer YOUR_TEST_TOKEN_HERE")

if echo "$CLEANUP_RESPONSE" | grep -q '"message"'; then
    print_status 0 "Test board cleaned up successfully"
else
    print_status 1 "Failed to clean up test board"
    echo "Response: $CLEANUP_RESPONSE"
fi

echo ""
print_info "=== Idea Management API Testing Complete ==="
print_info "Note: Some tests may fail if authentication is not properly configured"
print_info "Replace 'YOUR_TEST_TOKEN_HERE' with a valid Clerk JWT token for full testing"