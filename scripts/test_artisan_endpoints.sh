#!/bin/bash

# Artisan Endpoints Test Script
# ==============================

TOKEN="nu8lKkL_VTZDHVj3NnwEMKWOvwPIU4Um_fuSJnjAZWxDezIKjwn9pGffrwRyaouVLKT9QHY"
BASE_URL="http://localhost:3000/api/v1"
HEADER_AUTH="Authorization: Bearer $TOKEN"
HEADER_JSON="Content-Type: application/json"

# Test data IDs from previous tests
TENANT_ID="bd787230-0bce-4bd6-93cd-2340da8ac445"
EXISTING_ARTISAN_ID="885e7825-6137-4e0c-8334-02238218cf40"
EXISTING_USER_ID="a1234567-e89b-12d3-a456-426614174001"

echo "==============================="
echo "Artisan Endpoints Test"
echo "==============================="
echo ""
echo "Test Data:"
echo "- Tenant: $TENANT_ID"
echo "- Existing Artisan: $EXISTING_ARTISAN_ID"
echo "- Existing User: $EXISTING_USER_ID"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Variables
ARTISAN_ID=""

# Test function
test_endpoint() {
    local method=$1
    local endpoint=$2
    local description=$3
    local data=$4
    local use_auth=${5:-true}

    echo -e "${YELLOW}Testing: $description${NC}"
    echo "Method: $method"
    echo "Endpoint: $BASE_URL$endpoint"
    echo ""

    if [ "$use_auth" = "false" ]; then
        if [ -z "$data" ]; then
            response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X $method \
                -H "$HEADER_JSON" "$BASE_URL$endpoint")
        else
            response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X $method \
                -H "$HEADER_JSON" -d "$data" "$BASE_URL$endpoint")
        fi
    else
        if [ -z "$data" ]; then
            response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X $method \
                -H "$HEADER_AUTH" -H "$HEADER_JSON" "$BASE_URL$endpoint")
        else
            response=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X $method \
                -H "$HEADER_AUTH" -H "$HEADER_JSON" -d "$data" "$BASE_URL$endpoint")
        fi
    fi

    http_status=$(echo "$response" | grep "HTTP_STATUS" | cut -d: -f2)
    body=$(echo "$response" | sed '/HTTP_STATUS/d')

    if [ "$http_status" -ge 200 ] && [ "$http_status" -lt 300 ]; then
        echo -e "${GREEN}✓ Success (HTTP $http_status)${NC}"
    else
        echo -e "${RED}✗ Failed (HTTP $http_status)${NC}"
    fi

    echo "Response:"
    echo "$body" | jq '.' 2>/dev/null || echo "$body"

    # Extract artisan ID from create response
    if [ "$method" = "POST" ] && [[ "$endpoint" == "/artisans" ]] && [ "$http_status" -ge 200 ] && [ "$http_status" -lt 300 ]; then
        ARTISAN_ID=$(echo "$body" | jq -r '.data.id // .id // empty')
        if [ -n "$ARTISAN_ID" ]; then
            echo -e "${BLUE}📝 Saved Artisan ID: $ARTISAN_ID${NC}"
        fi
    fi

    echo ""
    echo "-----------------------------------"
    echo ""
}

echo -e "${BLUE}=== CORE ARTISAN OPERATIONS ===${NC}"
echo ""

# 1. Get Existing Artisan by ID
test_endpoint "GET" "/artisans/$EXISTING_ARTISAN_ID" "Get Existing Artisan by ID" "" "true"

# 2. List All Artisans
test_endpoint "GET" "/artisans?page=1&page_size=20" "List All Artisans" "" "true"

# 3. Get Artisan by User ID
test_endpoint "GET" "/artisans/user/$EXISTING_USER_ID" "Get Artisan by User ID" "" "true"

# 4. Update Artisan
test_endpoint "PUT" "/artisans/$EXISTING_ARTISAN_ID" "Update Artisan Profile" "{
  \"bio\": \"Updated bio: Expert craftsman with 10+ years of experience in custom woodworking and furniture design.\",
  \"years_experience\": 12,
  \"commission_rate\": 18.5
}" "true"

echo -e "${BLUE}=== ARTISAN DISCOVERY ===${NC}"
echo ""

# 5. Get Available Artisans
test_endpoint "GET" "/artisans/available?page=1&page_size=20" "Get Available Artisans" "" "true"

# 6. Get Top Rated Artisans
test_endpoint "GET" "/artisans/top-rated?limit=10" "Get Top Rated Artisans" "" "true"

# 7. Get Artisans by Specialization
test_endpoint "GET" "/artisans/specialization/woodworking?page=1&page_size=20" "Get Artisans by Specialization" "" "true"

# 8. Search Artisans
test_endpoint "POST" "/artisans/search" "Search Artisans" "{
  \"tenant_id\": \"$TENANT_ID\",
  \"query\": \"wood\",
  \"page\": 1,
  \"page_size\": 20
}" "true"

# 9. Find Nearby Artisans
test_endpoint "POST" "/artisans/nearby" "Find Nearby Artisans" "{
  \"tenant_id\": \"$TENANT_ID\",
  \"latitude\": 5.6037,
  \"longitude\": -0.1870,
  \"radius_km\": 10,
  \"page\": 1,
  \"page_size\": 20
}" "true"

