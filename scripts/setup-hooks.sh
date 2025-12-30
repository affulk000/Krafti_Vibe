#!/bin/bash

# Git Hooks Installation Script for Krafti_Vibe
# This script installs git hooks to maintain code quality

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Detect script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$PROJECT_ROOT/.git/hooks"
SOURCE_HOOKS_DIR="$PROJECT_ROOT/scripts/git-hooks"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Krafti Vibe Git Hooks Setup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if we're in a git repository
if [ ! -d "$PROJECT_ROOT/.git" ]; then
    echo -e "${RED}Error: Not a git repository${NC}"
    echo -e "${YELLOW}Please run this script from within the Krafti_Vibe project directory${NC}"
    echo ""
    exit 1
fi

# Check if source hooks directory exists
if [ ! -d "$SOURCE_HOOKS_DIR" ]; then
    echo -e "${RED}Error: Hooks directory not found${NC}"
    echo -e "${YELLOW}Expected: $SOURCE_HOOKS_DIR${NC}"
    echo ""
    exit 1
fi

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

echo -e "${GREEN}Installing git hooks...${NC}"
echo ""

# Install hooks
HOOKS_INSTALLED=0
HOOKS_FAILED=0

for hook in pre-commit commit-msg pre-push; do
    SOURCE="$SOURCE_HOOKS_DIR/$hook"
    TARGET="$HOOKS_DIR/$hook"

    if [ -f "$SOURCE" ]; then
        # Backup existing hook if it exists and is not our hook
        if [ -f "$TARGET" ] && ! grep -q "Krafti_Vibe" "$TARGET" 2>/dev/null; then
            BACKUP="$TARGET.backup.$(date +%s)"
            echo -e "${YELLOW}  Backing up existing $hook hook to:${NC}"
            echo -e "${YELLOW}    $(basename $BACKUP)${NC}"
            mv "$TARGET" "$BACKUP"
        fi

        # Install the hook
        echo -e "${GREEN}  Installing ${BLUE}$hook${GREEN} hook...${NC}"
        cp "$SOURCE" "$TARGET"
        chmod +x "$TARGET"
        HOOKS_INSTALLED=$((HOOKS_INSTALLED + 1))
        echo -e "${GREEN}    ✓ $hook installed${NC}"
    else
        echo -e "${RED}  ✗ $hook source not found${NC}"
        HOOKS_FAILED=$((HOOKS_FAILED + 1))
    fi
done

echo ""
echo -e "${BLUE}========================================${NC}"

if [ $HOOKS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ Success! All git hooks installed${NC}"
else
    echo -e "${YELLOW}⚠️  Partial installation: $HOOKS_INSTALLED installed, $HOOKS_FAILED failed${NC}"
fi

echo -e "${BLUE}========================================${NC}"
echo ""

if [ $HOOKS_INSTALLED -gt 0 ]; then
    echo -e "${YELLOW}Installed hooks:${NC}"
    echo ""
    echo -e "  ${GREEN}pre-commit${NC}"
    echo -e "    - Runs formatting checks (gofmt, goimports)"
    echo -e "    - Runs linting (golangci-lint)"
    echo -e "    - Runs go vet"
    echo -e "    - Detects secrets (passwords, API keys, tokens)"
    echo -e "    - Checks only staged files (fast)"
    echo ""
    echo -e "  ${GREEN}commit-msg${NC}"
    echo -e "    - Enforces conventional commit message format"
    echo -e "    - Format: type(scope): subject"
    echo -e "    - Valid types: feat, fix, docs, style, refactor, test, chore, perf, ci, build"
    echo -e "    - Example: feat(auth): add JWT token refresh"
    echo ""
    echo -e "  ${GREEN}pre-push${NC}"
    echo -e "    - Runs full formatting check"
    echo -e "    - Runs go vet on entire codebase"
    echo -e "    - Runs full linter (5m timeout)"
    echo -e "    - Runs full test suite with race detector"
    echo -e "    - Warns when pushing to main/master"
    echo ""
    echo -e "${BLUE}────────────────────────────────────────${NC}"
    echo ""
    echo -e "${YELLOW}Usage Tips:${NC}"
    echo ""
    echo -e "  ${BLUE}Make a commit:${NC}"
    echo -e "    git add ."
    echo -e "    git commit -m \"feat(api): add new endpoint\""
    echo -e "    ${GREEN}# Hooks run automatically!${NC}"
    echo ""
    echo -e "  ${BLUE}Bypass hooks (not recommended):${NC}"
    echo -e "    git commit --no-verify"
    echo -e "    git push --no-verify"
    echo ""
    echo -e "  ${BLUE}Recommended workflow:${NC}"
    echo -e "    1. Create feature branch: ${GREEN}git checkout -b feat/my-feature${NC}"
    echo -e "    2. Make changes and commit: ${GREEN}git commit -m \"feat: add feature\"${NC}"
    echo -e "    3. Push to branch: ${GREEN}git push origin feat/my-feature${NC}"
    echo -e "    4. Create Pull Request for review"
    echo ""
    echo -e "${BLUE}────────────────────────────────────────${NC}"
    echo ""
    echo -e "${GREEN}Prerequisites:${NC}"
    echo -e "  - Go 1.24+"
    echo -e "  - gofmt (included with Go)"
    echo -e "  - goimports: ${BLUE}go install golang.org/x/tools/cmd/goimports@latest${NC}"
    echo -e "  - golangci-lint: ${BLUE}make install-tools${NC}"
    echo ""
    echo -e "${YELLOW}If hooks fail due to missing tools, install them with:${NC}"
    echo -e "  ${BLUE}make install-tools${NC}"
    echo ""
fi

echo -e "${GREEN}Setup complete!${NC}"
echo ""

exit 0
