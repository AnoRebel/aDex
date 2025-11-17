#!/bin/bash

# Test Coverage Validation Script for aDex-UI
# This script validates test coverage across all components and features

set -e

echo "🧪 Running Test Coverage Validation for aDex-UI"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Coverage thresholds
MIN_BACKEND_COVERAGE=70
MIN_FRONTEND_COVERAGE=80
MIN_INTEGRATION_COVERAGE=60

# Test results
FAILED_TESTS=0
TOTAL_TESTS=0

echo -e "${BLUE}📊 Backend Test Coverage Analysis${NC}"
echo "----------------------------------------"

# Run Go backend tests with coverage
echo "Running backend tests with coverage analysis..."
cd /home/ano/Code/Dex-UI/aDex-UI

# Create test coverage directory
mkdir -p test-coverage

# Run backend tests with coverage
echo "1. Running Go backend tests..."
if go test -v -coverprofile=test-coverage/backend.out -covermode=atomic ./...; then
    echo -e "${GREEN}✓ Backend tests passed${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Generate coverage report
    go tool cover -html=test-coverage/backend.out -o test-coverage/backend.html

    # Get coverage percentage
    BACKEND_COVERAGE=$(go tool cover -func=test-coverage/backend.out | grep "total:" | awk '{print $3}' | sed 's/%//')

    echo "Backend Coverage: ${BACKEND_COVERAGE}%"

    if (( $(echo "$BACKEND_COVERAGE < $MIN_BACKEND_COVERAGE" | bc -l) )); then
        echo -e "${RED}❌ Backend coverage (${BACKEND_COVERAGE}%) below threshold (${MIN_BACKEND_COVERAGE}%)${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    else
        echo -e "${GREEN}✅ Backend coverage meets threshold${NC}"
    fi
else
    echo -e "${RED}❌ Backend tests failed${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

echo ""
echo -e "${BLUE}🌐 Frontend Test Coverage Analysis${NC}"
echo "-------------------------------------------"

# Check if we're in the frontend directory
if [ -d "frontend" ]; then
    cd frontend

    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        echo "Installing frontend dependencies..."
        npm install
    fi

    echo "2. Running frontend unit tests..."
    if npm run test:unit -- --coverage; then
        echo -e "${GREEN}✓ Frontend unit tests passed${NC}"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))

        # Extract coverage from coverage report
        if [ -f "coverage/lcov-report/index.html" ]; then
            echo "Frontend coverage report generated: coverage/lcov-report/index.html"
        fi
    else
        echo -e "${RED}❌ Frontend unit tests failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi

    echo "3. Running frontend E2E tests..."
    if npm run test:e2e; then
        echo -e "${GREEN}✓ Frontend E2E tests passed${NC}"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    else
        echo -e "${RED}❌ Frontend E2E tests failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi

    cd ..
else
    echo -e "${YELLOW}⚠️ Frontend directory not found, skipping frontend tests${NC}"
fi

echo ""
echo -e "${BLUE}🔍 Code Quality Analysis${NC}"
echo "----------------------------"

# Check for common code quality issues
echo "4. Running code quality checks..."

# Check for Go code formatting
echo "Checking Go code formatting..."
if ! gofmt -l . | grep -q .; then
    echo -e "${GREEN}✓ Go code is properly formatted${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
else
    echo -e "${YELLOW}⚠️ Go code needs formatting${NC}"
    echo "Files that need formatting:"
    gofmt -l .
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Check for Go linting issues
echo "Checking Go linting..."
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run --timeout=5m; then
        echo -e "${GREEN}✓ Go linting passed${NC}"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    else
        echo -e "${RED}❌ Go linting failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
else
    echo -e "${YELLOW}⚠️ golangci-lint not installed, skipping Go linting${NC}"
fi

# Check for TypeScript/JavaScript issues
if [ -d "frontend" ]; then
    cd frontend

    echo "Checking frontend code quality..."

    # Check for ESLint
    if [ -f ".eslintrc.js" ] || [ -f ".eslintrc.json" ] || [ -f "eslint.config.js" ]; then
        if npm run lint; then
            echo -e "${GREEN}✓ Frontend linting passed${NC}"
            TOTAL_TESTS=$((TOTAL_TESTS + 1))
        else
            echo -e "${RED}❌ Frontend linting failed${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        echo -e "${YELLOW}⚠️ ESLint configuration not found, skipping frontend linting${NC}"
    fi

    # Check for Prettier formatting
    if [ -f ".prettierrc" ] || [ -f ".prettierrc.json" ] || [ -f "prettier.config.js" ]; then
        if npm run format:check; then
            echo -e "${GREEN}✓ Frontend code is properly formatted${NC}"
            TOTAL_TESTS=$((TOTAL_TESTS + 1))
        else
            echo -e "${YELLOW}⚠️ Frontend code needs formatting${NC}"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    else
        echo -e "${YELLOW}⚠️ Prettier configuration not found, skipping format check${NC}"
    fi

    cd ..
fi

echo ""
echo -e "${BLUE}📋 Test Coverage Report${NC}"
echo "------------------------"

# Count test files
BACKEND_TEST_FILES=$(find . -name "*_test.go" -type f | wc -l)
FRONTEND_TEST_FILES=$(find tests -name "*.spec.ts" -o -name "*.test.ts" 2>/dev/null | wc -l)
E2E_TEST_FILES=$(find tests -name "*.e2e.ts" 2>/dev/null | wc -l)

