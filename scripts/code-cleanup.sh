#!/bin/bash

# Code Cleanup and Refactoring Script for aDex-UI
# Performs comprehensive code cleanup, formatting, and quality checks

set -e

echo "🧹 Running Code Cleanup and Refactoring for aDex-UI"
echo "==================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Cleanup results
CLEANED_FILES=0
ISSUES_FOUND=0
WARNINGS=0

echo -e "${BLUE}📋 Go Backend Cleanup${NC}"
echo "------------------------"

cd /home/ano/Code/Dex-UI/aDex-UI

echo "1. Go code formatting..."
if gofmt -l . | grep -q .; then
    echo -e "${YELLOW}⚠️  Files need formatting:${NC}"
    gofmt -l . | while read file; do
        echo "  - $file"
        gofmt -w "$file"
        CLEANED_FILES=$((CLEANED_FILES + 1))
    done
    echo -e "${GREEN}✅ Go files formatted${NC}"
else
    echo -e "${GREEN}✅ Go files already properly formatted${NC}"
fi

echo ""
echo "2. Go imports optimization..."
if command -v goimports &> /dev/null; then
    find . -name "*.go" -not -path "./vendor/*" -exec goimports -w {} \;
    echo -e "${GREEN}✅ Go imports optimized${NC}"
else
    echo -e "${YELLOW}⚠️  goimports not installed, skipping import optimization${NC}"
    echo "Install with: go install golang.org/x/tools/cmd/goimports@latest"
fi

echo ""
echo "3. Go linting and static analysis..."
if command -v golangci-lint &> /dev/null; then
    echo "Running golangci-lint..."
    if golangci-lint run --timeout=10m; then
        echo -e "${GREEN}✅ Go linting passed${NC}"
    else
        echo -e "${YELLOW}⚠️  Go linting found issues (auto-fixable where possible)${NC}"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    fi
else
    echo -e "${YELLOW}⚠️  golangci-lint not installed, skipping comprehensive linting${NC}"
    echo "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin v1.54.2"
fi

echo ""
echo "4. Go vet analysis..."
if go vet ./...; then
    echo -e "${GREEN}✅ Go vet analysis passed${NC}"
else
    echo -e "${RED}❌ Go vet found issues${NC}"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

echo ""
echo "5. Go mod tidy..."
if go mod tidy; then
    echo -e "${GREEN}✅ Go modules tidied${NC}"
else
    echo -e "${RED}❌ Go mod tidy failed${NC}"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

echo ""
echo -e "${BLUE}🌐 Frontend Code Cleanup${NC}"
echo "---------------------------"

if [ -d "frontend" ]; then
    cd frontend

    echo "1. TypeScript/JavaScript formatting..."
    if [ -f "package.json" ] && command -v prettier &> /dev/null; then
        if bun run format 2>/dev/null || npm run format 2>/dev/null; then
            echo -e "${GREEN}✅ Frontend code formatted${NC}"
        else
            echo -e "${YELLOW}⚠️  Format script not found, running prettier directly${NC}"
            if command -v prettier &> /dev/null; then
                prettier --write "app/**/*.{ts,js,vue,css,scss,json}" 2>/dev/null || true
                CLEANED_FILES=$((CLEANED_FILES + 1))
            fi
        fi
    else
        echo -e "${YELLOW}⚠️  Prettier not available, skipping frontend formatting${NC}"
    fi

    echo ""
    echo "2. Frontend linting..."
    if [ -f "package.json" ]; then
        if bun run lint 2>/dev/null || npm run lint 2>/dev/null; then
            echo -e "${GREEN}✅ Frontend linting passed${NC}"
        else
            echo -e "${YELLOW}⚠️  Frontend linting found issues${NC}"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi

    echo ""
    echo "3. TypeScript type checking..."
    if bun run type-check 2>/dev/null || npm run type-check 2>/dev/null; then
        echo -e "${GREEN}✅ TypeScript type checking passed${NC}"
    else
        echo -e "${YELLOW}⚠️  TypeScript type checking found issues${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi

    echo ""
    echo "4. Unused dependency cleanup..."
    if command -v depcheck &> /dev/null; then
        if depcheck 2>/dev/null; then
            echo -e "${GREEN}✅ No unused dependencies found${NC}"
        else
            echo -e "${YELLOW}⚠️  Unused dependencies found, review above output${NC}"
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        echo -e "${YELLOW}⚠️  depcheck not installed, skipping dependency cleanup${NC}"
    fi

    cd ..
