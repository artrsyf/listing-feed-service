#!/bin/bash
set -e

echo "📋 Initializing Elasticsearch index..."

ELASTIC_URL="${ELASTIC_URL:-http://localhost:9200}"

# Check if Elasticsearch is available
echo "Checking Elasticsearch connection..."
if ! curl -s "$ELASTIC_URL" > /dev/null 2>&1; then
    echo "❌ Elasticsearch is not available at $ELASTIC_URL"
    echo "   Make sure Elasticsearch is running: make docker-up"
    exit 1
fi

echo "✅ Elasticsearch is available"

# Delete existing index if exists
echo "Removing existing index (if any)..."
curl -s -X DELETE "$ELASTIC_URL/orders" > /dev/null 2>&1 || true

# Create index with settings
echo "Creating orders index with optimized settings..."
response=$(curl -s -X PUT "$ELASTIC_URL/orders" \
  -H "Content-Type: application/json" \
  -d @elasticsearch/index-settings.json)

if echo "$response" | grep -q '"acknowledged":true'; then
    echo "✅ Elasticsearch index created successfully"
else
    echo "⚠️  Index creation response:"
    echo "$response" | head -20
fi

# Show index info
echo ""
echo "📊 Index info:"
curl -s "$ELASTIC_URL/orders/_settings?pretty" | head -30