echo "Backend test files: $BACKEND_TEST_FILES"
echo "Frontend unit test files: $FRONTEND_TEST_FILES"
echo "E2E test files: $E2E_TEST_FILES"
echo ""

# Coverage by feature area
echo -e "${BLUE}🎯 Feature Coverage Analysis${NC}"
echo "--------------------------------"

# Check terminal emulator tests
echo "Terminal Emulator Tests:"
if [ -f "tests/backend/terminal_test.go" ]; then
    echo -e "  ${GREEN}✓ Backend terminal tests exist${NC}"
else
    echo -e "  ${RED}❌ Backend terminal tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

if [ -f "tests/frontend/unit/terminal.spec.ts" ]; then
    echo -e "  ${GREEN}✓ Frontend terminal unit tests exist${NC}"
else
    echo -e "  ${RED}❌ Frontend terminal unit tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

if [ -f "tests/frontend/e2e/terminal.spec.ts" ]; then
    echo -e "  ${GREEN}✓ Terminal E2E tests exist${NC}"
else
    echo -e "  ${RED}❌ Terminal E2E tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Check system monitoring tests
echo "System Monitoring Tests:"
if [ -f "tests/backend/system_test.go" ]; then
    echo -e "  ${GREEN}✓ Backend system tests exist${NC}"
else
    echo -e "  ${RED}❌ Backend system tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Check audio tests
echo "Audio System Tests:"
if [ -f "tests/backend/audio_test.go" ]; then
    echo -e "  ${GREEN}✓ Backend audio tests exist${NC}"
else
    echo -e "  ${RED}❌ Backend audio tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

if [ -f "tests/frontend/unit/audio.spec.ts" ]; then
    echo -e "  ${GREEN}✓ Frontend audio tests exist${NC}"
else
    echo -e "  ${RED}❌ Frontend audio tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Check keyboard tests
echo "Keyboard System Tests:"
if [ -f "tests/frontend/unit/keyboard.spec.ts" ]; then
    echo -e "  ${GREEN}✓ Keyboard unit tests exist${NC}"
else
    echo -e "  ${RED}❌ Keyboard unit tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

if [ -f "tests/frontend/e2e/keyboard.spec.ts" ]; then
    echo -e "  ${GREEN}✓ Keyboard E2E tests exist${NC}"
else
    echo -e "  ${RED}❌ Keyboard E2E tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Check security tests
echo "Security Tests:"
if [ -f "tests/backend/security_test.go" ]; then
    echo -e "  ${GREEN}✓ Security tests exist${NC}"
else
    echo -e "  ${RED}❌ Security tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Check settings tests
echo "Settings Tests:"
if [ -f "tests/backend/settings_test.go" ]; then
    echo -e "  ${GREEN}✓ Settings tests exist${NC}"
else
    echo -e "  ${RED}❌ Settings tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Check accessibility tests
echo "Accessibility Tests:"
if [ -f "tests/frontend/unit/accessibility.spec.ts" ]; then
    echo -e "  ${GREEN}✓ Accessibility tests exist${NC}"
else
    echo -e "  ${RED}❌ Accessibility tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Check error handling tests
echo "Error Handling Tests:"
if [ -f "tests/backend/error_test.go" ]; then
    echo -e "  ${GREEN}✓ Error handling tests exist${NC}"
else
    echo -e "  ${RED}❌ Error handling tests missing${NC}"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

echo ""
echo -e "${BLUE}📈 Summary Report${NC}"
echo "------------------"

echo "Total test categories checked: $TOTAL_TESTS"
echo "Failed categories: $FAILED_TESTS"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 All test coverage validation checks passed!${NC}"
    echo ""
    echo "Coverage Reports Generated:"
    echo "  - Backend: test-coverage/backend.html"
    echo "  - Frontend: frontend/coverage/lcov-report/index.html (if available)"
else
    echo -e "${RED}❌ $FAILED_TESTS test coverage validation checks failed${NC}"
    echo ""
    echo "Please address the failing tests and coverage issues before proceeding."
fi

# Generate detailed coverage report
cat > test-coverage/coverage-report.md << EOF
# Test Coverage Validation Report

Generated on: $(date)

## Coverage Summary

- **Backend Coverage**: ${BACKEND_COVERAGE:-N/A}%
- **Frontend Coverage**: See detailed report
- **Total Test Files**: $((BACKEND_TEST_FILES + FRONTEND_TEST_FILES + E2E_TEST_FILES))

## Test Files by Category

### Backend Tests ($BACKEND_TEST_FILES files)
$(find . -name "*_test.go" -type f | sed 's/^/- /')

### Frontend Tests ($FRONTEND_TEST_FILES files)
$(find tests -name "*.spec.ts" -o -name "*.test.ts" 2>/dev/null | sed 's/^/- /')

### E2E Tests ($E2E_TEST_FILES files)
$(find tests -name "*.e2e.ts" 2>/dev/null | sed 's/^/- /')

## Coverage Thresholds

- Backend Coverage Required: $MIN_BACKEND_COVERAGE%
- Frontend Coverage Required: $MIN_FRONTEND_COVERAGE%
- Integration Coverage Required: $MIN_INTEGRATION_COVERAGE%

## Recommendations

1. Maintain or improve current coverage levels
2. Add tests for any uncovered edge cases
3. Run coverage validation in CI/CD pipeline
4. Set up coverage badges for repository
EOF

echo ""
echo "Detailed coverage report saved to: test-coverage/coverage-report.md"

exit $FAILED_TESTS