#!/bin/bash

# Test script for idea position update API
# This script tests the drag-and-drop position update functionality

echo "Testing Idea Position Update API..."

# First, let's create a test board and idea
echo "1. Creating test board..."
BOARD_RESPONSE=$(curl -s -X POST http://localhost:8080/api/boards \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "name": "Test Drag Drop Board",
    "description": "Testing drag and drop functionality"
  }')

echo "Board response: $BOARD_RESPONSE"

# Extract board ID (assuming JSON response)
BOARD_ID=$(echo $BOARD_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Board ID: $BOARD_ID"

if [ -z "$BOARD_ID" ]; then
  echo "Failed to create board or extract board ID"
  exit 1
fi

echo "2. Creating test idea..."
IDEA_RESPONSE=$(curl -s -X POST http://localhost:8080/api/boards/$BOARD_ID/ideas \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "oneLiner": "Test drag drop idea",
    "description": "This idea will be used to test drag and drop functionality",
    "valueStatement": "Provides testing value for drag and drop",
    "riceScore": {
      "reach": 80,
      "impact": 70,
      "confidence": 4,
      "effort": 60
    },
    "column": "parking"
  }')

echo "Idea response: $IDEA_RESPONSE"

# Extract idea ID
IDEA_ID=$(echo $IDEA_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Idea ID: $IDEA_ID"

if [ -z "$IDEA_ID" ]; then
  echo "Failed to create idea or extract idea ID"
  exit 1
fi

echo "3. Testing position update (moving from parking to now)..."
POSITION_RESPONSE=$(curl -s -X PUT http://localhost:8080/api/ideas/$IDEA_ID/position \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "column": "now",
    "position": 1
  }')

echo "Position update response: $POSITION_RESPONSE"

echo "4. Verifying the update by fetching the idea..."
VERIFY_RESPONSE=$(curl -s -X GET http://localhost:8080/api/boards/$BOARD_ID/ideas \
  -H "Authorization: Bearer test-token")

echo "Verification response: $VERIFY_RESPONSE"

echo "5. Testing another position update (moving to next)..."
POSITION_RESPONSE2=$(curl -s -X PUT http://localhost:8080/api/ideas/$IDEA_ID/position \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "column": "next",
    "position": 1
  }')

echo "Second position update response: $POSITION_RESPONSE2"

echo "6. Cleanup - deleting test idea and board..."
curl -s -X DELETE http://localhost:8080/api/ideas/$IDEA_ID \
  -H "Authorization: Bearer test-token"

curl -s -X DELETE http://localhost:8080/api/boards/$BOARD_ID \
  -H "Authorization: Bearer test-token"

echo "Test completed!"