echo -e "${BLUE}=== AVAILABILITY MANAGEMENT ===${NC}"
echo ""

# 10. Update Availability
test_endpoint "PUT" "/artisans/$EXISTING_ARTISAN_ID/availability" "Update Artisan Availability" "{
  \"is_available\": true,
  \"availability_note\": \"Available for bookings Monday-Friday, 8AM-6PM\"
}" "true"

sleep 1

# 11. Update Availability - Set to Unavailable
test_endpoint "PUT" "/artisans/$EXISTING_ARTISAN_ID/availability" "Set Artisan Unavailable" "{
  \"is_available\": false,
  \"availability_note\": \"On vacation until next week\"
}" "true"

sleep 1

# 12. Update Availability - Back to Available
test_endpoint "PUT" "/artisans/$EXISTING_ARTISAN_ID/availability" "Set Artisan Available Again" "{
  \"is_available\": true,
  \"availability_note\": \"\"
}" "true"

echo -e "${BLUE}=== STATISTICS & ANALYTICS ===${NC}"
echo ""

# 13. Get Artisan Statistics
test_endpoint "GET" "/artisans/$EXISTING_ARTISAN_ID/stats" "Get Artisan Statistics" "" "true"

# 14. Get Dashboard Stats
test_endpoint "GET" "/artisans/$EXISTING_ARTISAN_ID/dashboard" "Get Artisan Dashboard Stats" "" "true"

echo -e "${BLUE}=== CREATE ARTISAN (Will likely fail - requires user creation) ===${NC}"
echo ""

# 15. Create Artisan - This will likely fail because we need to create a user first
# But let's test it to see the validation
test_endpoint "POST" "/artisans" "Create Artisan (Test Validation)" "{
  \"user_id\": \"b1234567-e89b-12d3-a456-426614174002\",
  \"tenant_id\": \"$TENANT_ID\",
  \"bio\": \"Master plumber with extensive experience\",
  \"specialization\": [\"plumbing\", \"pipe_fitting\", \"drainage\"],
  \"years_experience\": 8,
  \"commission_rate\": 15.0,
  \"auto_accept_bookings\": true,
  \"booking_lead_time\": 24,
  \"max_advance_booking\": 90,
  \"simultaneous_bookings\": 3,
  \"service_radius\": 25
}" "true"

echo -e "${BLUE}=== VALIDATION TESTS ===${NC}"
echo ""

# 16. Create Artisan - Missing Required Field
test_endpoint "POST" "/artisans" "Create Artisan - Missing User ID" "{
  \"tenant_id\": \"$TENANT_ID\",
  \"specialization\": [\"carpentry\"],
  \"years_experience\": 5,
  \"commission_rate\": 10.0,
  \"simultaneous_bookings\": 2
}" "true"

# 17. Create Artisan - Invalid Commission Rate
test_endpoint "POST" "/artisans" "Create Artisan - Invalid Commission Rate" "{
  \"user_id\": \"c1234567-e89b-12d3-a456-426614174003\",
  \"tenant_id\": \"$TENANT_ID\",
  \"specialization\": [\"electrical\"],
  \"years_experience\": 3,
  \"commission_rate\": 150.0,
  \"simultaneous_bookings\": 1
}" "true"

# 18. Create Artisan - Negative Years Experience
test_endpoint "POST" "/artisans" "Create Artisan - Negative Years Experience" "{
  \"user_id\": \"d1234567-e89b-12d3-a456-426614174004\",
  \"tenant_id\": \"$TENANT_ID\",
  \"specialization\": [\"welding\"],
  \"years_experience\": -5,
  \"commission_rate\": 12.0,
  \"simultaneous_bookings\": 2
}" "true"

# 19. Create Artisan - No Specialization
test_endpoint "POST" "/artisans" "Create Artisan - No Specialization" "{
  \"user_id\": \"e1234567-e89b-12d3-a456-426614174005\",
  \"tenant_id\": \"$TENANT_ID\",
  \"specialization\": [],
  \"years_experience\": 5,
  \"commission_rate\": 10.0,
  \"simultaneous_bookings\": 2
}" "true"

# 20. Update Artisan - Invalid Commission Rate
test_endpoint "PUT" "/artisans/$EXISTING_ARTISAN_ID" "Update Artisan - Invalid Commission" "{
  \"commission_rate\": 120.0
}" "true"

echo -e "${BLUE}=== FILTER & QUERY TESTS ===${NC}"
echo ""

# 21. List Artisans with Filters
test_endpoint "GET" "/artisans?page=1&page_size=10&is_available=true&min_rating=4.0" "List Artisans - Filter by Availability and Rating" "" "true"

# 22. Search with Short Query (should fail validation)
test_endpoint "POST" "/artisans/search" "Search Artisans - Short Query" "{
  \"tenant_id\": \"$TENANT_ID\",
  \"query\": \"a\",
  \"page\": 1,
  \"page_size\": 20
}" "true"

echo "==============================="
echo "Test Complete!"
echo "==============================="
echo ""
echo "Summary:"
echo "- Tested artisan profile management"
echo "- Tested artisan discovery endpoints"
echo "- Tested availability management"
echo "- Tested statistics and analytics"
echo "- Tested validation rules"
echo "- Tested search and filtering"
