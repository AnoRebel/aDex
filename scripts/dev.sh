#!/bin/bash
# Development script for aDex-UI
# Starts Nuxt dev server first, waits for it, then starts Wails

set -e

echo "🚀 Starting aDex-UI Development Environment..."

# Start Nuxt dev server in background
echo "📦 Starting Nuxt dev server..."
cd frontend
bun run dev &
NUXT_PID=$!
cd ..

# Wait for Nuxt to be ready
echo "⏳ Waiting for Nuxt dev server to be ready..."
MAX_ATTEMPTS=30
ATTEMPT=0
while ! curl -s http://127.0.0.1:9245 > /dev/null 2>&1; do
    ATTEMPT=$((ATTEMPT + 1))
    if [ $ATTEMPT -ge $MAX_ATTEMPTS ]; then
        echo "❌ Nuxt dev server failed to start after ${MAX_ATTEMPTS} seconds"
        kill $NUXT_PID 2>/dev/null || true
        exit 1
    fi
    echo "  Attempt $ATTEMPT/$MAX_ATTEMPTS..."
    sleep 1
done

echo "✅ Nuxt dev server is ready!"
echo "🔧 Starting Wails..."

# Start Wails dev (it will use the already running Nuxt server)
wails dev

# Cleanup on exit
trap "kill $NUXT_PID 2>/dev/null || true" EXIT