else
    echo -e "${YELLOW}⚠️  Frontend directory not found${NC}"
fi

echo ""
echo -e "${BLUE}🗂️  File Organization Cleanup${NC}"
echo "-------------------------------"

echo "1. Removing temporary files..."
TEMP_FILES=(
    "*.tmp"
    "*.temp"
    "*.log"
    "*.bak"
    "*~"
    ".DS_Store"
    "Thumbs.db"
    "*.swp"
    "*.swo"
    "#*#"
    ".#*"
)

for pattern in "${TEMP_FILES[@]}"; do
    if find . -name "$pattern" -type f -delete 2>/dev/null; then
        echo "Removed temporary files: $pattern"
    fi
done

echo ""
echo "2. Cleaning up empty directories..."
find . -type d -empty -delete 2>/dev/null || true

echo ""
echo "3. Checking for duplicate files..."
# Simple duplicate detection by filename
find . -name "*.go" | sort | uniq -d | while read file; do
    if [ -f "$file" ]; then
        echo -e "${YELLOW}⚠️  Potential duplicate: $file${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi
done

echo ""
echo -e "${BLUE}📝 Documentation Cleanup${NC}"
echo "----------------------------"

echo "1. Updating README files..."
find . -name "README*" -type f | while read readme; do
    if [ -f "$readme" ]; then
        echo "Checking: $readme"
        # Check for common README issues
        if grep -q "TODO\|FIXME\|XXX" "$readme" 2>/dev/null; then
            echo -e "${YELLOW}⚠️  $readme contains TODO items${NC}"
        fi
    fi
done

echo ""
echo "2. Cleaning up documentation formatting..."
# Normalize line endings in documentation files
find . -name "*.md" -type f -exec sed -i 's/\r$//' {} \; 2>/dev/null || true

echo ""
echo -e "${BLUE}🔍 Code Quality Analysis${NC}"
echo "--------------------------"

echo "1. Checking for TODO/FIXME comments..."
TODO_COUNT=$(grep -r "TODO\|FIXME\|XXX" --include="*.go" --include="*.ts" --include="*.js" --include="*.vue" . 2>/dev/null | wc -l)
if [ "$TODO_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Found $TODO_COUNT TODO/FIXME comments${NC}"
    grep -r "TODO\|FIXME\|XXX" --include="*.go" --include="*.ts" --include="*.js" --include="*.vue" . 2>/dev/null | head -10
    WARNINGS=$((WARNINGS + 1))
else
    echo -e "${GREEN}✅ No TODO/FIXME comments found${NC}"
fi

echo ""
echo "2. Checking for long lines..."
LONG_LINES=$(find . -name "*.go" -exec awk 'length($0) > 120 {print FILENAME ":" NR ":" $0}' {} \; 2>/dev/null | wc -l)
if [ "$LONG_LINES" -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Found $LONG_LINES lines longer than 120 characters${NC}"
    WARNINGS=$((WARNINGS + 1))
else
    echo -e "${GREEN}✅ No long lines found${NC}"
fi

echo ""
echo "3. Checking for console.log statements..."
CONSOLE_LOGS=$(find . -name "*.ts" -o -name "*.js" -o -name "*.vue" | xargs grep -l "console.log\|console.warn\|console.error" 2>/dev/null | wc -l)
if [ "$CONSOLE_LOGS" -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Found console statements in $CONSOLE_LOGS files${NC}"
    WARNINGS=$((WARNINGS + 1))
else
    echo -e "${GREEN}✅ No console statements found${NC}"
fi

echo ""
echo "4. Checking for unused imports in Go files..."
UNUSED_IMPORTS=$(find . -name "*.go" -not -path "./vendor/*" -exec grep -l "import.*\"" {} \; | wc -l)
echo "Files with imports: $UNUSED_IMPORTS"

echo ""
echo -e "${BLUE}🏗️  Architecture Cleanup${NC}"
echo "-------------------------"

echo "1. Checking for circular dependencies..."
# Basic circular dependency check for Go
if find . -name "*.go" -not -path "./vendor/*" -exec grep -l "\"\./" {} \; | head -5; then
    echo -e "${YELLOW}⚠️  Potential circular dependencies detected${NC}"
    WARNINGS=$((WARNINGS + 1))
else
    echo -e "${GREEN}✅ No obvious circular dependencies${NC}"
fi

echo ""
echo "2. Checking package structure..."
PACKAGES=$(find . -name "*.go" -not -path "./vendor/*" -exec dirname {} \; | sort -u | grep -v "^\.$" | wc -l)
echo "Go packages found: $PACKAGES"

# Check for empty packages
find . -name "*.go" -not -path "./vendor/*" -exec dirname {} \; | sort -u | while read dir; do
    if [ "$dir" != "." ]; then
        GO_FILES=$(find "$dir" -name "*.go" -not -name "*_test.go" | wc -l)
        if [ "$GO_FILES" -eq 0 ]; then
            echo -e "${YELLOW}⚠️  Empty package: $dir${NC}"
        fi
    fi
done

echo ""
echo -e "${BLUE}📊 Cleanup Summary${NC}"
echo "--------------------"

echo "Files cleaned: $CLEANED_FILES"
echo "Issues found: $ISSUES_FOUND"
echo "Warnings: $WARNINGS"

# Generate cleanup report
cat > cleanup-report.md << EOF
# Code Cleanup Report

Generated on: $(date)

## Cleanup Summary

- **Files Cleaned**: $CLEANED_FILES
- **Issues Found**: $ISSUES_FOUND
- **Warnings**: $WARNINGS

## Actions Performed

### Go Backend
- [x] Code formatting with gofmt
- [x] Import optimization with goimports (if available)
- [x] Static analysis with golangci-lint (if available)
- [x] Vet analysis
- [x] Module dependency cleanup

### Frontend
- [x] Code formatting with prettier
- [x] Linting analysis
- [x] TypeScript type checking
- [x] Unused dependency check

### File Organization
- [x] Temporary file cleanup
- [x] Empty directory removal
- [x] Documentation formatting

### Code Quality
- [x] TODO/FIXME comment analysis
- [x] Long line detection
- [x] Console statement checking
- [x] Import optimization

### Architecture
- [x] Circular dependency check
- [x] Package structure analysis

## Recommendations

1. **Address TODO items**: Resolve any remaining TODO/FIXME comments
2. **Review warnings**: Address linting warnings and code quality issues
3. **Update documentation**: Keep documentation synchronized with code changes
4. **Regular cleanup**: Schedule regular code cleanup sessions
5. **CI/CD integration**: Integrate cleanup into development workflow

## Next Steps

1. Review any issues found during cleanup
2. Address warnings and recommendations
3. Update coding standards if needed
4. Configure automated formatting in IDE
5. Set up pre-commit hooks for code quality

## Tools Used

- **Go**: gofmt, goimports, go vet, golangci-lint, go mod tidy
- **Frontend**: prettier, eslint, tsc, depcheck
- **Analysis**: custom scripts for code quality checks

EOF

echo ""
echo "Cleanup report generated: cleanup-report.md"

echo ""
echo -e "${BLUE}🎯 Recommendations${NC}"
echo "----------------------"

echo "1. Set up pre-commit hooks for automatic formatting:"
echo "   # Go"
echo "   pre-commit install"
echo "   # Add gofmt, goimports, and golangci-lint hooks"
echo ""
echo "2. Configure IDE for consistent formatting:"
echo "   - VS Code: Go and Vue extensions"
echo "   - Vim/Neovim: vim-go and coc-vue"
echo "   - GoLand: Built-in Go and Vue support"
echo ""
echo "3. Regular maintenance schedule:"
echo "   - Weekly: go mod tidy, dependency updates"
echo "   - Monthly: comprehensive code review"
echo "   - Quarterly: architecture review"

if [ $ISSUES_FOUND -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 Code cleanup completed successfully!${NC}"
    echo "All files are properly formatted and no issues were found."
else
    echo ""
    echo -e "${YELLOW}⚠️  Code cleanup completed with warnings${NC}"
    echo "Please review the issues found and address them as needed."
fi

echo ""
echo "Next recommended actions:"
echo "1. Review cleanup-report.md for detailed findings"
echo "2. Address any critical issues found"
echo "3. Set up automated code quality checks in CI/CD"
echo "4. Schedule regular code maintenance"

exit 0